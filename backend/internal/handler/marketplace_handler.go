package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ybapat/screener/backend/internal/middleware"
	"github.com/ybapat/screener/backend/internal/models"
	"github.com/ybapat/screener/backend/internal/service"
	"github.com/ybapat/screener/backend/pkg/apierror"
	"github.com/ybapat/screener/backend/pkg/response"
	"github.com/ybapat/screener/backend/pkg/validator"
)

type MarketplaceHandler struct {
	marketplace *service.MarketplaceService
	credits     *service.CreditService
}

func NewMarketplaceHandler(mp *service.MarketplaceService, credits *service.CreditService) *MarketplaceHandler {
	return &MarketplaceHandler{marketplace: mp, credits: credits}
}

func (h *MarketplaceHandler) Purchase(w http.ResponseWriter, r *http.Request) {
	buyerID := middleware.GetUserID(r.Context())
	datasetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apierror.BadRequest("invalid dataset id"))
		return
	}

	purchase, err := h.marketplace.Purchase(r.Context(), buyerID, datasetID)
	if err != nil {
		if apiErr, ok := err.(*apierror.APIError); ok {
			response.Error(w, apiErr)
			return
		}
		response.Error(w, apierror.Internal("purchase failed"))
		return
	}

	response.JSON(w, http.StatusCreated, purchase)
}

func (h *MarketplaceHandler) GetPurchases(w http.ResponseWriter, r *http.Request) {
	buyerID := middleware.GetUserID(r.Context())
	limit, offset := parsePagination(r)

	purchases, total, err := h.marketplace.GetPurchases(r.Context(), buyerID, limit, offset)
	if err != nil {
		response.Error(w, apierror.Internal("failed to list purchases"))
		return
	}

	response.JSONWithMeta(w, http.StatusOK, purchases, map[string]int{"total": total})
}

func (h *MarketplaceHandler) CreateSegment(w http.ResponseWriter, r *http.Request) {
	buyerID := middleware.GetUserID(r.Context())

	var seg models.DataSegment
	if err := validator.DecodeAndValidate(r, &seg); err != nil {
		response.Error(w, apierror.BadRequest(err.Error()))
		return
	}

	if err := h.marketplace.CreateSegment(r.Context(), buyerID, &seg); err != nil {
		if apiErr, ok := err.(*apierror.APIError); ok {
			response.Error(w, apiErr)
			return
		}
		response.Error(w, apierror.Internal("failed to create segment"))
		return
	}

	response.JSON(w, http.StatusCreated, seg)
}

func (h *MarketplaceHandler) ListSegments(w http.ResponseWriter, r *http.Request) {
	buyerID := middleware.GetUserID(r.Context())
	segments, err := h.marketplace.ListSegments(r.Context(), buyerID)
	if err != nil {
		response.Error(w, apierror.Internal("failed to list segments"))
		return
	}
	response.JSON(w, http.StatusOK, segments)
}

func (h *MarketplaceHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	buyerID := middleware.GetUserID(r.Context())
	segmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apierror.BadRequest("invalid segment id"))
		return
	}

	var req struct {
		BidCredits      int64 `json:"bid_credits" validate:"required,gt=0"`
		DurationMinutes int   `json:"duration_minutes" validate:"required,gt=0,max=10080"`
	}
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, apierror.BadRequest(err.Error()))
		return
	}

	bid, err := h.marketplace.PlaceBid(r.Context(), buyerID, segmentID, req.BidCredits, time.Duration(req.DurationMinutes)*time.Minute)
	if err != nil {
		if apiErr, ok := err.(*apierror.APIError); ok {
			response.Error(w, apiErr)
			return
		}
		response.Error(w, apierror.Internal("failed to place bid"))
		return
	}

	response.JSON(w, http.StatusCreated, bid)
}

func (h *MarketplaceHandler) ListBids(w http.ResponseWriter, r *http.Request) {
	buyerID := middleware.GetUserID(r.Context())
	bids, err := h.marketplace.ListBids(r.Context(), buyerID)
	if err != nil {
		response.Error(w, apierror.Internal("failed to list bids"))
		return
	}
	response.JSON(w, http.StatusOK, bids)
}

func (h *MarketplaceHandler) CancelBid(w http.ResponseWriter, r *http.Request) {
	buyerID := middleware.GetUserID(r.Context())
	bidID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apierror.BadRequest("invalid bid id"))
		return
	}

	if err := h.marketplace.CancelBid(r.Context(), buyerID, bidID); err != nil {
		if apiErr, ok := err.(*apierror.APIError); ok {
			response.Error(w, apiErr)
			return
		}
		response.Error(w, apierror.Internal("failed to cancel bid"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "bid cancelled"})
}

func (h *MarketplaceHandler) TopupCredits(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Amount int64 `json:"amount" validate:"required,gt=0,max=100000"`
	}
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, apierror.BadRequest(err.Error()))
		return
	}

	balance, err := h.credits.TopUp(r.Context(), userID, req.Amount)
	if err != nil {
		if apiErr, ok := err.(*apierror.APIError); ok {
			response.Error(w, apiErr)
			return
		}
		response.Error(w, apierror.Internal("topup failed"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]int64{"credit_balance": balance})
}
