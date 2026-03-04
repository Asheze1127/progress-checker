package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	SigningSecret string
	BotToken      string
	Port          string
}

func Load() (Config, error) {
	_ = LoadDotEnv(
		".env",
		"slack-implemention/.env",
	)

	cfg := Config{
		SigningSecret: os.Getenv("SLACK_SIGNING_SECRET"),
		BotToken:      os.Getenv("SLACK_BOT_TOKEN"),
		Port:          os.Getenv("PORT"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.SigningSecret == "" {
		return Config{}, errors.New("SLACK_SIGNING_SECRET is required")
	}
	if cfg.BotToken == "" {
		return Config{}, errors.New("SLACK_BOT_TOKEN is required")
	}

	return cfg, nil
}

func (c Config) String() string {
	return fmt.Sprintf("port=%s", c.Port)
}
