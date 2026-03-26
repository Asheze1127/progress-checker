package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/application/port"
	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

const issueQueueName = "issue"

// TriggerIssueCreationInput contains the data needed to trigger issue creation.
type TriggerIssueCreationInput struct {
	ChannelID     string
	ThreadTS      string
	TriggerUserID string
	TriggerType   string // "reaction" or "message_action"
}

// Validate checks that all required fields are present.
func (i TriggerIssueCreationInput) Validate() error {
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
	Type           string               `json:"type"`
	ChannelID      string               `json:"channel_id"`
	ThreadTS       string               `json:"thread_ts"`
	TriggerUserID  string               `json:"trigger_user_id"`
	TriggerType    string               `json:"trigger_type"`
	ThreadMessages []slack.ThreadMessage `json:"thread_messages"`
}

// TriggerIssueCreationUseCase orchestrates fetching thread history and
// enqueuing an issue creation message.
type TriggerIssueCreationUseCase struct {
	threadFetcher service.SlackThreadFetcher
	queue         port.MessageQueue
}

// NewTriggerIssueCreationUseCase creates a new TriggerIssueCreationUseCase.
func NewTriggerIssueCreationUseCase(
	fetcher service.SlackThreadFetcher,
	queue port.MessageQueue,
) *TriggerIssueCreationUseCase {
	return &TriggerIssueCreationUseCase{
		threadFetcher: fetcher,
		queue:         queue,
	}
}

// Execute validates input, fetches thread history, and enqueues to the issue queue.
func (uc *TriggerIssueCreationUseCase) Execute(ctx context.Context, input TriggerIssueCreationInput) error {
	if err := input.Validate(); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	messages, err := uc.threadFetcher.FetchThreadMessages(ctx, input.ChannelID, input.ThreadTS)
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

	if err := uc.queue.Send(ctx, issueQueueName, body); err != nil {
		return fmt.Errorf("failed to publish to queue: %w", err)
	}

	return nil
}
