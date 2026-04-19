package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ybapat/screener/backend/pkg/apierror"
	"github.com/ybapat/screener/backend/pkg/response"
)

// HandleError processes an error and sends appropriate HTTP response.
// If the error is an APIError, it uses that directly. Otherwise, it wraps
// it with the provided fallback message as an internal error.
func HandleError(w http.ResponseWriter, err error, fallbackMsg string) {
	if err == nil {
		return
	}

	if apiErr, ok := err.(*apierror.APIError); ok {
		response.Error(w, apiErr)
		return
	}

	response.Error(w, apierror.Internal(fallbackMsg))
}

// ParseUUIDParam extracts and parses a UUID from URL parameters.
// Returns APIError if the parameter is missing or invalid.
func ParseUUIDParam(r *http.Request, paramName string) (uuid.UUID, error) {
	paramValue := chi.URLParam(r, paramName)
	if paramValue == "" {
		return uuid.Nil, apierror.BadRequest(paramName + " is required")
	}

	id, err := uuid.Parse(paramValue)
	if err != nil {
		return uuid.Nil, apierror.BadRequest("invalid " + paramName + " format")
	}

	return id, nil
}

// ParseIntQuery extracts and parses an integer from query parameters.
// Returns the default value if the parameter is missing or invalid.
func ParseIntQuery(r *http.Request, paramName string, defaultValue int) int {
	value := r.URL.Query().Get(paramName)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// PaginationParams represents common pagination parameters
type PaginationParams struct {
	Limit  int
	Offset int
}

// ParsePagination extracts pagination parameters from query string.
// Applies sensible defaults and enforces maximum limits.
func ParsePagination(r *http.Request) PaginationParams {
	const (
		defaultLimit = 20
		maxLimit     = 100
	)

	limit := ParseIntQuery(r, "limit", defaultLimit)
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := ParseIntQuery(r, "offset", 0)
	if offset < 0 {
		offset = 0
	}

	return PaginationParams{
		Limit:  limit,
		Offset: offset,
	}
}

// ValidateRequired checks if a required field is present and non-empty.
// Returns APIError if validation fails.
func ValidateRequired(fieldName, value string) error {
	if value == "" {
		return apierror.BadRequest(fieldName + " is required")
	}
	return nil
}

// ValidatePositive checks if a value is positive.
// Returns APIError if validation fails.
func ValidatePositive(fieldName string, value int) error {
	if value <= 0 {
		return apierror.BadRequest(fieldName + " must be positive")
	}
	return nil
}

// ValidateNonNegative checks if a value is non-negative.
// Returns APIError if validation fails.
func ValidateNonNegative(fieldName string, value int) error {
	if value < 0 {
		return apierror.BadRequest(fieldName + " must be non-negative")
	}
	return nil
}
