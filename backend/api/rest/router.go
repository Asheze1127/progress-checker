package rest

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Asheze1127/progress-checker/backend/api/middleware"
	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/api/rest/handlers"
	"github.com/Asheze1127/progress-checker/backend/api/webhook"
	"github.com/Asheze1127/progress-checker/backend/application/service/jwt"
	pkgslack "github.com/Asheze1127/progress-checker/backend/pkg/slack"
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

// RouterConfig holds all dependencies needed to create the Gin router.
type RouterConfig struct {
	AuthHandler        *handlers.AuthHandler
	ProgressHandler    *handlers.ProgressHandler
	GitHubHandler      *handlers.GitHubHandler
	InternalHandler    *handlers.InternalHandler
	WebhookHandler     *webhook.WebhookHandler
	QuestionHandler    *webhook.QuestionHandler
	EventHandler       *webhook.EventHandler
	InteractionHandler *webhook.InteractionHandler
	JWTService         *jwt.JWTService
	SlackVerifier      *pkgslack.Verifier
	InternalToken      string
	CORSAllowedOrigins []string
}

// NewRouter creates and configures the Gin router with all routes.
// OpenAPI-defined routes are registered automatically via RegisterHandlersWithOptions.
// Authentication is handled by SecurityMiddleware, which dispatches based on
// the security scopes set by the generated code.
func NewRouter(cfg RouterConfig) http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	// --- Health ---
	r.GET("/healthz", handleHealthz)

	// --- OpenAPI routes (auto-registered with security + CORS middleware) ---
	composite := &compositeHandler{
		auth:     cfg.AuthHandler,
		progress: cfg.ProgressHandler,
		github:   cfg.GitHubHandler,
		internal: cfg.InternalHandler,
	}
	si := openapi.NewStrictHandler(composite, nil)

	// Middlewares are called by the generated wrapper in a loop with c.IsAborted()
	// checks between each call. They must NOT call c.Next() — the wrapper drives
	// the chain and calls the handler directly after all middlewares pass.
	openapi.RegisterHandlersWithOptions(r, si, openapi.GinServerOptions{
		Middlewares: []openapi.MiddlewareFunc{
			CORSMiddleware(cfg.CORSAllowedOrigins),
			middleware.SecurityMiddleware(cfg.JWTService, cfg.InternalToken),
		},
		ErrorHandler: func(c *gin.Context, err error, statusCode int) {
			slog.Warn("request parameter error", slog.String("error", err.Error()), slog.String("path", c.Request.URL.Path))
			c.JSON(statusCode, gin.H{"error": "invalid request parameters"})
		},
	})

	// --- CORS preflight for API routes ---
	registerPreflightRoutes(r, cfg.CORSAllowedOrigins)

	// --- Webhook routes (Slack, with signature verification) ---
	// RetryRejection runs first to avoid unnecessary HMAC verification on retries.
	slackGroup := r.Group("/webhook/slack",
		middleware.SlackRetryRejection(),
		middleware.SlackVerification(cfg.SlackVerifier),
	)
	{
		slackGroup.POST("", cfg.WebhookHandler.HandleWebhook)
		slackGroup.POST("/questions", cfg.QuestionHandler.HandleWebhook)
		slackGroup.POST("/events", cfg.EventHandler.HandleSlackEvents)
		slackGroup.POST("/interactions", cfg.InteractionHandler.HandleInteraction)
	}

	return r
}

// setCORSHeaders sets CORS headers on the response if the request Origin matches an allowed origin.
func setCORSHeaders(c *gin.Context, allowedOrigins []string) {
	c.Header("Vary", "Origin")
	origin := c.GetHeader("Origin")
	if origin == "" {
		return
	}
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Max-Age", "86400")
			return
		}
	}
}

// CORSMiddleware returns an oapi-codegen MiddlewareFunc that sets CORS headers for allowed origins.
func CORSMiddleware(allowedOrigins []string) openapi.MiddlewareFunc {
	return func(c *gin.Context) {
		setCORSHeaders(c, allowedOrigins)
	}
}

// registerPreflightRoutes registers OPTIONS handlers for CORS preflight requests.
func registerPreflightRoutes(r *gin.Engine, allowedOrigins []string) {
	corsHandler := func(c *gin.Context) {
		setCORSHeaders(c, allowedOrigins)
		c.Status(http.StatusNoContent)
	}

	r.OPTIONS("/api/v1/auth/login", corsHandler)
	r.OPTIONS("/api/v1/progress", corsHandler)
	r.OPTIONS("/api/v1/teams/:teamId/github-repos", corsHandler)
	r.OPTIONS("/api/v1/teams/:teamId/github-repos/:repoId", corsHandler)
	r.OPTIONS("/api/v1/teams/:teamId/github-repos/:repoId/token", corsHandler)
}

func handleHealthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
