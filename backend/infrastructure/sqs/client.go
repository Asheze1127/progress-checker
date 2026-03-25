package sqs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSAPI defines the subset of the SQS client interface we use.
// This enables testing without depending on the full AWS SDK client.
type SQSAPI interface {
	GetQueueUrl(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

// Client implements the application.MessageQueue interface using AWS SQS.
type Client struct {
	api SQSAPI
}

// NewClient creates a new SQS Client with the given AWS SQS API implementation.
func NewClient(api SQSAPI) *Client {
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
