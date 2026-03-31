package apierror

import (
	"fmt"
	"net/http"
)

type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func BadRequest(msg string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: msg}
}

func Unauthorized(msg string) *APIError {
	return &APIError{Status: http.StatusUnauthorized, Code: "UNAUTHORIZED", Message: msg}
}

func Forbidden(msg string) *APIError {
	return &APIError{Status: http.StatusForbidden, Code: "FORBIDDEN", Message: msg}
}

func NotFound(msg string) *APIError {
	return &APIError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: msg}
}

func Conflict(msg string) *APIError {
	return &APIError{Status: http.StatusConflict, Code: "CONFLICT", Message: msg}
}

func TooManyRequests(msg string) *APIError {
	return &APIError{Status: http.StatusTooManyRequests, Code: "RATE_LIMITED", Message: msg}
}

func Internal(msg string) *APIError {
	return &APIError{Status: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: msg}
}
