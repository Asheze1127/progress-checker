package port

import "context"

// MessageQueue sends messages to a named queue.
type MessageQueue interface {
	Send(ctx context.Context, queueName string, message []byte) error
}
