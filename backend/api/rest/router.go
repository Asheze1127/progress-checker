package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/api/webhook"
)

// NewRouter creates and configures the HTTP router with all routes.
func NewRouter(
	webhookHandler *webhook.WebhookHandler,
	questionHandler *webhook.QuestionHandler,
	strictHandler *StrictHandler,
	eventHandler *webhook.EventHandler,
	interactionHandler *webhook.InteractionHandler,
	authMiddleware func(http.Handler) http.Handler,
	slackMiddleware func(http.Handler) http.Handler,
	internalMiddleware func(http.Handler) http.Handler,
	corsMiddleware func(http.Handler) http.Handler,
) http.Handler {
	mux := http.NewServeMux()

	// Create ServerInterface from StrictServerInterface via generated adapter.
	si := openapi.NewStrictHandler(strictHandler, nil)

	// Wrap to get parameter-extracting handlers.
	wrapper := openapi.ServerInterfaceWrapper{
		Handler: si,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Warn("request parameter error", slog.String("error", err.Error()), slog.String("path", r.URL.Path))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			body, _ := json.Marshal(map[string]string{"error": "invalid request parameters"})
			_, _ = w.Write(body)
		},
	}

	// Health check
	mux.HandleFunc("GET /healthz", handleHealthz)

	// Auth (public)
	mux.HandleFunc("POST /api/v1/auth/login", wrapper.Login)

	// Slack webhooks (with verification + retry rejection)
	mux.Handle("POST /webhook/slack", slackMiddleware(http.HandlerFunc(webhookHandler.HandleWebhook)))
	mux.Handle("POST /webhook/slack/questions", slackMiddleware(http.HandlerFunc(questionHandler.HandleWebhook)))
	mux.Handle("POST /webhook/slack/events", slackMiddleware(http.HandlerFunc(eventHandler.HandleSlackEvents)))
	mux.Handle("POST /webhook/slack/interactions", slackMiddleware(http.HandlerFunc(interactionHandler.HandleInteraction)))

	// Progress REST API (with auth + CORS)
	mux.Handle("GET /api/v1/progress", corsMiddleware(authMiddleware(http.HandlerFunc(wrapper.ListProgress))))
	mux.Handle("OPTIONS /api/v1/progress", corsMiddleware(http.HandlerFunc(handlePreflight)))

	// GitHub repo management (with auth + CORS)
	mux.Handle("POST /api/v1/teams/{teamId}/github-repos", corsMiddleware(authMiddleware(http.HandlerFunc(wrapper.RegisterRepository))))
	mux.Handle("GET /api/v1/teams/{teamId}/github-repos", corsMiddleware(authMiddleware(http.HandlerFunc(wrapper.ListRepositories))))
	mux.Handle("DELETE /api/v1/teams/{teamId}/github-repos/{repoId}", corsMiddleware(authMiddleware(http.HandlerFunc(wrapper.RemoveRepository))))
	mux.Handle("PUT /api/v1/teams/{teamId}/github-repos/{repoId}/token", corsMiddleware(authMiddleware(http.HandlerFunc(wrapper.UpdateToken))))
	mux.Handle("OPTIONS /api/v1/teams/{teamId}/github-repos", corsMiddleware(http.HandlerFunc(handlePreflight)))
	mux.Handle("OPTIONS /api/v1/teams/{teamId}/github-repos/{repoId}", corsMiddleware(http.HandlerFunc(handlePreflight)))
	mux.Handle("OPTIONS /api/v1/teams/{teamId}/github-repos/{repoId}/token", corsMiddleware(http.HandlerFunc(handlePreflight)))

	// Internal API (token-protected - accessed via internal ALB)
	mux.Handle("POST /internal/issues", internalMiddleware(http.HandlerFunc(wrapper.CreateIssue)))

	return mux
}

// handleHealthz returns a 200 OK response for health checks.
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// handlePreflight handles CORS preflight requests.
func handlePreflight(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
