package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// mockSlackClient is a test double for SlackClient.
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

func TestSlackPosterPostProgress(t *testing.T) {
	tests := []struct {
		name           string
		channelID      string
		teamID         string
		log            *entities.ProgressLog
		slackErr       error
		wantErr        bool
		wantErrContain string
		wantMsgParts   []string
	}{
		{
			name:      "successful post without SOS",
			channelID: "C12345",
			teamID:    "team-alpha",
			log: &entities.ProgressLog{
				ID:            "prog-1",
				ParticipantID: "U12345",
				ProgressBodies: []entities.ProgressBody{
					{
						Phase:   entities.ProgressPhaseCoding,
						SOS:     false,
						Comment: "Working on feature",
					},
				},
			},
			wantMsgParts: []string{
				"進捗報告",
				"チーム: team-alpha",
				"フェーズ: coding",
				"コメント: Working on feature",
			},
		},
		{
			name:      "successful post with SOS",
			channelID: "C12345",
			teamID:    "team-alpha",
			log: &entities.ProgressLog{
				ID:            "prog-2",
				ParticipantID: "U12345",
				ProgressBodies: []entities.ProgressBody{
					{
						Phase:   entities.ProgressPhaseTesting,
						SOS:     true,
						Comment: "Need help",
					},
				},
			},
			wantMsgParts: []string{":sos:"},
		},
		{
			name:      "slack client error",
			channelID: "C12345",
			teamID:    "team-alpha",
			log: &entities.ProgressLog{
				ID:            "prog-3",
				ParticipantID: "U12345",
				ProgressBodies: []entities.ProgressBody{
					{
						Phase:   entities.ProgressPhaseCoding,
						Comment: "test",
					},
				},
			},
			slackErr:       errors.New("slack api error"),
			wantErr:        true,
			wantErrContain: "failed to post message to slack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockSlackClient{err: tt.slackErr}
			formatter := &ProgressFormatter{}
			poster := &SlackPoster{client: client, formatter: formatter}

			err := poster.PostProgress(context.Background(), tt.channelID, tt.teamID, tt.log)

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

			if client.postedChannelID != tt.channelID {
				t.Fatalf("expected channel %q, got %q", tt.channelID, client.postedChannelID)
			}

			for _, part := range tt.wantMsgParts {
				if !strings.Contains(client.postedText, part) {
					t.Fatalf("message %q does not contain %q", client.postedText, part)
				}
			}
		})
	}
}

func TestProgressFormatterFormatSlackMessage(t *testing.T) {
	formatter := &ProgressFormatter{}

	t.Run("formats message without SOS", func(t *testing.T) {
		log := &entities.ProgressLog{
			ProgressBodies: []entities.ProgressBody{
				{Phase: entities.ProgressPhaseCoding, SOS: false, Comment: "Working"},
			},
		}

		msg := formatter.FormatSlackMessage("team-alpha", log)

		expectedParts := []string{"進捗報告", "チーム: team-alpha", "フェーズ: coding", "コメント: Working"}
		for _, part := range expectedParts {
			if !strings.Contains(msg, part) {
				t.Fatalf("message %q does not contain %q", msg, part)
			}
		}

		if strings.Contains(msg, ":sos:") {
			t.Fatal("message should not contain :sos: when SOS is false")
		}
	})

	t.Run("formats message with SOS", func(t *testing.T) {
		log := &entities.ProgressLog{
			ProgressBodies: []entities.ProgressBody{
				{Phase: entities.ProgressPhaseTesting, SOS: true, Comment: "Help"},
			},
		}

		msg := formatter.FormatSlackMessage("team-beta", log)

		if !strings.Contains(msg, ":sos:") {
			t.Fatal("message should contain :sos: when SOS is true")
		}
	})

	t.Run("handles empty progress bodies", func(t *testing.T) {
		log := &entities.ProgressLog{}

		msg := formatter.FormatSlackMessage("team-gamma", log)

		if !strings.Contains(msg, "進捗報告") {
			t.Fatalf("message %q does not contain 進捗報告", msg)
		}
	})
}
