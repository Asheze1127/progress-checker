package application

import (
	"context"
	"fmt"
	"time"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// ProgressRepository defines the interface for persisting progress logs.
type ProgressRepository interface {
	Save(ctx context.Context, log *entities.ProgressLog) error
}

// SlackClient defines the interface for posting messages to Slack.
type SlackClient interface {
	PostMessage(ctx context.Context, channelID string, text string) error
}

// IDGenerator defines the interface for generating unique IDs.
type IDGenerator interface {
	Generate() string
}

// ProgressCommandInput represents the input from a /progress slash command.
type ProgressCommandInput struct {
	SlackUserID string
	TeamID      string
	ChannelID   string
	Phase       string
	SOS         bool
	Comment     string
}

// ProgressService handles the business logic for /progress commands.
type ProgressService struct {
	repository  ProgressRepository
	slackClient SlackClient
	idGenerator IDGenerator
}

// NewProgressService creates a new ProgressService with the given dependencies.
func NewProgressService(repo ProgressRepository, slack SlackClient, idGen IDGenerator) *ProgressService {
	return &ProgressService{
		repository:  repo,
		slackClient: slack,
		idGenerator: idGen,
	}
}

// HandleProgressCommand validates, saves, and posts a progress report.
func (s *ProgressService) HandleProgressCommand(ctx context.Context, input ProgressCommandInput) error {
	if input.SlackUserID == "" {
		return fmt.Errorf("slack_user_id is required")
	}
	if input.ChannelID == "" {
		return fmt.Errorf("channel_id is required")
	}
	if input.Phase == "" {
		return fmt.Errorf("phase is required")
	}

	now := time.Now().UTC()

	progressLog := &entities.ProgressLog{
		ID:            entities.ProgressLogID(s.idGenerator.Generate()),
		ParticipantID: entities.ParticipantID(input.SlackUserID),
		ProgressBodies: []entities.ProgressBody{
			{
				Phase:       entities.ProgressPhase(input.Phase),
				SOS:         input.SOS,
				Comment:     input.Comment,
				SubmittedAt: now,
			},
		},
	}

	if err := progressLog.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.repository.Save(ctx, progressLog); err != nil {
		return fmt.Errorf("failed to save progress log: %w", err)
	}

	sosEmoji := ""
	if input.SOS {
		sosEmoji = " :sos:"
	}

	message := fmt.Sprintf(
		":bar_chart: 進捗報告\nチーム: %s\nフェーズ: %s%s\nコメント: %s",
		input.TeamID,
		input.Phase,
		sosEmoji,
		input.Comment,
	)

	if err := s.slackClient.PostMessage(ctx, input.ChannelID, message); err != nil {
		return fmt.Errorf("failed to post message to slack: %w", err)
	}

	return nil
}
