package util

import (
	"github.com/kelseyhightower/envconfig"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Port               string `envconfig:"PORT" default:"8080"`
	SlackBotToken      string `envconfig:"SLACK_BOT_TOKEN" required:"true"`
	DatabaseHost       string `envconfig:"DATABASE_HOST" required:"true"`
	DatabasePort       string `envconfig:"DATABASE_PORT" default:"5432"`
	DatabaseName       string `envconfig:"DATABASE_NAME" required:"true"`
	DatabaseUser       string `envconfig:"DATABASE_USER" required:"true"`
	DatabasePass       string `envconfig:"DATABASE_PASS" required:"true"`
	SlackSigningSecret string `envconfig:"SLACK_SIGNING_SECRET" required:"true"`
}

// LoadConfig reads configuration from environment variables.
// It returns an error if any required environment variable is missing.
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
