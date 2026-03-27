package rest

import (
  "context"
  "encoding/json"
  "log/slog"
  "net/http"

  "github.com/Asheze1127/progress-checker/backend/api/openapi"
  "github.com/Asheze1127/progress-checker/backend/api/rest/handlers"
  "github.com/Asheze1127/progress-checker/backend/api/webhook"
)

// compositeHandler combines individual handlers into openapi.StrictServerInterface.
type compositeHandler struct {
  auth     *handlers.AuthHandler
  progress *handlers.ProgressHandler
  github   *handlers.GitHubHandler
  internal *handlers.InternalHandler
}

var _ openapi.StrictServerInterface = (*compositeHandler)(nil)

func (c *compositeHandler) Login(ctx context.Context, req openapi.LoginRequestObject) (openapi.LoginResponseObject, error) {
  return c.auth.Login(ctx, req)
}

func (c *compositeHandler) ListProgress(ctx context.Context, req openapi.ListProgressRequestObject) (openapi.ListProgressResponseObject, error) {
  return c.progress.ListProgress(ctx, req)
}

func (c *compositeHandler) ListRepositories(ctx context.Context, req openapi.ListRepositoriesRequestObject) (openapi.ListRepositoriesResponseObject, error) {
  return c.github.ListRepositories(ctx, req)
}

func (c *compositeHandler) RegisterRepository(ctx context.Context, req openapi.RegisterRepositoryRequestObject) (openapi.RegisterRepositoryResponseObject, error) {
  return c.github.RegisterRepository(ctx, req)
}

func (c *compositeHandler) RemoveRepository(ctx context.Context, req openapi.RemoveRepositoryRequestObject) (openapi.RemoveRepositoryResponseObject, error) {
  return c.github.RemoveRepository(ctx, req)
}

func (c *compositeHandler) UpdateToken(ctx context.Context, req openapi.UpdateTokenRequestObject) (openapi.UpdateTokenResponseObject, error) {
  return c.github.UpdateToken(ctx, req)
}

func (c *compositeHandler) CreateIssue(ctx context.Context, req openapi.CreateIssueRequestObject) (openapi.CreateIssueResponseObject, error) {
  return c.internal.CreateIssue(ctx, req)
}

// NewRouter creates and configures the HTTP router with all routes.
func NewRouter(
  authHandler *handlers.AuthHandler,
  progressHandler *handlers.ProgressHandler,
  githubHandler *handlers.GitHubHandler,
  internalHandler *handlers.InternalHandler,
  webhookHandler *webhook.WebhookHandler,
  questionHandler *webhook.QuestionHandler,
  eventHandler *webhook.EventHandler,
  interactionHandler *webhook.InteractionHandler,
  authMiddleware func(http.Handler) http.Handler,
  slackMiddleware func(http.Handler) http.Handler,
  internalMiddleware func(http.Handler) http.Handler,
  corsMiddleware func(http.Handler) http.Handler,
) http.Handler {
  mux := http.NewServeMux()

  // Build StrictServerInterface from individual handlers.
  composite := &compositeHandler{
    auth:     authHandler,
    progress: progressHandler,
    github:   githubHandler,
    internal: internalHandler,
  }
  si := openapi.NewStrictHandler(composite, nil)

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

  // --- Health ---
  mux.HandleFunc("GET /healthz", handleHealthz)

  // --- REST API routes (browser-facing, with auth + CORS) ---
  registerAPIRoutes(mux, &wrapper, authMiddleware, corsMiddleware)

  // --- Webhook routes (Slack, with signature verification) ---
  registerWebhookRoutes(mux, webhookHandler, questionHandler, eventHandler, interactionHandler, slackMiddleware)

  // --- Internal routes (service-to-service, with token auth) ---
  registerInternalRoutes(mux, &wrapper, internalMiddleware)

  return mux
}

// registerAPIRoutes registers browser-facing REST API routes with auth and CORS middleware.
func registerAPIRoutes(
  mux *http.ServeMux,
  wrapper *openapi.ServerInterfaceWrapper,
  authMiddleware func(http.Handler) http.Handler,
  corsMiddleware func(http.Handler) http.Handler,
) {
  // Auth (public - no auth required)
  mux.HandleFunc("POST /api/v1/auth/login", wrapper.Login)

  // Progress
  mux.Handle("GET /api/v1/progress", corsMiddleware(authMiddleware(http.HandlerFunc(wrapper.ListProgress))))
  mux.Handle("OPTIONS /api/v1/progress", corsMiddleware(http.HandlerFunc(handlePreflight)))

  // GitHub repository management
  mux.Handle("POST /api/v1/teams/{teamId}/github-repos", corsMiddleware(authMiddleware(http.HandlerFunc(wrapper.RegisterRepository))))
  mux.Handle("GET /api/v1/teams/{teamId}/github-repos", corsMiddleware(authMiddleware(http.HandlerFunc(wrapper.ListRepositories))))
  mux.Handle("DELETE /api/v1/teams/{teamId}/github-repos/{repoId}", corsMiddleware(authMiddleware(http.HandlerFunc(wrapper.RemoveRepository))))
  mux.Handle("PUT /api/v1/teams/{teamId}/github-repos/{repoId}/token", corsMiddleware(authMiddleware(http.HandlerFunc(wrapper.UpdateToken))))
  mux.Handle("OPTIONS /api/v1/teams/{teamId}/github-repos", corsMiddleware(http.HandlerFunc(handlePreflight)))
  mux.Handle("OPTIONS /api/v1/teams/{teamId}/github-repos/{repoId}", corsMiddleware(http.HandlerFunc(handlePreflight)))
  mux.Handle("OPTIONS /api/v1/teams/{teamId}/github-repos/{repoId}/token", corsMiddleware(http.HandlerFunc(handlePreflight)))
}

// registerWebhookRoutes registers Slack webhook routes with signature verification middleware.
func registerWebhookRoutes(
  mux *http.ServeMux,
  webhookHandler *webhook.WebhookHandler,
  questionHandler *webhook.QuestionHandler,
  eventHandler *webhook.EventHandler,
  interactionHandler *webhook.InteractionHandler,
  slackMiddleware func(http.Handler) http.Handler,
) {
  mux.Handle("POST /webhook/slack", slackMiddleware(http.HandlerFunc(webhookHandler.HandleWebhook)))
  mux.Handle("POST /webhook/slack/questions", slackMiddleware(http.HandlerFunc(questionHandler.HandleWebhook)))
  mux.Handle("POST /webhook/slack/events", slackMiddleware(http.HandlerFunc(eventHandler.HandleSlackEvents)))
  mux.Handle("POST /webhook/slack/interactions", slackMiddleware(http.HandlerFunc(interactionHandler.HandleInteraction)))
}

// registerInternalRoutes registers internal service-to-service routes with token auth middleware.
func registerInternalRoutes(
  mux *http.ServeMux,
  wrapper *openapi.ServerInterfaceWrapper,
  internalMiddleware func(http.Handler) http.Handler,
) {
  mux.Handle("POST /internal/issues", internalMiddleware(http.HandlerFunc(wrapper.CreateIssue)))
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  _, _ = w.Write([]byte(`{"status":"ok"}`))
}

func handlePreflight(w http.ResponseWriter, _ *http.Request) {
  w.WriteHeader(http.StatusNoContent)
}
