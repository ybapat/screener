package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ybapat/screener/backend/internal/service"
	"github.com/ybapat/screener/backend/pkg/apierror"
	"github.com/ybapat/screener/backend/pkg/response"
	"github.com/ybapat/screener/backend/pkg/validator"
)

type DatasetHandler struct {
	marketplace   *service.MarketplaceService
	anonymization *service.AnonymizationService
}

func NewDatasetHandler(mp *service.MarketplaceService, anon *service.AnonymizationService) *DatasetHandler {
	return &DatasetHandler{marketplace: mp, anonymization: anon}
}

func (h *DatasetHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	var categories []string
	if cat := r.URL.Query().Get("categories"); cat != "" {
		categories = strings.Split(cat, ",")
	}

	datasets, total, err := h.marketplace.ListDatasets(r.Context(), categories, limit, offset)
	if err != nil {
		response.Error(w, apierror.Internal("failed to list datasets"))
		return
	}

	response.JSONWithMeta(w, http.StatusOK, datasets, map[string]int{"total": total})
}

func (h *DatasetHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apierror.BadRequest("invalid dataset id"))
		return
	}

	dataset, err := h.marketplace.GetDataset(r.Context(), id)
	if err != nil || dataset == nil {
		response.Error(w, apierror.NotFound("dataset not found"))
		return
	}

	samples, _ := h.marketplace.GetSamples(r.Context(), id)

	response.JSON(w, http.StatusOK, map[string]any{
		"dataset": dataset,
		"samples": samples,
	})
}

func (h *DatasetHandler) GetSamples(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apierror.BadRequest("invalid dataset id"))
		return
	}

	samples, err := h.marketplace.GetSamples(r.Context(), id)
	if err != nil {
		response.Error(w, apierror.Internal("failed to get samples"))
		return
	}

	response.JSON(w, http.StatusOK, samples)
}

// AssembleDataset triggers the anonymization pipeline to create a new dataset.
func (h *DatasetHandler) Assemble(w http.ResponseWriter, r *http.Request) {
	var req service.AssembleParams
	if err := validator.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, apierror.BadRequest(err.Error()))
		return
	}

	result, err := h.anonymization.AssembleDataset(r.Context(), req)
	if err != nil {
		response.Error(w, apierror.BadRequest(err.Error()))
		return
	}

	response.JSON(w, http.StatusCreated, result)
}
