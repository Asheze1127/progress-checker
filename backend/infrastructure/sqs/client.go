package sqs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Client implements the port.MessageQueue interface using AWS SQS.
type Client struct {
	api *sqs.Client
}

// NewClient creates a new SQS Client with the given AWS SQS client.
func NewClient(api *sqs.Client) *Client {
	return &Client{api: api}
}

// Send sends a message to the specified SQS queue by name.
func (c *Client) Send(ctx context.Context, queueName string, message []byte) error {
	urlOutput, err := c.api.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return fmt.Errorf("failed to get queue URL for %q: %w", queueName, err)
	}

	_, err = c.api.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    urlOutput.QueueUrl,
		MessageBody: aws.String(string(message)),
	})
	if err != nil {
		return fmt.Errorf("failed to send message to queue %q: %w", queueName, err)
	}

	return nil
}
