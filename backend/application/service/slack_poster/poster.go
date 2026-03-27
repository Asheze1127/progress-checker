package slackposter

import (
	"context"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/application/service/progress_formatter"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

type SlackClient interface {
	PostMessage(ctx context.Context, channelID string, text string) error
}

type SlackPoster struct {
	client    SlackClient
	formatter *progressformatter.ProgressFormatter
}

func NewSlackPoster(client SlackClient, formatter *progressformatter.ProgressFormatter) *SlackPoster {
	return &SlackPoster{client: client, formatter: formatter}
}

func (s *SlackPoster) PostProgress(ctx context.Context, channelID string, teamID string, log *entities.ProgressLog) error {
	message := s.formatter.FormatSlackMessage(teamID, log)
	if err := s.client.PostMessage(ctx, channelID, message); err != nil {
		return fmt.Errorf("failed to post message to slack: %w", err)
	}
	return nil
}
