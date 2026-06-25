package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Response struct {
	Data any `json:"data,omitempty"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Write(w http.ResponseWriter, r *http.Request, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode response", "error", err)
	}
}

func WriteOK(w http.ResponseWriter, r *http.Request, data any) {
	Write(w, r, http.StatusOK, data)
}

func WriteError(w http.ResponseWriter, r *http.Request, code int, errCode, message string) {
	Write(w, r, code, ErrorResponse{
		Error: ErrorBody{
			Code:    errCode,
			Message: message,
		},
	})
}
