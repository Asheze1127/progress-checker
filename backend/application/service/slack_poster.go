package service

import (
	"context"
	"fmt"

	"github.com/samber/do/v2"

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

// NewSlackPoster creates a new SlackPoster via DI container.
func NewSlackPoster(i do.Injector) (*SlackPoster, error) {
	client := do.MustInvoke[SlackClient](i)
	formatter := do.MustInvoke[*ProgressFormatter](i)
	return &SlackPoster{
		client:    client,
		formatter: formatter,
	}, nil
}

// PostProgress formats a progress log and posts it to the specified Slack channel.
func (s *SlackPoster) PostProgress(ctx context.Context, channelID string, teamID string, log *entities.ProgressLog) error {
	message := s.formatter.FormatSlackMessage(teamID, log)

	if err := s.client.PostMessage(ctx, channelID, message); err != nil {
		return fmt.Errorf("failed to post message to slack: %w", err)
	}

	return nil
}
