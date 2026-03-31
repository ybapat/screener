package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ybapat/screener/backend/internal/models"
	"github.com/ybapat/screener/backend/internal/repository"
	"github.com/ybapat/screener/backend/pkg/apierror"
)

type AuthService struct {
	users         repository.UserRepository
	refreshTokens repository.RefreshTokenRepository
	jwtSecret     []byte
}

func NewAuthService(users repository.UserRepository, rt repository.RefreshTokenRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:         users,
		refreshTokens: rt,
		jwtSecret:     []byte(jwtSecret),
	}
}

type RegisterRequest struct {
	Email       string          `json:"email" validate:"required,email"`
	Password    string          `json:"password" validate:"required,min=8,max=128"`
	DisplayName string          `json:"display_name" validate:"required,min=1,max=100"`
	Role        models.UserRole `json:"role" validate:"required,oneof=seller buyer"`
	AgeRange    *string         `json:"age_range,omitempty"`
	Country     *string         `json:"country,omitempty"`
	Timezone    *string         `json:"timezone,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*models.User, *TokenPair, error) {
	existing, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, apierror.Internal("failed to check email")
	}
	if existing != nil {
		return nil, nil, apierror.Conflict("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, apierror.Internal("failed to hash password")
	}

	user := &models.User{
		ID:                  uuid.New(),
		Email:               req.Email,
		PasswordHash:        string(hash),
		DisplayName:         req.DisplayName,
		Role:                req.Role,
		AgeRange:            req.AgeRange,
		Country:             req.Country,
		Timezone:            req.Timezone,
		CreditBalance:       1000, // starter credits
		GlobalEpsilonBudget: 10.0,
		EpsilonSpent:        0.0,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, nil, apierror.Internal("failed to create user")
	}

	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*models.User, *TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, apierror.Internal("failed to look up user")
	}
	if user == nil {
		return nil, nil, apierror.Unauthorized("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, nil, apierror.Unauthorized("invalid email or password")
	}

	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	hash := hashToken(refreshToken)

	stored, err := s.refreshTokens.GetByHash(ctx, hash)
	if err != nil {
		return nil, apierror.Internal("failed to look up token")
	}
	if stored == nil || stored.Revoked || time.Now().After(stored.ExpiresAt) {
		return nil, apierror.Unauthorized("invalid or expired refresh token")
	}

	// Revoke old token (rotation)
	if err := s.refreshTokens.Revoke(ctx, stored.ID); err != nil {
		return nil, apierror.Internal("failed to revoke old token")
	}

	user, err := s.users.GetByID(ctx, stored.UserID)
	if err != nil || user == nil {
		return nil, apierror.Unauthorized("user not found")
	}

	return s.generateTokenPair(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.refreshTokens.RevokeAllForUser(ctx, userID)
}

func (s *AuthService) generateTokenPair(ctx context.Context, user *models.User) (*TokenPair, error) {
	expiresAt := time.Now().Add(15 * time.Minute)

	claims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": string(user.Role),
		"exp":  expiresAt.Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, apierror.Internal("failed to sign access token")
	}

	refreshTokenRaw := uuid.New().String()
	refreshHash := hashToken(refreshTokenRaw)

	rt := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.refreshTokens.Create(ctx, rt); err != nil {
		return nil, apierror.Internal("failed to store refresh token")
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenRaw,
		ExpiresAt:    expiresAt.Unix(),
	}, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *AuthService) ParseToken(tokenStr string) (uuid.UUID, models.UserRole, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return uuid.Nil, "", apierror.Unauthorized("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, "", apierror.Unauthorized("invalid token claims")
	}

	sub, _ := claims.GetSubject()
	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, "", apierror.Unauthorized("invalid user id in token")
	}

	role := models.UserRole(claims["role"].(string))
	return userID, role, nil
}
