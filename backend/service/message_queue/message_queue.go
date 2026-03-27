package messagequeue

import (
	"context"
	"log/slog"
)

type MessageQueue interface {
	Send(ctx context.Context, queueName string, message []byte) error
}

type NoopMessageQueue struct{}

func (n *NoopMessageQueue) Send(_ context.Context, queueName string, _ []byte) error {
	slog.Warn("message queue not configured, message not sent", slog.String("queue_name", queueName))
	return nil
}
