package serve

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Asheze1127/progress-checker/backend/api/rest"
	"github.com/Asheze1127/progress-checker/backend/application"
	slackinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/slack"
)

const defaultPort = "8080"

// Config holds the server configuration loaded from environment variables.
type Config struct {
	Port          string
	SlackBotToken string
	DatabaseHost  string
	DatabasePort  string
	DatabaseName  string
	DatabaseUser  string
	DatabasePass  string
}

// LoadConfigFromEnv reads configuration from environment variables.
func LoadConfigFromEnv() Config {
	return Config{
		Port:          getEnvOrDefault("PORT", defaultPort),
		SlackBotToken: os.Getenv("SLACK_BOT_TOKEN"),
		DatabaseHost:  os.Getenv("DATABASE_HOST"),
		DatabasePort:  getEnvOrDefault("DATABASE_PORT", "5432"),
		DatabaseName:  os.Getenv("DATABASE_NAME"),
		DatabaseUser:  os.Getenv("DATABASE_USER"),
		DatabasePass:  os.Getenv("DATABASE_PASS"),
	}
}

// Run starts the HTTP server with all dependencies wired.
func Run() error {
	cfg := LoadConfigFromEnv()

	// Wire infrastructure dependencies
	slackClient := slackinfra.NewClient(cfg.SlackBotToken, nil)

	// For now, use a no-op repository until database infrastructure is implemented
	repo := &noopProgressRepository{}

	idGen := &uuidGenerator{}

	// Wire application services
	progressService := application.NewProgressService(repo, slackClient, idGen)

	// Wire HTTP handlers
	webhookHandler := rest.NewWebhookHandler(progressService)
	router := rest.NewRouter(webhookHandler)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s", addr)

	return http.ListenAndServe(addr, router)
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
