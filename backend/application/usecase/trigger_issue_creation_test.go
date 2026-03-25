package usecase

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

// --- Test doubles ---

type stubThreadFetcher struct {
	messages []slack.ThreadMessage
	err      error
}

func (s *stubThreadFetcher) FetchThreadMessages(_ context.Context, _, _ string) ([]slack.ThreadMessage, error) {
	return s.messages, s.err
}

type spyMessageQueue struct {
	calledQueue string
	calledBody  []byte
	err         error
}

func (s *spyMessageQueue) Publish(_ context.Context, queueName string, message []byte) error {
	s.calledQueue = queueName
	s.calledBody = message
	return s.err
}

// --- Tests ---

func TestTriggerIssueCreationInput_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   TriggerIssueCreationInput
		wantErr bool
	}{
		{
			name: "valid reaction input",
			input: TriggerIssueCreationInput{
				ChannelID:     "C123",
				ThreadTS:      "1234567890.123456",
				TriggerUserID: "U123",
				TriggerType:   "reaction",
			},
			wantErr: false,
		},
		{
			name: "valid message_action input",
			input: TriggerIssueCreationInput{
				ChannelID:     "C123",
				ThreadTS:      "1234567890.123456",
				TriggerUserID: "U123",
				TriggerType:   "message_action",
			},
			wantErr: false,
		},
		{
			name: "missing channel_id",
			input: TriggerIssueCreationInput{
				ThreadTS:      "1234567890.123456",
				TriggerUserID: "U123",
				TriggerType:   "reaction",
			},
			wantErr: true,
		},
		{
			name: "missing thread_ts",
			input: TriggerIssueCreationInput{
				ChannelID:     "C123",
				TriggerUserID: "U123",
				TriggerType:   "reaction",
			},
			wantErr: true,
		},
		{
			name: "missing trigger_user_id",
			input: TriggerIssueCreationInput{
				ChannelID:   "C123",
				ThreadTS:    "1234567890.123456",
				TriggerType: "reaction",
			},
			wantErr: true,
		},
		{
			name: "invalid trigger_type",
			input: TriggerIssueCreationInput{
				ChannelID:     "C123",
				ThreadTS:      "1234567890.123456",
				TriggerUserID: "U123",
				TriggerType:   "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecute_Success(t *testing.T) {
	t.Parallel()

	threadMessages := []slack.ThreadMessage{
		{User: "U001", Text: "Let's create an issue for this", TS: "1234567890.000001"},
		{User: "U002", Text: "Agreed, this needs tracking", TS: "1234567890.000002"},
	}

	fetcher := &stubThreadFetcher{messages: threadMessages}
	queue := &spyMessageQueue{}
	uc := NewTriggerIssueCreationUseCase(fetcher, queue)

	input := TriggerIssueCreationInput{
		ChannelID:     "C123",
		ThreadTS:      "1234567890.123456",
		TriggerUserID: "U123",
		TriggerType:   "reaction",
	}

	err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if queue.calledQueue != "issue" {
		t.Errorf("expected queue 'issue', got %q", queue.calledQueue)
	}

	var msg IssueQueueMessage
	if err := json.Unmarshal(queue.calledBody, &msg); err != nil {
		t.Fatalf("failed to unmarshal published message: %v", err)
	}

	if msg.Type != "issue" {
		t.Errorf("expected type 'issue', got %q", msg.Type)
	}
	if msg.ChannelID != "C123" {
		t.Errorf("expected channel_id 'C123', got %q", msg.ChannelID)
	}
	if msg.ThreadTS != "1234567890.123456" {
		t.Errorf("expected thread_ts '1234567890.123456', got %q", msg.ThreadTS)
	}
	if msg.TriggerUserID != "U123" {
		t.Errorf("expected trigger_user_id 'U123', got %q", msg.TriggerUserID)
	}
	if msg.TriggerType != "reaction" {
		t.Errorf("expected trigger_type 'reaction', got %q", msg.TriggerType)
	}
	if len(msg.ThreadMessages) != 2 {
		t.Errorf("expected 2 thread messages, got %d", len(msg.ThreadMessages))
	}
}

func TestExecute_InvalidInput(t *testing.T) {
	t.Parallel()

	fetcher := &stubThreadFetcher{}
	queue := &spyMessageQueue{}
	uc := NewTriggerIssueCreationUseCase(fetcher, queue)

	input := TriggerIssueCreationInput{
		ChannelID: "",
		ThreadTS:  "1234567890.123456",
	}

	err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestExecute_FetchError(t *testing.T) {
	t.Parallel()

	fetcher := &stubThreadFetcher{err: context.DeadlineExceeded}
	queue := &spyMessageQueue{}
	uc := NewTriggerIssueCreationUseCase(fetcher, queue)

	input := TriggerIssueCreationInput{
		ChannelID:     "C123",
		ThreadTS:      "1234567890.123456",
		TriggerUserID: "U123",
		TriggerType:   "reaction",
	}

	err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error when thread fetcher fails")
	}
}

func TestExecute_PublishError(t *testing.T) {
	t.Parallel()

	fetcher := &stubThreadFetcher{messages: []slack.ThreadMessage{}}
	queue := &spyMessageQueue{err: context.DeadlineExceeded}
	uc := NewTriggerIssueCreationUseCase(fetcher, queue)

	input := TriggerIssueCreationInput{
		ChannelID:     "C123",
		ThreadTS:      "1234567890.123456",
		TriggerUserID: "U123",
		TriggerType:   "reaction",
	}

	err := uc.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error when queue publish fails")
	}
}
