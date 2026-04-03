package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ybapat/screener/backend/internal/service"
	"github.com/ybapat/screener/backend/pkg/apierror"
	"github.com/ybapat/screener/backend/pkg/response"
)

// solErr converts a generic error to an *apierror.APIError for response.Error.
func solErr(err error) *apierror.APIError {
	var apiErr *apierror.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return apierror.Internal(err.Error())
}

type SolanaHandler struct {
	solana *service.SolanaService
}

func NewSolanaHandler(s *service.SolanaService) *SolanaHandler {
	return &SolanaHandler{solana: s}
}

// GetServerInfo returns the server wallet, balance, exchange rate, and program ID.
func (h *SolanaHandler) GetServerInfo(w http.ResponseWriter, r *http.Request) {
	if h.solana == nil {
		response.Error(w, apierror.Internal("Solana integration not configured"))
		return
	}
	info, err := h.solana.GetServerInfo(r.Context())
	if err != nil {
		response.Error(w, solErr(err))
		return
	}
	response.JSON(w, http.StatusOK, info)
}

// LinkWallet associates a Solana wallet with the authenticated user.
func (h *SolanaHandler) LinkWallet(w http.ResponseWriter, r *http.Request) {
	if h.solana == nil {
		response.Error(w, apierror.Internal("Solana integration not configured"))
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	var req struct {
		Wallet    string `json:"wallet"`
		Signature string `json:"signature"`
		Message   string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequest("invalid request body"))
		return
	}
	if req.Wallet == "" || req.Signature == "" || req.Message == "" {
		response.Error(w, apierror.BadRequest("wallet, signature, and message are required"))
		return
	}

	if err := h.solana.LinkWallet(r.Context(), userID, req.Wallet, req.Signature, req.Message); err != nil {
		response.Error(w, solErr(err))
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "linked"})
}

// InitTopup returns the server wallet for the user to send SOL to.
func (h *SolanaHandler) InitTopup(w http.ResponseWriter, r *http.Request) {
	if h.solana == nil {
		response.Error(w, apierror.Internal("Solana integration not configured"))
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	var req struct {
		AmountLamports int64 `json:"amount_lamports"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequest("invalid request body"))
		return
	}

	resp, err := h.solana.InitTopup(r.Context(), userID, req.AmountLamports)
	if err != nil {
		response.Error(w, solErr(err))
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

// ConfirmTopup verifies the on-chain transfer and credits the user.
func (h *SolanaHandler) ConfirmTopup(w http.ResponseWriter, r *http.Request) {
	if h.solana == nil {
		response.Error(w, apierror.Internal("Solana integration not configured"))
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	var req struct {
		TxSignature string `json:"tx_signature"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequest("invalid request body"))
		return
	}
	if req.TxSignature == "" {
		response.Error(w, apierror.BadRequest("tx_signature is required"))
		return
	}

	resp, err := h.solana.ConfirmTopup(r.Context(), userID, req.TxSignature)
	if err != nil {
		response.Error(w, solErr(err))
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

// InitPurchase returns the escrow PDA and amount for a SOL purchase.
func (h *SolanaHandler) InitPurchase(w http.ResponseWriter, r *http.Request) {
	if h.solana == nil {
		response.Error(w, apierror.Internal("Solana integration not configured"))
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	var req struct {
		DatasetID string `json:"dataset_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequest("invalid request body"))
		return
	}

	datasetID, err := uuid.Parse(req.DatasetID)
	if err != nil {
		response.Error(w, apierror.BadRequest("invalid dataset_id"))
		return
	}

	resp, err := h.solana.InitPurchase(r.Context(), userID, datasetID)
	if err != nil {
		response.Error(w, solErr(err))
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

// ConfirmPurchase verifies the escrow deposit and completes the purchase.
func (h *SolanaHandler) ConfirmPurchase(w http.ResponseWriter, r *http.Request) {
	if h.solana == nil {
		response.Error(w, apierror.Internal("Solana integration not configured"))
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	var req struct {
		TxSignature string `json:"tx_signature"`
		DatasetID   string `json:"dataset_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequest("invalid request body"))
		return
	}
	if req.TxSignature == "" || req.DatasetID == "" {
		response.Error(w, apierror.BadRequest("tx_signature and dataset_id are required"))
		return
	}

	datasetID, err := uuid.Parse(req.DatasetID)
	if err != nil {
		response.Error(w, apierror.BadRequest("invalid dataset_id"))
		return
	}

	resp, err := h.solana.ConfirmPurchase(r.Context(), userID, req.TxSignature, datasetID)
	if err != nil {
		response.Error(w, solErr(err))
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

// GetTransactions returns the user's Solana transaction history.
func (h *SolanaHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	if h.solana == nil {
		response.Error(w, apierror.Internal("Solana integration not configured"))
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	txs, total, err := h.solana.GetTransactions(r.Context(), userID, limit, offset)
	if err != nil {
		response.Error(w, solErr(err))
		return
	}
	response.JSONWithMeta(w, http.StatusOK, txs, map[string]interface{}{"total": total})
}

// Unused but keeps chi happy for URL params
var _ = chi.URLParam
