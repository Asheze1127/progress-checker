package util

import "os"

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
func LoadConfig() *Config {
	return &Config{
		Port:               getEnvOrDefault("PORT", "8080"),
		SlackBotToken:      os.Getenv("SLACK_BOT_TOKEN"),
		DatabaseHost:       os.Getenv("DATABASE_HOST"),
		DatabasePort:       getEnvOrDefault("DATABASE_PORT", "5432"),
		DatabaseName:       os.Getenv("DATABASE_NAME"),
		DatabaseUser:       os.Getenv("DATABASE_USER"),
		DatabasePass:       os.Getenv("DATABASE_PASS"),
		SlackSigningSecret: os.Getenv("SLACK_SIGNING_SECRET"),
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
