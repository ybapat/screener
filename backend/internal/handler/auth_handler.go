package handler

import (
	"net/http"

	"github.com/ybapat/screener/backend/internal/middleware"
	"github.com/ybapat/screener/backend/internal/service"
	"github.com/ybapat/screener/backend/pkg/apierror"
	"github.com/ybapat/screener/backend/pkg/response"
	"github.com/ybapat/screener/backend/pkg/validator"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, apierror.BadRequest(err.Error()))
		return
	}

	user, tokens, err := h.auth.Register(r.Context(), req)
	if err != nil {
		if apiErr, ok := err.(*apierror.APIError); ok {
			response.Error(w, apiErr)
			return
		}
		response.Error(w, apierror.Internal("registration failed"))
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, apierror.BadRequest(err.Error()))
		return
	}

	user, tokens, err := h.auth.Login(r.Context(), req)
	if err != nil {
		if apiErr, ok := err.(*apierror.APIError); ok {
			response.Error(w, apiErr)
			return
		}
		response.Error(w, apierror.Internal("login failed"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, apierror.BadRequest(err.Error()))
		return
	}

	tokens, err := h.auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if apiErr, ok := err.(*apierror.APIError); ok {
			response.Error(w, apiErr)
			return
		}
		response.Error(w, apierror.Internal("refresh failed"))
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if err := h.auth.Logout(r.Context(), userID); err != nil {
		response.Error(w, apierror.Internal("logout failed"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}
