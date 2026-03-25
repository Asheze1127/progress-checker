package service

import (
	"context"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// SlackClient defines the interface for posting messages to Slack.
type SlackClient interface {
	PostMessage(ctx context.Context, channelID string, text string) error
}

// SlackPoster posts formatted progress messages to Slack.
type SlackPoster struct {
	client    SlackClient
	formatter *ProgressFormatter
}

// NewSlackPoster creates a new SlackPoster with the given Slack client and formatter.
func NewSlackPoster(client SlackClient, formatter *ProgressFormatter) *SlackPoster {
	return &SlackPoster{
		client:    client,
		formatter: formatter,
	}
}

// PostProgress formats a progress log and posts it to the specified Slack channel.
func (s *SlackPoster) PostProgress(ctx context.Context, channelID string, teamID string, log *entities.ProgressLog) error {
	message := s.formatter.FormatSlackMessage(teamID, log)

	if err := s.client.PostMessage(ctx, channelID, message); err != nil {
		return fmt.Errorf("failed to post message to slack: %w", err)
	}

	return nil
}
