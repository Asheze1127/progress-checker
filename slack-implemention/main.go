package main

import (
	"log"
	"net/http"

	"github.com/Asheze1127/progress-checker/slack-implemention/internal/config"
	"github.com/Asheze1127/progress-checker/slack-implemention/internal/service"
	"github.com/Asheze1127/progress-checker/slack-implemention/internal/slackapp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	slackService := service.NewSlackService(cfg.BotToken)
	commandHandler := slackapp.NewCommandHandler(cfg.SigningSecret, slackService)
	interactionHandler := slackapp.NewInteractionHandler(cfg.SigningSecret, slackService)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/slack/commands", commandHandler.Handle)
	mux.HandleFunc("/slack/interactions", interactionHandler.Handle)

	log.Printf("slack server listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
