package rest

import "net/http"

// NewRouter creates and configures the HTTP router with all routes.
func NewRouter(webhookHandler *WebhookHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /webhook/slack", webhookHandler.HandleWebhook)
	mux.HandleFunc("GET /healthz", handleHealthz)

	return mux
}

// handleHealthz returns a 200 OK response for health checks.
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
