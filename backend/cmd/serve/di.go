package serve

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"

	_ "github.com/lib/pq"
	"github.com/samber/do/v2"

	"github.com/Asheze1127/progress-checker/backend/api/middleware"
	"github.com/Asheze1127/progress-checker/backend/api/rest"
	"github.com/Asheze1127/progress-checker/backend/application/port"
	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/encryption"
	githubinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/github"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/postgres"
	slackinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/slack"
	pkgslack "github.com/Asheze1127/progress-checker/backend/pkg/slack"
	"github.com/Asheze1127/progress-checker/backend/util"
)

// wireRouter builds all dependencies using samber/do DI container
// and returns the configured HTTP router.
func wireRouter(cfg *util.Config) (http.Handler, error) {
	injector := do.New()

	// --- Config ---
	do.ProvideValue(injector, cfg)

	// --- Infrastructure ---
	do.Provide(injector, func(i do.Injector) (*sql.DB, error) {
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			url.PathEscape(cfg.DatabaseUser),
			url.PathEscape(cfg.DatabasePass),
			url.PathEscape(cfg.DatabaseHost),
			url.PathEscape(cfg.DatabasePort),
			url.PathEscape(cfg.DatabaseName),
			url.QueryEscape(cfg.DatabaseSSLMode),
		)
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
		if err := db.Ping(); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		return db, nil
	})

	// Register infrastructure as interface types for service consumption
	do.Provide[service.SlackClient](injector, func(i do.Injector) (service.SlackClient, error) {
		return slackinfra.NewClient(cfg.SlackBotToken), nil
	})

	do.Provide[port.TokenEncryptor](injector, func(i do.Injector) (port.TokenEncryptor, error) {
		return encryption.NewAESEncryptor([]byte(cfg.EncryptionKey))
	})

	do.Provide[port.GitHubIssueCreator](injector, func(i do.Injector) (port.GitHubIssueCreator, error) {
		return githubinfra.NewClient(cfg.GitHubAPIBaseURL), nil
	})

	do.Provide[port.MessageQueue](injector, func(_ do.Injector) (port.MessageQueue, error) {
		return &service.NoopMessageQueue{}, nil
	})

	do.Provide[service.SlackNotifier](injector, func(_ do.Injector) (service.SlackNotifier, error) {
		return &service.NoopSlackNotifier{}, nil
	})

	do.Provide[service.SlackThreadFetcher](injector, func(_ do.Injector) (service.SlackThreadFetcher, error) {
		return &service.NoopSlackThreadFetcher{}, nil
	})

	do.Provide(injector, func(i do.Injector) (*pkgslack.Verifier, error) {
		return pkgslack.NewVerifier(cfg.SlackSigningSecret), nil
	})

	// --- Repositories (registered as interface types) ---
	do.Provide[entities.ProgressRepository](injector, func(i do.Injector) (entities.ProgressRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewProgressRepository(db), nil
	})

	do.Provide[entities.ProgressQueryRepository](injector, func(i do.Injector) (entities.ProgressQueryRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewProgressQueryRepository(db), nil
	})

	do.Provide[entities.QuestionRepository](injector, func(i do.Injector) (entities.QuestionRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewQuestionRepository(db), nil
	})

	do.Provide[entities.UserRepository](injector, func(i do.Injector) (entities.UserRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewUserRepository(db), nil
	})

	do.Provide[port.GitHubRepoRepository](injector, func(i do.Injector) (port.GitHubRepoRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewGitHubRepoRepository(db), nil
	})

	// --- Services ---
	do.Provide(injector, service.NewProgressFormatter)
	do.Provide(injector, service.NewPasswordHasher)
	do.Provide(injector, service.NewJWTService)
	do.Provide(injector, service.NewSlackPoster)
	do.Provide(injector, service.NewGitHubService)
	do.Provide(injector, service.NewQuestionSender)

	// --- Use Cases ---
	do.Provide(injector, usecase.NewHandleProgressUseCase)
	do.Provide(injector, usecase.NewListProgressUseCase)
	do.Provide(injector, usecase.NewLoginUseCase)
	do.Provide(injector, usecase.NewResolveQuestionUseCase)
	do.Provide(injector, usecase.NewContinueQuestionUseCase)
	do.Provide(injector, usecase.NewEscalateQuestionUseCase)
	do.Provide(injector, usecase.NewHandleNewQuestionUseCase)
	do.Provide(injector, usecase.NewTriggerIssueCreationUseCase)

	// --- Handlers ---
	do.Provide(injector, rest.NewWebhookHandler)
	do.Provide(injector, rest.NewQuestionHandler)
	do.Provide(injector, rest.NewProgressHandler)
	do.Provide(injector, rest.NewAuthHandler)
	do.Provide(injector, rest.NewGitHubHandler)
	do.Provide(injector, rest.NewInternalHandler)
	do.Provide(injector, rest.NewEventHandler)
	do.Provide(injector, rest.NewInteractionHandler)

	// --- Router ---
	do.Provide(injector, func(i do.Injector) (http.Handler, error) {
		webhookHandler := do.MustInvoke[*rest.WebhookHandler](i)
		questionHandler := do.MustInvoke[*rest.QuestionHandler](i)
		progressHandler := do.MustInvoke[*rest.ProgressHandler](i)
		authHandler := do.MustInvoke[*rest.AuthHandler](i)
		ghHandler := do.MustInvoke[*rest.GitHubHandler](i)
		internalHandler := do.MustInvoke[*rest.InternalHandler](i)
		eventHandler := do.MustInvoke[*rest.EventHandler](i)
		interactionHandler := do.MustInvoke[*rest.InteractionHandler](i)

		jwtService := do.MustInvoke[*service.JWTService](i)
		verifier := do.MustInvoke[*pkgslack.Verifier](i)

		return rest.NewRouter(
			webhookHandler,
			questionHandler,
			progressHandler,
			authHandler,
			ghHandler,
			internalHandler,
			eventHandler,
			interactionHandler,
			middleware.AuthMiddleware(jwtService),
			middleware.SlackWebhookMiddleware(verifier),
			middleware.InternalTokenMiddleware(cfg.InternalToken),
		), nil
	})

	router, err := do.Invoke[http.Handler](injector)
	if err != nil {
		return nil, fmt.Errorf("failed to build router: %w", err)
	}

	return router, nil
}
