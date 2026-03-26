package rest

import "net/http"

// NewRouter creates and configures the HTTP router with all routes.
func NewRouter(
	webhookHandler *WebhookHandler,
	questionHandler *QuestionHandler,
	progressHandler *ProgressHandler,
	authHandler *AuthHandler,
	ghHandler *GitHubHandler,
	internalHandler *InternalHandler,
	eventHandler *EventHandler,
	interactionHandler *InteractionHandler,
	authMiddleware func(http.Handler) http.Handler,
	slackMiddleware func(http.Handler) http.Handler,
	internalMiddleware func(http.Handler) http.Handler,
) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /healthz", handleHealthz)

	// Auth (public)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.HandleLogin)

	// Slack webhooks (with verification + idempotency)
	mux.Handle("POST /webhook/slack", slackMiddleware(http.HandlerFunc(webhookHandler.HandleWebhook)))
	mux.Handle("POST /webhook/slack/questions", slackMiddleware(http.HandlerFunc(questionHandler.HandleWebhook)))
	mux.Handle("POST /webhook/slack/events", slackMiddleware(http.HandlerFunc(eventHandler.HandleSlackEvents)))
	mux.Handle("POST /webhook/slack/interactions", slackMiddleware(http.HandlerFunc(interactionHandler.HandleInteraction)))

	// Progress REST API (with auth)
	mux.Handle("GET /api/v1/progress", authMiddleware(http.HandlerFunc(progressHandler.HandleListProgress)))
	mux.Handle("OPTIONS /api/v1/progress", http.HandlerFunc(progressHandler.HandleProgressPreflight))

	// GitHub repo management (with auth)
	mux.Handle("POST /api/v1/teams/{teamId}/github-repos", authMiddleware(http.HandlerFunc(ghHandler.RegisterRepository)))
	mux.Handle("GET /api/v1/teams/{teamId}/github-repos", authMiddleware(http.HandlerFunc(ghHandler.ListRepositories)))
	mux.Handle("DELETE /api/v1/teams/{teamId}/github-repos/{repoId}", authMiddleware(http.HandlerFunc(ghHandler.RemoveRepository)))
	mux.Handle("PUT /api/v1/teams/{teamId}/github-repos/{repoId}/token", authMiddleware(http.HandlerFunc(ghHandler.UpdateToken)))

	// Internal API (token-protected - accessed via internal ALB)
	mux.Handle("POST /internal/issues", internalMiddleware(http.HandlerFunc(internalHandler.CreateIssue)))

	return mux
}

// handleHealthz returns a 200 OK response for health checks.
func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
