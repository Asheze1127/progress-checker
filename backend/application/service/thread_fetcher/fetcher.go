package threadfetcher

import (
	"context"
	"log/slog"

	"github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

type SlackThreadFetcher interface {
	FetchThreadMessages(ctx context.Context, channelID, threadTS string) ([]slack.ThreadMessage, error)
}

type NoopSlackThreadFetcher struct{}

func (n *NoopSlackThreadFetcher) FetchThreadMessages(_ context.Context, channelID, threadTS string) ([]slack.ThreadMessage, error) {
	slog.Warn("slack thread fetcher not configured, returning empty messages", slog.String("channel_id", channelID), slog.String("thread_ts", threadTS))
	return nil, nil
}
