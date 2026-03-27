package rest

import (
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
)

// NewRouter creates and configures the HTTP router with all routes.
func NewRouter(
	apiHandler *openapi.APIHandler,
	webhookHandler *WebhookHandler,
	questionHandler *QuestionHandler,
	eventHandler *EventHandler,
	interactionHandler *InteractionHandler,
	authMiddleware func(http.Handler) http.Handler,
	slackMiddleware func(http.Handler) http.Handler,
	internalMiddleware func(http.Handler) http.Handler,
) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /healthz", handleHealthz)

	// --- REST API routes (from OpenAPI spec) ---

	// Auth (public)
	mux.HandleFunc("POST /api/v1/auth/login", apiHandler.Login)

	// Progress (with auth + CORS preflight)
	mux.Handle("GET /api/v1/progress", authMiddleware(http.HandlerFunc(wrapListProgress(apiHandler))))
	mux.Handle("OPTIONS /api/v1/progress", http.HandlerFunc(apiHandler.HandleProgressPreflight))

	// GitHub repo management (with auth)
	mux.Handle("POST /api/v1/teams/{teamId}/github-repos", authMiddleware(http.HandlerFunc(wrapRegisterRepository(apiHandler))))
	mux.Handle("GET /api/v1/teams/{teamId}/github-repos", authMiddleware(http.HandlerFunc(wrapListRepositories(apiHandler))))
	mux.Handle("DELETE /api/v1/teams/{teamId}/github-repos/{repoId}", authMiddleware(http.HandlerFunc(wrapRemoveRepository(apiHandler))))
	mux.Handle("PUT /api/v1/teams/{teamId}/github-repos/{repoId}/token", authMiddleware(http.HandlerFunc(wrapUpdateToken(apiHandler))))

	// Internal API (token-protected - accessed via internal ALB)
	mux.Handle("POST /internal/issues", internalMiddleware(http.HandlerFunc(apiHandler.CreateIssue)))

	// --- Webhook routes (not in OpenAPI spec) ---

	// Slack webhooks (with verification + retry rejection)
	mux.Handle("POST /webhook/slack", slackMiddleware(http.HandlerFunc(webhookHandler.HandleWebhook)))
	mux.Handle("POST /webhook/slack/questions", slackMiddleware(http.HandlerFunc(questionHandler.HandleWebhook)))
	mux.Handle("POST /webhook/slack/events", slackMiddleware(http.HandlerFunc(eventHandler.HandleSlackEvents)))
	mux.Handle("POST /webhook/slack/interactions", slackMiddleware(http.HandlerFunc(interactionHandler.HandleInteraction)))

	return mux
}

// wrapListProgress extracts the team_id query param and calls the handler method.
func wrapListProgress(h *openapi.APIHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		teamID := r.URL.Query().Get("team_id")
		var params openapi.ListProgressParams
		if teamID != "" {
			params.TeamId = &teamID
		}
		h.ListProgress(w, r, params)
	}
}

// wrapRegisterRepository extracts the teamId path param and calls the handler method.
func wrapRegisterRepository(h *openapi.APIHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h.RegisterRepository(w, r, r.PathValue("teamId"))
	}
}

// wrapListRepositories extracts the teamId path param and calls the handler method.
func wrapListRepositories(h *openapi.APIHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ListRepositories(w, r, r.PathValue("teamId"))
	}
}

// wrapRemoveRepository extracts the teamId and repoId path params.
func wrapRemoveRepository(h *openapi.APIHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h.RemoveRepository(w, r, r.PathValue("teamId"), r.PathValue("repoId"))
	}
}

// wrapUpdateToken extracts the teamId and repoId path params.
func wrapUpdateToken(h *openapi.APIHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h.UpdateToken(w, r, r.PathValue("teamId"), r.PathValue("repoId"))
	}
}

// handleHealthz returns a 200 OK response for health checks.
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
