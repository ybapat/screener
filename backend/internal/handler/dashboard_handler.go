package handler

import (
	"net/http"

	"github.com/ybapat/screener/backend/internal/middleware"
	"github.com/ybapat/screener/backend/internal/privacy"
	"github.com/ybapat/screener/backend/internal/repository"
	"github.com/ybapat/screener/backend/internal/service"
	"github.com/ybapat/screener/backend/pkg/apierror"
	"github.com/ybapat/screener/backend/pkg/response"
)

type DashboardHandler struct {
	users         repository.UserRepository
	credits       *service.CreditService
	ingestion     *service.IngestionService
	marketplace   *service.MarketplaceService
	budgetTracker *privacy.BudgetTracker
}

func NewDashboardHandler(
	users repository.UserRepository,
	credits *service.CreditService,
	ingestion *service.IngestionService,
	marketplace *service.MarketplaceService,
	bt *privacy.BudgetTracker,
) *DashboardHandler {
	return &DashboardHandler{
		users:         users,
		credits:       credits,
		ingestion:     ingestion,
		marketplace:   marketplace,
		budgetTracker: bt,
	}
}

// SellerOverview returns a combined dashboard for sellers.
func (h *DashboardHandler) SellerOverview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		response.Error(w, apierror.Internal("user not found"))
		return
	}

	batches, totalBatches, _ := h.ingestion.GetBatches(r.Context(), userID, 5, 0)
	budget, spent, _ := h.budgetTracker.GetRemainingBudget(r.Context(), userID)
	txns, _, _ := h.credits.GetHistory(r.Context(), userID, 10, 0)

	response.JSON(w, http.StatusOK, map[string]any{
		"user":            user,
		"credit_balance":  user.CreditBalance,
		"epsilon_budget":  budget,
		"epsilon_spent":   spent,
		"epsilon_remaining": budget - spent,
		"recent_batches":  batches,
		"total_batches":   totalBatches,
		"recent_transactions": txns,
	})
}

// BuyerOverview returns a combined dashboard for buyers.
func (h *DashboardHandler) BuyerOverview(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		response.Error(w, apierror.Internal("user not found"))
		return
	}

	purchases, totalPurchases, _ := h.marketplace.GetPurchases(r.Context(), userID, 5, 0)
	bids, _ := h.marketplace.ListBids(r.Context(), userID)
	txns, _, _ := h.credits.GetHistory(r.Context(), userID, 10, 0)

	activeBids := 0
	for _, b := range bids {
		if b.Status == "active" {
			activeBids++
		}
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"user":                user,
		"credit_balance":      user.CreditBalance,
		"recent_purchases":    purchases,
		"total_purchases":     totalPurchases,
		"active_bids":         activeBids,
		"recent_transactions": txns,
	})
}

func (h *DashboardHandler) CreditHistory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	limit, offset := parsePagination(r)

	txns, total, err := h.credits.GetHistory(r.Context(), userID, limit, offset)
	if err != nil {
		response.Error(w, apierror.Internal("failed to get credit history"))
		return
	}

	response.JSONWithMeta(w, http.StatusOK, txns, map[string]int{"total": total})
}
