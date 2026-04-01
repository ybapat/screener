package handler

import (
	"net/http"

	"github.com/ybapat/screener/backend/internal/middleware"
	"github.com/ybapat/screener/backend/internal/privacy"
	"github.com/ybapat/screener/backend/internal/repository"
	"github.com/ybapat/screener/backend/pkg/apierror"
	"github.com/ybapat/screener/backend/pkg/response"
	"github.com/ybapat/screener/backend/pkg/validator"
)

type UserHandler struct {
	users         repository.UserRepository
	budgetTracker *privacy.BudgetTracker
}

func NewUserHandler(users repository.UserRepository, bt *privacy.BudgetTracker) *UserHandler {
	return &UserHandler{users: users, budgetTracker: bt}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		response.Error(w, apierror.NotFound("user not found"))
		return
	}
	response.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		DisplayName *string `json:"display_name,omitempty"`
		AgeRange    *string `json:"age_range,omitempty"`
		Country     *string `json:"country,omitempty"`
		Timezone    *string `json:"timezone,omitempty"`
	}
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, apierror.BadRequest(err.Error()))
		return
	}

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		response.Error(w, apierror.NotFound("user not found"))
		return
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.AgeRange != nil {
		user.AgeRange = req.AgeRange
	}
	if req.Country != nil {
		user.Country = req.Country
	}
	if req.Timezone != nil {
		user.Timezone = req.Timezone
	}

	if err := h.users.Update(r.Context(), user); err != nil {
		response.Error(w, apierror.Internal("failed to update profile"))
		return
	}

	response.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetPrivacyBudget(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	budget, spent, err := h.budgetTracker.GetRemainingBudget(r.Context(), userID)
	if err != nil {
		response.Error(w, apierror.Internal("failed to get budget"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"total_budget":     budget,
		"epsilon_spent":    spent,
		"epsilon_remaining": budget - spent,
		"percent_used":     (spent / budget) * 100,
	})
}

func (h *UserHandler) GetEpsilonLedger(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	limit, offset := parsePagination(r)

	entries, total, err := h.budgetTracker.GetLedger(r.Context(), userID, limit, offset)
	if err != nil {
		response.Error(w, apierror.Internal("failed to get ledger"))
		return
	}

	response.JSONWithMeta(w, http.StatusOK, entries, map[string]int{"total": total})
}
