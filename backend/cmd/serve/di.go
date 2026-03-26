package serve

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"

	_ "github.com/lib/pq"

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

// wireRouter builds all dependencies and returns the configured HTTP router.
func wireRouter(cfg *util.Config) (http.Handler, error) {
	// --- Infrastructure ---
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

	slackClient := slackinfra.NewClient(cfg.SlackBotToken)

	encryptor, err := encryption.NewAESEncryptor([]byte(cfg.EncryptionKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	ghClient := githubinfra.NewClient(cfg.GitHubAPIBaseURL)

	// --- Repositories ---
	progressRepo := postgres.NewProgressRepository(db)
	progressQueryRepo := postgres.NewProgressQueryRepository(db)
	questionRepo := postgres.NewQuestionRepository(db)
	userRepo := postgres.NewUserRepository(db)
	ghRepoRepo := postgres.NewGitHubRepoRepository(db)
	idempotencyStore := idempotency.NewPostgresStore(db)

	// --- Services ---
	formatter := service.NewProgressFormatter()
	poster := service.NewSlackPoster(slackClient, formatter)
	jwtService := service.NewJWTService(cfg.JWTSecret)
	hasher := service.NewPasswordHasher()
	idempotencySvc := idempotencysvc.NewService(idempotencyStore)
	slackVerifier := pkgslack.NewVerifier(cfg.SlackSigningSecret)
	ghService := service.NewGitHubService(ghRepoRepo, encryptor, ghClient, func() string { return uuid.New().String() })

	// --- Use Cases ---
	handleProgressUC := usecase.NewHandleProgressUseCase(progressRepo, poster)
	listProgressUC := usecase.NewListProgressUseCase(progressQueryRepo)
	loginUC := usecase.NewLoginUseCase(userRepo, jwtService, hasher)
	resolveQuestionUC := usecase.NewResolveQuestionUsecase(questionRepo)
	continueQuestionUC := usecase.NewContinueQuestionUsecase(questionRepo)
	// TODO: Wire a proper SlackNotifier implementation when available.
	escalateQuestionUC := usecase.NewEscalateQuestionUsecase(questionRepo, nil)

	// --- Handlers ---
	webhookHandler := rest.NewWebhookHandler(handleProgressUC)
	questionHandler := rest.NewQuestionHandler(
		usecase.NewHandleNewQuestionUseCase(questionRepo, service.NewQuestionSender(nil)),
	)
	progressHandler := rest.NewProgressHandler(listProgressUC)
	authHandler := rest.NewAuthHandler(loginUC)
	ghHandler := rest.NewGitHubHandler(ghService)
	internalHandler := rest.NewInternalHandler(ghService)
	// TODO: Wire SlackThreadFetcher and MessageQueue implementations when available.
	eventHandler := rest.NewEventHandler(
		usecase.NewTriggerIssueCreationUseCase(nil, nil),
	)
	interactionHandler := rest.NewInteractionHandler(resolveQuestionUC, continueQuestionUC, escalateQuestionUC)

	// --- Router ---
	router := rest.NewRouter(
		webhookHandler,
		questionHandler,
		progressHandler,
		authHandler,
		ghHandler,
		internalHandler,
		eventHandler,
		interactionHandler,
		middleware.AuthMiddleware(jwtService),
		middleware.SlackWebhookMiddleware(slackVerifier, idempotencySvc),
		middleware.InternalTokenMiddleware(cfg.InternalToken),
	)

	return router, nil
}
