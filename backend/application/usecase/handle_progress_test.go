package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/application/service/progress_formatter"
	"github.com/Asheze1127/progress-checker/backend/application/service/slack_poster"
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

// mockSlackClient is a test double for slackposter.SlackClient.
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

// mockUserRepoForProgress is a test double for entities.UserRepository (progress tests).
type mockUserRepoForProgress struct {
	user *entities.User
	err  error
}

func (m *mockUserRepoForProgress) GetByID(_ context.Context, _ entities.UserID) (*entities.User, error) {
	return m.user, m.err
}
func (m *mockUserRepoForProgress) GetByEmail(_ context.Context, _ string) (*entities.User, error) {
	return m.user, m.err
}
func (m *mockUserRepoForProgress) GetBySlackUserID(_ context.Context, _ entities.SlackUserID) (*entities.User, error) {
	return m.user, m.err
}
func (m *mockUserRepoForProgress) FindByEmail(_ context.Context, _ string) (*entities.UserWithPassword, error) {
	return nil, m.err
}
func (m *mockUserRepoForProgress) Create(_ context.Context, _ *entities.User, _ string) (*entities.User, error) {
	return m.user, m.err
}
func (m *mockUserRepoForProgress) UpdatePasswordHash(_ context.Context, _ entities.UserID, _ string) error {
	return m.err
}
func (m *mockUserRepoForProgress) List(_ context.Context) ([]*entities.User, error) {
	return nil, m.err
}

func newTestUseCase(repo entities.ProgressRepository, userRepo entities.UserRepository, slackClient slackposter.SlackClient) *HandleProgressUseCase {
	formatter := progressformatter.NewProgressFormatter()
	poster := slackposter.NewSlackPoster(slackClient, formatter)
	return NewHandleProgressUseCase(repo, userRepo, poster)
}

func defaultMockUserRepo() *mockUserRepoForProgress {
	return &mockUserRepoForProgress{
		user: &entities.User{
			ID:          "user-uuid-123",
			SlackUserID: "U12345",
			Name:        "Test User",
			Email:       "test@example.com",
			Role:        entities.UserRoleParticipant,
		},
	}
}

func TestHandleProgressUseCaseExecute(t *testing.T) {
	tests := []struct {
		name           string
		input          HandleProgressInput
		repoErr        error
		userRepoErr    error
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
				Phase:       entities.ProgressPhaseCoding,
				SOS:         false,
				Comment:     "Working on feature X",
			},
			wantSlackMsg:  "進捗報告",
			wantChannelID: "C12345",
		},
		{
			name: "missing slack_user_id",
			input: HandleProgressInput{
				ChannelID: "C12345",
				Phase:     entities.ProgressPhaseCoding,
				Comment:   "some text",
			},
			wantErr:        true,
			wantErrContain: "slack_user_id is required",
		},
		{
			name: "missing channel_id",
			input: HandleProgressInput{
				SlackUserID: "U12345",
				Phase:       entities.ProgressPhaseCoding,
				Comment:     "some text",
			},
			wantErr:        true,
			wantErrContain: "channel_id is required",
		},
		{
			name: "invalid phase",
			input: HandleProgressInput{
				SlackUserID: "U12345",
				ChannelID:   "C12345",
				Phase:       entities.ProgressPhase("invalid"),
				Comment:     "some text",
			},
			wantErr:        true,
			wantErrContain: "invalid phase",
		},
		{
			name: "user not found",
			input: HandleProgressInput{
				SlackUserID: "U99999",
				TeamID:      "team-alpha",
				ChannelID:   "C12345",
				Phase:       entities.ProgressPhaseCoding,
				Comment:     "progress update",
			},
			userRepoErr:    errors.New("user not found"),
			wantErr:        true,
			wantErrContain: "failed to find user",
		},
		{
			name: "repository save error",
			input: HandleProgressInput{
				SlackUserID: "U12345",
				TeamID:      "team-alpha",
				ChannelID:   "C12345",
				Phase:       entities.ProgressPhaseCoding,
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
				Phase:       entities.ProgressPhaseCoding,
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
			userRepo := defaultMockUserRepo()
			if tt.userRepoErr != nil {
				userRepo.err = tt.userRepoErr
				userRepo.user = nil
			}

			uc := newTestUseCase(repo, userRepo, slackClient)
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
			if repo.savedLog.ID == "" {
				t.Fatal("expected progress log to have a generated ID")
			}
			if repo.savedLog.ParticipantID != entities.ParticipantID("user-uuid-123") {
				t.Fatalf("expected ParticipantID %q, got %q", "user-uuid-123", repo.savedLog.ParticipantID)
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
	userRepo := defaultMockUserRepo()

	uc := newTestUseCase(repo, userRepo, slackClient)

	input := HandleProgressInput{
		SlackUserID: "U12345",
		TeamID:      "team-alpha",
		ChannelID:   "C12345",
		Phase:       entities.ProgressPhaseCoding,
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
}
