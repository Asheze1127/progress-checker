package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

// IssueTriggerInput contains the data needed to trigger issue creation.
type IssueTriggerInput struct {
	ChannelID     string
	ThreadTS      string
	TriggerUserID string
	TriggerType   string // "reaction" or "message_action"
}

// Validate checks that all required fields are present.
func (i IssueTriggerInput) Validate() error {
	var errs []error

	if strings.TrimSpace(i.ChannelID) == "" {
		errs = append(errs, fmt.Errorf("channel_id is required"))
	}
	if strings.TrimSpace(i.ThreadTS) == "" {
		errs = append(errs, fmt.Errorf("thread_ts is required"))
	}
	if strings.TrimSpace(i.TriggerUserID) == "" {
		errs = append(errs, fmt.Errorf("trigger_user_id is required"))
	}
	if i.TriggerType != "reaction" && i.TriggerType != "message_action" {
		errs = append(errs, fmt.Errorf("trigger_type must be 'reaction' or 'message_action'"))
	}

	return errors.Join(errs...)
}

// IssueQueueMessage is the message format sent to the SQS issue queue.
type IssueQueueMessage struct {
	Type           string                 `json:"type"`
	ChannelID      string                 `json:"channel_id"`
	ThreadTS       string                 `json:"thread_ts"`
	TriggerUserID  string                 `json:"trigger_user_id"`
	TriggerType    string                 `json:"trigger_type"`
	ThreadMessages []slack.ThreadMessage   `json:"thread_messages"`
}

// SlackThreadFetcher retrieves thread messages from Slack.
type SlackThreadFetcher interface {
	FetchThreadMessages(ctx context.Context, channelID, threadTS string) ([]slack.ThreadMessage, error)
}

// SQSPublisher publishes messages to an SQS queue.
type SQSPublisher interface {
	Publish(ctx context.Context, queueName string, message []byte) error
}

const issueQueueName = "issue"

// IssueTriggerService handles the business logic of triggering issue creation.
type IssueTriggerService struct {
	threadFetcher SlackThreadFetcher
	sqsPublisher  SQSPublisher
}

// NewIssueTriggerService creates a new IssueTriggerService.
func NewIssueTriggerService(fetcher SlackThreadFetcher, publisher SQSPublisher) *IssueTriggerService {
	return &IssueTriggerService{
		threadFetcher: fetcher,
		sqsPublisher:  publisher,
	}
}

// TriggerIssueCreation fetches thread history and enqueues to the SQS issue queue.
func (s *IssueTriggerService) TriggerIssueCreation(ctx context.Context, input IssueTriggerInput) error {
	if err := input.Validate(); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	messages, err := s.threadFetcher.FetchThreadMessages(ctx, input.ChannelID, input.ThreadTS)
	if err != nil {
		return fmt.Errorf("failed to fetch thread messages: %w", err)
	}

	queueMessage := IssueQueueMessage{
		Type:           "issue",
		ChannelID:      input.ChannelID,
		ThreadTS:       input.ThreadTS,
		TriggerUserID:  input.TriggerUserID,
		TriggerType:    input.TriggerType,
		ThreadMessages: messages,
	}

	body, err := json.Marshal(queueMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal queue message: %w", err)
	}

	if err := s.sqsPublisher.Publish(ctx, issueQueueName, body); err != nil {
		return fmt.Errorf("failed to publish to SQS: %w", err)
	}

	return nil
}
