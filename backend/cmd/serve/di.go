package serve

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	_ "github.com/lib/pq"

	"github.com/Asheze1127/progress-checker/backend/api/middleware"
	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/api/rest"
	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/encryption"
	githubinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/github"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/postgres"
	slackinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/slack"
	pkgslack "github.com/Asheze1127/progress-checker/backend/pkg/slack"
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

	// --- Services ---
	formatter := service.NewProgressFormatter()
	poster := service.NewSlackPoster(slackClient, formatter)
	jwtService := service.NewJWTService(cfg.JWTSecret)
	hasher := service.NewPasswordHasher()
	slackVerifier := pkgslack.NewVerifier(cfg.SlackSigningSecret)
	ghService := service.NewGitHubService(ghRepoRepo, encryptor, ghClient, func() string { return uuid.New().String() })

	// --- Use Cases ---
	handleProgressUC := usecase.NewHandleProgressUseCase(progressRepo, poster)
	listProgressUC := usecase.NewListProgressUseCase(progressQueryRepo)
	loginUC := usecase.NewLoginUseCase(userRepo, jwtService, hasher)
	resolveQuestionUC := usecase.NewResolveQuestionUseCase(questionRepo)
	continueQuestionUC := usecase.NewContinueQuestionUseCase(questionRepo)
	// TODO: Wire a proper SlackNotifier implementation when available.
	escalateQuestionUC := usecase.NewEscalateQuestionUseCase(questionRepo, &service.NoopSlackNotifier{})

	// --- Handlers ---
	var corsOrigins []string
	if cfg.CORSAllowedOrigin != "" {
		corsOrigins = strings.Split(cfg.CORSAllowedOrigin, ",")
	}
	apiHandler := openapi.NewAPIHandler(loginUC, listProgressUC, ghService, corsOrigins)

	webhookHandler := rest.NewWebhookHandler(handleProgressUC)
	questionHandler := rest.NewQuestionHandler(
		usecase.NewHandleNewQuestionUseCase(questionRepo, service.NewQuestionSender(&service.NoopMessageQueue{})),
	)
	// TODO: Wire real SlackThreadFetcher and MessageQueue implementations when available.
	eventHandler := rest.NewEventHandler(
		usecase.NewTriggerIssueCreationUseCase(&service.NoopSlackThreadFetcher{}, &service.NoopMessageQueue{}),
		cfg.IssueTriggerEmoji,
	)
	interactionHandler := rest.NewInteractionHandler(resolveQuestionUC, continueQuestionUC, escalateQuestionUC)

	// --- Router ---
	router := rest.NewRouter(
		apiHandler,
		webhookHandler,
		questionHandler,
		eventHandler,
		interactionHandler,
		middleware.AuthMiddleware(jwtService),
		middleware.SlackWebhookMiddleware(slackVerifier),
		middleware.InternalTokenMiddleware(cfg.InternalToken),
	)

	return router, nil
}
