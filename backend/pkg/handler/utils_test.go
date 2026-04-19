package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ybapat/screener/backend/pkg/apierror"
)

func TestParseUUIDParam(t *testing.T) {
	tests := []struct {
		name      string
		paramName string
		paramValue string
		wantError bool
	}{
		{
			name:       "valid UUID",
			paramName:  "id",
			paramValue: "550e8400-e29b-41d4-a716-446655440000",
			wantError:  false,
		},
		{
			name:       "invalid UUID format",
			paramName:  "id",
			paramValue: "invalid-uuid",
			wantError:  true,
		},
		{
			name:       "empty UUID",
			paramName:  "id",
			paramValue: "",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			rctx := chi.NewRouteContext()
			if tt.paramValue != "" {
				rctx.URLParams.Add(tt.paramName, tt.paramValue)
			}
			r = r.WithContext(chi.NewRouteContext().WithValue(chi.RouteCtxKey, rctx))

			id, err := ParseUUIDParam(r, tt.paramName)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if id != uuid.Nil {
					t.Errorf("expected uuid.Nil on error, got %v", id)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				expectedUUID, _ := uuid.Parse(tt.paramValue)
				if id != expectedUUID {
					t.Errorf("got UUID %v, want %v", id, expectedUUID)
				}
			}
		})
	}
}

func TestParseIntQuery(t *testing.T) {
	tests := []struct {
		name         string
		queryParam   string
		queryValue   string
		defaultValue int
		want         int
	}{
		{
			name:         "valid integer",
			queryParam:   "limit",
			queryValue:   "50",
			defaultValue: 20,
			want:         50,
		},
		{
			name:         "missing parameter",
			queryParam:   "limit",
			queryValue:   "",
			defaultValue: 20,
			want:         20,
		},
		{
			name:         "invalid integer",
			queryParam:   "limit",
			queryValue:   "abc",
			defaultValue: 20,
			want:         20,
		},
		{
			name:         "negative integer",
			queryParam:   "limit",
			queryValue:   "-10",
			defaultValue: 20,
			want:         -10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/"
			if tt.queryValue != "" {
				url += "?" + tt.queryParam + "=" + tt.queryValue
			}
			r := httptest.NewRequest("GET", url, nil)

			got := ParseIntQuery(r, tt.queryParam, tt.defaultValue)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantLimit  int
		wantOffset int
	}{
		{
			name:       "default values",
			url:        "/",
			wantLimit:  20,
			wantOffset: 0,
		},
		{
			name:       "custom valid values",
			url:        "/?limit=50&offset=100",
			wantLimit:  50,
			wantOffset: 100,
		},
		{
			name:       "exceeds max limit",
			url:        "/?limit=200",
			wantLimit:  100,
			wantOffset: 0,
		},
		{
			name:       "negative limit",
			url:        "/?limit=-10",
			wantLimit:  20,
			wantOffset: 0,
		},
		{
			name:       "negative offset",
			url:        "/?offset=-10",
			wantLimit:  20,
			wantOffset: 0,
		},
		{
			name:       "invalid values",
			url:        "/?limit=abc&offset=xyz",
			wantLimit:  20,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.url, nil)
			params := ParsePagination(r)

			if params.Limit != tt.wantLimit {
				t.Errorf("limit: got %d, want %d", params.Limit, tt.wantLimit)
			}
			if params.Offset != tt.wantOffset {
				t.Errorf("offset: got %d, want %d", params.Offset, tt.wantOffset)
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		wantError bool
	}{
		{
			name:      "valid value",
			fieldName: "username",
			value:     "john",
			wantError: false,
		},
		{
			name:      "empty value",
			fieldName: "username",
			value:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.fieldName, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("got error %v, wantError %v", err, tt.wantError)
			}
			if err != nil {
				if _, ok := err.(*apierror.APIError); !ok {
					t.Error("expected APIError type")
				}
			}
		})
	}
}

func TestValidatePositive(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     int
		wantError bool
	}{
		{
			name:      "positive value",
			fieldName: "count",
			value:     10,
			wantError: false,
		},
		{
			name:      "zero",
			fieldName: "count",
			value:     0,
			wantError: true,
		},
		{
			name:      "negative",
			fieldName: "count",
			value:     -5,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositive(tt.fieldName, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("got error %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidateNonNegative(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     int
		wantError bool
	}{
		{
			name:      "positive value",
			fieldName: "count",
			value:     10,
			wantError: false,
		},
		{
			name:      "zero",
			fieldName: "count",
			value:     0,
			wantError: false,
		},
		{
			name:      "negative",
			fieldName: "count",
			value:     -5,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNonNegative(tt.fieldName, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("got error %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		fallbackMsg  string
		wantStatus   int
	}{
		{
			name:        "nil error",
			err:         nil,
			fallbackMsg: "something failed",
			wantStatus:  0,
		},
		{
			name:        "API error",
			err:         apierror.BadRequest("invalid input"),
			fallbackMsg: "something failed",
			wantStatus:  400,
		},
		{
			name:        "generic error",
			err:         http.ErrMissingFile,
			fallbackMsg: "operation failed",
			wantStatus:  500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			HandleError(w, tt.err, tt.fallbackMsg)

			if tt.wantStatus == 0 {
				if w.Code != 200 && w.Code != 0 {
					t.Errorf("expected no response for nil error, got status %d", w.Code)
				}
			} else {
				if w.Code != tt.wantStatus {
					t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
				}
			}
		})
	}
}
