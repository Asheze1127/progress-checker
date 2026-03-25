package port

import "context"

// MessageQueue publishes messages to a named queue.
type MessageQueue interface {
	Publish(ctx context.Context, queueName string, message []byte) error
}
