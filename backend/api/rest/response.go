package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// WriteJSON writes a JSON response with the given status code and payload.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON response", slog.String("error", err.Error()))
	}
}

// WriteError writes a JSON error response with the given status code and message.
func WriteError(w http.ResponseWriter, status int, message string) {
	resp := map[string]string{"error": message}
	WriteJSON(w, status, resp)
}

// errorResponse represents a JSON error response.
type errorResponse struct {
	Error string `json:"error"`
}
