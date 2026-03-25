package serve

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/api/rest"
	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/postgres"
	slackinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/slack"
	"github.com/Asheze1127/progress-checker/backend/util"
)

// wireRouter builds all dependencies and returns the configured HTTP router.
func wireRouter(cfg *util.Config) (*http.ServeMux, error) {
	// Wire infrastructure: Slack client
	slackClient := slackinfra.NewClient(cfg.SlackBotToken)

	// Wire infrastructure: database and repository
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseUser, cfg.DatabasePass, cfg.DatabaseName,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	repo := postgres.NewProgressRepository(db)

	// Wire application services
	formatter := service.NewProgressFormatter()
	poster := service.NewSlackPoster(slackClient, formatter)

	// Wire use case
	handleProgressUC := usecase.NewHandleProgressUseCase(repo, poster)

	// Wire HTTP handlers
	webhookHandler := rest.NewWebhookHandler(handleProgressUC)
	router := rest.NewRouter(webhookHandler)

	return router, nil
}
