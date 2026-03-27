package serve

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	_ "github.com/lib/pq"
	"github.com/samber/do/v2"

	"github.com/Asheze1127/progress-checker/backend/api/middleware"
	"github.com/Asheze1127/progress-checker/backend/api/rest"
	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/encryption"
	githubinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/github"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/idempotency"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/postgres"
	slackinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/slack"
	pkgslack "github.com/Asheze1127/progress-checker/backend/pkg/slack"
	idempotencysvc "github.com/Asheze1127/progress-checker/backend/service/idempotency"
	"github.com/Asheze1127/progress-checker/backend/util"
	"github.com/google/uuid"
)

// wireRouter builds all dependencies using samber/do DI container
// and returns the configured HTTP router.
func wireRouter(cfg *util.Config) (http.Handler, error) {
	injector := do.New()

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

	do.Provide(injector, func(i do.Injector) (*slackinfra.Client, error) {
		return slackinfra.NewClient(cfg.SlackBotToken), nil
	})

	do.Provide(injector, func(i do.Injector) (*encryption.AESEncryptor, error) {
		return encryption.NewAESEncryptor([]byte(cfg.EncryptionKey))
	})

	do.Provide(injector, func(i do.Injector) (*githubinfra.Client, error) {
		return githubinfra.NewClient(cfg.GitHubAPIBaseURL), nil
	})

	// --- Repositories ---
	do.Provide(injector, func(i do.Injector) (*postgres.ProgressRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewProgressRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.ProgressQueryRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewProgressQueryRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.QuestionRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewQuestionRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.UserRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewUserRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*postgres.GitHubRepoRepository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return postgres.NewGitHubRepoRepository(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*idempotency.PostgresStore, error) {
		db := do.MustInvoke[*sql.DB](i)
		return idempotency.NewPostgresStore(db), nil
	})

	// --- Services ---
	do.Provide(injector, func(i do.Injector) (*service.ProgressFormatter, error) {
		return service.NewProgressFormatter(), nil
	})

	do.Provide(injector, func(i do.Injector) (*service.SlackPoster, error) {
		client := do.MustInvoke[*slackinfra.Client](i)
		formatter := do.MustInvoke[*service.ProgressFormatter](i)
		return service.NewSlackPoster(client, formatter), nil
	})

	do.Provide(injector, func(i do.Injector) (*service.JWTService, error) {
		return service.NewJWTService(cfg.JWTSecret), nil
	})

	do.Provide(injector, func(i do.Injector) (*service.PasswordHasher, error) {
		return service.NewPasswordHasher(), nil
	})

	do.Provide(injector, func(i do.Injector) (*idempotencysvc.Service, error) {
		store := do.MustInvoke[*idempotency.PostgresStore](i)
		return idempotencysvc.NewService(store), nil
	})

	do.Provide(injector, func(i do.Injector) (*pkgslack.Verifier, error) {
		return pkgslack.NewVerifier(cfg.SlackSigningSecret), nil
	})

	do.Provide(injector, func(i do.Injector) (*service.GitHubService, error) {
		ghRepoRepo := do.MustInvoke[*postgres.GitHubRepoRepository](i)
		encryptor := do.MustInvoke[*encryption.AESEncryptor](i)
		ghClient := do.MustInvoke[*githubinfra.Client](i)
		return service.NewGitHubService(ghRepoRepo, encryptor, ghClient, func() string {
			return uuid.New().String()
		}), nil
	})

	// --- Use Cases ---
	do.Provide(injector, func(i do.Injector) (*usecase.HandleProgressUseCase, error) {
		progressRepo := do.MustInvoke[*postgres.ProgressRepository](i)
		poster := do.MustInvoke[*service.SlackPoster](i)
		return usecase.NewHandleProgressUseCase(progressRepo, poster), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.ListProgressUseCase, error) {
		progressQueryRepo := do.MustInvoke[*postgres.ProgressQueryRepository](i)
		return usecase.NewListProgressUseCase(progressQueryRepo), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.LoginUseCase, error) {
		userRepo := do.MustInvoke[*postgres.UserRepository](i)
		jwtService := do.MustInvoke[*service.JWTService](i)
		hasher := do.MustInvoke[*service.PasswordHasher](i)
		return usecase.NewLoginUseCase(userRepo, jwtService, hasher), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.ResolveQuestionUseCase, error) {
		questionRepo := do.MustInvoke[*postgres.QuestionRepository](i)
		return usecase.NewResolveQuestionUseCase(questionRepo), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.ContinueQuestionUseCase, error) {
		questionRepo := do.MustInvoke[*postgres.QuestionRepository](i)
		return usecase.NewContinueQuestionUseCase(questionRepo), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.EscalateQuestionUseCase, error) {
		questionRepo := do.MustInvoke[*postgres.QuestionRepository](i)
		// TODO: Wire a proper SlackNotifier implementation when available.
		return usecase.NewEscalateQuestionUseCase(questionRepo, &service.NoopSlackNotifier{}), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.HandleNewQuestionUseCase, error) {
		questionRepo := do.MustInvoke[*postgres.QuestionRepository](i)
		sender := service.NewQuestionSender(&service.NoopMessageQueue{})
		return usecase.NewHandleNewQuestionUseCase(questionRepo, sender), nil
	})

	do.Provide(injector, func(i do.Injector) (*usecase.TriggerIssueCreationUseCase, error) {
		// TODO: Wire real SlackThreadFetcher and MessageQueue implementations when available.
		return usecase.NewTriggerIssueCreationUseCase(
			&service.NoopSlackThreadFetcher{}, &service.NoopMessageQueue{},
		), nil
	})

	// --- Handlers ---
	do.Provide(injector, func(i do.Injector) (*rest.WebhookHandler, error) {
		uc := do.MustInvoke[*usecase.HandleProgressUseCase](i)
		return rest.NewWebhookHandler(uc), nil
	})

	do.Provide(injector, func(i do.Injector) (*rest.QuestionHandler, error) {
		uc := do.MustInvoke[*usecase.HandleNewQuestionUseCase](i)
		return rest.NewQuestionHandler(uc), nil
	})

	do.Provide(injector, func(i do.Injector) (*rest.ProgressHandler, error) {
		uc := do.MustInvoke[*usecase.ListProgressUseCase](i)
		var corsOrigins []string
		if cfg.CORSAllowedOrigin != "" {
			corsOrigins = strings.Split(cfg.CORSAllowedOrigin, ",")
		}
		return rest.NewProgressHandler(uc, corsOrigins), nil
	})

	do.Provide(injector, func(i do.Injector) (*rest.AuthHandler, error) {
		uc := do.MustInvoke[*usecase.LoginUseCase](i)
		return rest.NewAuthHandler(uc), nil
	})

	do.Provide(injector, func(i do.Injector) (*rest.GitHubHandler, error) {
		svc := do.MustInvoke[*service.GitHubService](i)
		return rest.NewGitHubHandler(svc), nil
	})

	do.Provide(injector, func(i do.Injector) (*rest.InternalHandler, error) {
		svc := do.MustInvoke[*service.GitHubService](i)
		return rest.NewInternalHandler(svc), nil
	})

	do.Provide(injector, func(i do.Injector) (*rest.EventHandler, error) {
		uc := do.MustInvoke[*usecase.TriggerIssueCreationUseCase](i)
		return rest.NewEventHandler(uc, cfg.IssueTriggerEmoji), nil
	})

	do.Provide(injector, func(i do.Injector) (*rest.InteractionHandler, error) {
		resolveUC := do.MustInvoke[*usecase.ResolveQuestionUseCase](i)
		continueUC := do.MustInvoke[*usecase.ContinueQuestionUseCase](i)
		escalateUC := do.MustInvoke[*usecase.EscalateQuestionUseCase](i)
		return rest.NewInteractionHandler(resolveUC, continueUC, escalateUC), nil
	})

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
		idempotencySvc := do.MustInvoke[*idempotencysvc.Service](i)

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
			middleware.SlackWebhookMiddleware(verifier, idempotencySvc),
			middleware.InternalTokenMiddleware(cfg.InternalToken),
		), nil
	})

	router, err := do.Invoke[http.Handler](injector)
	if err != nil {
		return nil, fmt.Errorf("failed to build router: %w", err)
	}

	return router, nil
}
