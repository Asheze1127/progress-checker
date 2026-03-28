package sqs

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	messagequeue "github.com/Asheze1127/progress-checker/backend/application/service/message_queue"
)

// Compile-time check that Client implements messagequeue.MessageQueue.
var _ messagequeue.MessageQueue = (*Client)(nil)

// Client implements the port.MessageQueue interface using AWS SQS.
type Client struct {
	api      *sqs.Client
	mu       sync.RWMutex
	urlCache map[string]string
}

// NewClient creates a new SQS Client with the given AWS SQS client.
func NewClient(api *sqs.Client) *Client {
	return &Client{api: api, urlCache: make(map[string]string)}
}

// Send sends a message to the specified SQS queue by name.
func (c *Client) Send(ctx context.Context, queueName string, message []byte) error {
	queueURL, err := c.resolveQueueURL(ctx, queueName)
	if err != nil {
		return err
	}

	_, err = c.api.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(message)),
	})
	if err != nil {
		return fmt.Errorf("failed to send message to queue %q: %w", queueName, err)
	}

	return nil
}

func (c *Client) resolveQueueURL(ctx context.Context, queueName string) (string, error) {
	c.mu.RLock()
	if url, ok := c.urlCache[queueName]; ok {
		c.mu.RUnlock()
		return url, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if url, ok := c.urlCache[queueName]; ok {
		return url, nil
	}

	urlOutput, err := c.api.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get queue URL for %q: %w", queueName, err)
	}
	if urlOutput.QueueUrl == nil {
		return "", fmt.Errorf("AWS returned nil queue URL for %q", queueName)
	}

	c.urlCache[queueName] = *urlOutput.QueueUrl
	return *urlOutput.QueueUrl, nil
}
