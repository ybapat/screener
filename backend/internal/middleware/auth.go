package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ybapat/screener/backend/internal/models"
	"github.com/ybapat/screener/backend/pkg/response"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "user_role"
)

type TokenParser interface {
	ParseToken(tokenStr string) (uuid.UUID, models.UserRole, error)
}

func Auth(parser TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				response.ErrorMsg(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.ErrorMsg(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			userID, role, err := parser.ParseToken(parts[1])
			if err != nil {
				response.ErrorMsg(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, RoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleKey).(models.UserRole)
			if !ok {
				response.ErrorMsg(w, http.StatusUnauthorized, "no role in context")
				return
			}

			for _, allowed := range roles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.ErrorMsg(w, http.StatusForbidden, "insufficient permissions")
		})
	}
}

func GetUserID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(UserIDKey).(uuid.UUID)
	return id
}

func GetRole(ctx context.Context) models.UserRole {
	role, _ := ctx.Value(RoleKey).(models.UserRole)
	return role
}
