package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// mockProgressRepository is a test double for entities.ProgressRepository.
type mockProgressRepository struct {
	savedLog *entities.ProgressLog
	err      error
}

func (m *mockProgressRepository) Save(_ context.Context, log *entities.ProgressLog) error {
	m.savedLog = log
	return m.err
}

// mockSlackClient is a test double for service.SlackClient.
type mockSlackClient struct {
	postedChannelID string
	postedText      string
	err             error
}

func (m *mockSlackClient) PostMessage(_ context.Context, channelID string, text string) error {
	m.postedChannelID = channelID
	m.postedText = text
	return m.err
}

// mockIDGenerator is a test double for service.IDGenerator.
type mockIDGenerator struct {
	id string
}

func (m *mockIDGenerator) Generate() string {
	return m.id
}

func newTestUseCase(repo entities.ProgressRepository, slackClient service.SlackClient, idGen service.IDGenerator) *HandleProgressUseCase {
	formatter := service.NewProgressFormatter()
	poster := service.NewSlackPoster(slackClient, formatter)
	return NewHandleProgressUseCase(repo, poster, idGen)
}

func TestHandleProgressUseCaseExecute(t *testing.T) {
	tests := []struct {
		name           string
		input          HandleProgressInput
		repoErr        error
		slackErr       error
		wantErr        bool
		wantErrContain string
		wantSlackMsg   string
		wantChannelID  string
	}{
		{
			name: "successful progress command",
			input: HandleProgressInput{
				SlackUserID: "U12345",
				TeamID:      "team-alpha",
				ChannelID:   "C12345",
				Phase:       "coding",
				SOS:         false,
				Comment:     "Working on feature X",
			},
			wantSlackMsg:  "進捗報告",
			wantChannelID: "C12345",
		},
		{
			name: "successful progress command with SOS",
			input: HandleProgressInput{
				SlackUserID: "U12345",
				TeamID:      "team-alpha",
				ChannelID:   "C12345",
				Phase:       "testing",
				SOS:         true,
				Comment:     "Need help with tests",
			},
			wantSlackMsg:  ":sos:",
			wantChannelID: "C12345",
		},
		{
			name: "missing slack_user_id",
			input: HandleProgressInput{
				ChannelID: "C12345",
				Phase:     "coding",
			},
			wantErr:        true,
			wantErrContain: "slack_user_id is required",
		},
		{
			name: "missing channel_id",
			input: HandleProgressInput{
				SlackUserID: "U12345",
				Phase:       "coding",
			},
			wantErr:        true,
			wantErrContain: "channel_id is required",
		},
		{
			name: "missing phase",
			input: HandleProgressInput{
				SlackUserID: "U12345",
				ChannelID:   "C12345",
			},
			wantErr:        true,
			wantErrContain: "phase is required",
		},
		{
			name: "invalid phase",
			input: HandleProgressInput{
				SlackUserID: "U12345",
				TeamID:      "team-alpha",
				ChannelID:   "C12345",
				Phase:       "invalid_phase",
			},
			wantErr:        true,
			wantErrContain: "validation failed",
		},
		{
			name: "repository save error",
			input: HandleProgressInput{
				SlackUserID: "U12345",
				TeamID:      "team-alpha",
				ChannelID:   "C12345",
				Phase:       "coding",
				Comment:     "progress update",
			},
			repoErr:        errors.New("database connection error"),
			wantErr:        true,
			wantErrContain: "failed to save progress log",
		},
		{
			name: "slack post error",
			input: HandleProgressInput{
				SlackUserID: "U12345",
				TeamID:      "team-alpha",
				ChannelID:   "C12345",
				Phase:       "coding",
				Comment:     "progress update",
			},
			slackErr:       errors.New("slack api error"),
			wantErr:        true,
			wantErrContain: "failed to post message to slack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProgressRepository{err: tt.repoErr}
			slackClient := &mockSlackClient{err: tt.slackErr}
			idGen := &mockIDGenerator{id: "test-progress-id"}

			uc := newTestUseCase(repo, slackClient, idGen)
			err := uc.Execute(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrContain != "" && !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErrContain)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify repository was called with correct data
			if repo.savedLog == nil {
				t.Fatal("expected progress log to be saved")
			}
			if repo.savedLog.ID != "test-progress-id" {
				t.Fatalf("expected ID %q, got %q", "test-progress-id", repo.savedLog.ID)
			}
			if repo.savedLog.ParticipantID != entities.ParticipantID(tt.input.SlackUserID) {
				t.Fatalf("expected ParticipantID %q, got %q", tt.input.SlackUserID, repo.savedLog.ParticipantID)
			}
			if len(repo.savedLog.ProgressBodies) != 1 {
				t.Fatalf("expected 1 progress body, got %d", len(repo.savedLog.ProgressBodies))
			}

			body := repo.savedLog.ProgressBodies[0]
			if string(body.Phase) != tt.input.Phase {
				t.Fatalf("expected phase %q, got %q", tt.input.Phase, body.Phase)
			}
			if body.SOS != tt.input.SOS {
				t.Fatalf("expected SOS %v, got %v", tt.input.SOS, body.SOS)
			}
			if body.Comment != tt.input.Comment {
				t.Fatalf("expected comment %q, got %q", tt.input.Comment, body.Comment)
			}

			// Verify Slack was called
			if slackClient.postedChannelID != tt.wantChannelID {
				t.Fatalf("expected channel %q, got %q", tt.wantChannelID, slackClient.postedChannelID)
			}
			if tt.wantSlackMsg != "" && !strings.Contains(slackClient.postedText, tt.wantSlackMsg) {
				t.Fatalf("slack message %q does not contain %q", slackClient.postedText, tt.wantSlackMsg)
			}
		})
	}
}

func TestHandleProgressUseCaseSlackMessageFormat(t *testing.T) {
	repo := &mockProgressRepository{}
	slackClient := &mockSlackClient{}
	idGen := &mockIDGenerator{id: "test-id"}

	uc := newTestUseCase(repo, slackClient, idGen)

	input := HandleProgressInput{
		SlackUserID: "U12345",
		TeamID:      "team-alpha",
		ChannelID:   "C12345",
		Phase:       "coding",
		SOS:         false,
		Comment:     "Implementing feature",
	}

	err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParts := []string{
		"進捗報告",
		"チーム: team-alpha",
		"フェーズ: coding",
		"コメント: Implementing feature",
	}

	for _, part := range expectedParts {
		if !strings.Contains(slackClient.postedText, part) {
			t.Fatalf("slack message %q does not contain %q", slackClient.postedText, part)
		}
	}

	if strings.Contains(slackClient.postedText, ":sos:") {
		t.Fatal("slack message should not contain :sos: when SOS is false")
	}
}

func TestHandleProgressUseCaseSOSFormat(t *testing.T) {
	repo := &mockProgressRepository{}
	slackClient := &mockSlackClient{}
	idGen := &mockIDGenerator{id: "test-id"}

	uc := newTestUseCase(repo, slackClient, idGen)

	input := HandleProgressInput{
		SlackUserID: "U12345",
		TeamID:      "team-alpha",
		ChannelID:   "C12345",
		Phase:       "testing",
		SOS:         true,
		Comment:     "Need help",
	}

	err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(slackClient.postedText, ":sos:") {
		t.Fatal("slack message should contain :sos: when SOS is true")
	}
}
