package response

import (
	"encoding/json"
	"net/http"

	"github.com/ybapat/screener/backend/pkg/apierror"
)

type Envelope struct {
	Data  any    `json:"data,omitempty"`
	Error any    `json:"error,omitempty"`
	Meta  any    `json:"meta,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Envelope{Data: data})
}

func JSONWithMeta(w http.ResponseWriter, status int, data any, meta any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Envelope{Data: data, Meta: meta})
}

func Error(w http.ResponseWriter, err *apierror.APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(Envelope{Error: err})
}

func ErrorMsg(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Envelope{
		Error: map[string]string{"message": msg},
	})
}
