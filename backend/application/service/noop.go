package service

import (
	"context"
	"log/slog"

	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

// NoopSlackNotifier is a no-op implementation of SlackNotifier that logs a warning.
type NoopSlackNotifier struct{}

// PostToMentorChannel logs a warning that the notifier is not configured.
func (n *NoopSlackNotifier) PostToMentorChannel(_ context.Context, question *entities.Question) error {
	slog.Warn("slack notifier not configured, skipping mentor notification",
		slog.String("question_id", string(question.ID)))
	return nil
}

// NoopSlackThreadFetcher is a no-op implementation of SlackThreadFetcher that logs a warning.
type NoopSlackThreadFetcher struct{}

// FetchThreadMessages logs a warning and returns an empty slice.
func (n *NoopSlackThreadFetcher) FetchThreadMessages(_ context.Context, channelID, threadTS string) ([]slack.ThreadMessage, error) {
	slog.Warn("slack thread fetcher not configured, returning empty messages",
		slog.String("channel_id", channelID),
		slog.String("thread_ts", threadTS))
	return nil, nil
}

// NoopMessageQueue is a no-op implementation of port.MessageQueue that logs a warning.
type NoopMessageQueue struct{}

// Send logs a warning that the message queue is not configured.
func (n *NoopMessageQueue) Send(_ context.Context, queueName string, _ []byte) error {
	slog.Warn("message queue not configured, message not sent",
		slog.String("queue_name", queueName))
	return nil
}
