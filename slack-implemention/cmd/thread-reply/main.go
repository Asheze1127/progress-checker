package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Asheze1127/progress-checker/slack-implemention/internal/config"
	"github.com/Asheze1127/progress-checker/slack-implemention/internal/service"
)

func main() {
	var channelID string
	var threadTS string
	var text string
	var delaySeconds int

	flag.StringVar(&channelID, "channel", "", "target channel id")
	flag.StringVar(&threadTS, "thread", "", "target thread timestamp")
	flag.StringVar(&text, "text", "", "reply message")
	flag.IntVar(&delaySeconds, "delay", 0, "delay seconds before posting")
	flag.Parse()

	if err := config.LoadDotEnv(".env", "slack-implemention/.env"); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("SLACK_BOT_TOKEN is required")
	}

	svc := service.NewSlackService(botToken)

	if delaySeconds > 0 {
		log.Printf("waiting %d seconds before replying...", delaySeconds)
		time.Sleep(time.Duration(delaySeconds) * time.Second)
	}

	if err := svc.PostThreadReply(channelID, threadTS, text); err != nil {
		log.Fatalf("failed to post thread reply: %v", err)
	}

	fmt.Printf("thread reply sent: channel=%s thread=%s\n", channelID, threadTS)
}
