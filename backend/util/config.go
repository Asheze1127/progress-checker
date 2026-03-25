package util

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Port               string
	SlackBotToken      string
	DatabaseHost       string
	DatabasePort       string
	DatabaseName       string
	DatabaseUser       string
	DatabasePass       string
	SlackSigningSecret string
}

// LoadConfig reads configuration from environment variables.
// It returns an error if any required environment variable is missing.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:         getEnvOrDefault("PORT", "8080"),
		DatabasePort: getEnvOrDefault("DATABASE_PORT", "5432"),
	}

	var missing []string
	if cfg.SlackBotToken = os.Getenv("SLACK_BOT_TOKEN"); cfg.SlackBotToken == "" {
		missing = append(missing, "SLACK_BOT_TOKEN")
	}
	if cfg.DatabaseHost = os.Getenv("DATABASE_HOST"); cfg.DatabaseHost == "" {
		missing = append(missing, "DATABASE_HOST")
	}
	if cfg.DatabaseName = os.Getenv("DATABASE_NAME"); cfg.DatabaseName == "" {
		missing = append(missing, "DATABASE_NAME")
	}
	if cfg.DatabaseUser = os.Getenv("DATABASE_USER"); cfg.DatabaseUser == "" {
		missing = append(missing, "DATABASE_USER")
	}
	if cfg.DatabasePass = os.Getenv("DATABASE_PASS"); cfg.DatabasePass == "" {
		missing = append(missing, "DATABASE_PASS")
	}
	if cfg.SlackSigningSecret = os.Getenv("SLACK_SIGNING_SECRET"); cfg.SlackSigningSecret == "" {
		missing = append(missing, "SLACK_SIGNING_SECRET")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
