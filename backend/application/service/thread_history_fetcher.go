package service

import (
	"context"

	"github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

// SlackThreadFetcher retrieves thread messages from Slack.
type SlackThreadFetcher interface {
	FetchThreadMessages(ctx context.Context, channelID, threadTS string) ([]slack.ThreadMessage, error)
}
