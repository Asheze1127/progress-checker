package messagequeue

import "context"

// MessageQueue is the interface for sending messages to a queue.
type MessageQueue interface {
  Send(ctx context.Context, queueName string, message []byte) error
}
