package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ybapat/screener/backend/internal/middleware"
	"github.com/ybapat/screener/backend/internal/models"
	"github.com/ybapat/screener/backend/internal/service"
	"github.com/ybapat/screener/backend/pkg/apierror"
	"github.com/ybapat/screener/backend/pkg/response"
	"github.com/ybapat/screener/backend/pkg/validator"
)

type IngestionHandler struct {
	ingestion *service.IngestionService
}

func NewIngestionHandler(ingestion *service.IngestionService) *IngestionHandler {
	return &IngestionHandler{ingestion: ingestion}
}

func (h *IngestionHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req models.ScreenTimeUploadRequest
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, apierror.BadRequest(err.Error()))
		return
	}

	result, err := h.ingestion.Upload(r.Context(), userID, req.Records)
	if err != nil {
		if apiErr, ok := err.(*apierror.APIError); ok {
			response.Error(w, apiErr)
			return
		}
		response.Error(w, apierror.Internal("upload failed"))
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *IngestionHandler) ListBatches(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	limit, offset := parsePagination(r)

	batches, total, err := h.ingestion.GetBatches(r.Context(), userID, limit, offset)
	if err != nil {
		response.Error(w, apierror.Internal("failed to list batches"))
		return
	}

	response.JSONWithMeta(w, http.StatusOK, batches, map[string]int{"total": total, "limit": limit, "offset": offset})
}

func (h *IngestionHandler) GetBatch(w http.ResponseWriter, r *http.Request) {
	batchID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apierror.BadRequest("invalid batch id"))
		return
	}

	batch, err := h.ingestion.GetBatch(r.Context(), batchID)
	if err != nil {
		response.Error(w, apierror.Internal("failed to get batch"))
		return
	}

	records, err := h.ingestion.GetBatchRecords(r.Context(), batchID)
	if err != nil {
		response.Error(w, apierror.Internal("failed to get records"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"batch":   batch,
		"records": records,
	})
}

func (h *IngestionHandler) WithdrawBatch(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	batchID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apierror.BadRequest("invalid batch id"))
		return
	}

	if err := h.ingestion.WithdrawBatch(r.Context(), batchID, userID); err != nil {
		if apiErr, ok := err.(*apierror.APIError); ok {
			response.Error(w, apiErr)
			return
		}
		response.Error(w, apierror.Internal("withdrawal failed"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "batch withdrawn"})
}

func parsePagination(r *http.Request) (int, int) {
	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}
