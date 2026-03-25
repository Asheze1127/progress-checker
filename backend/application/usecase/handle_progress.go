package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

// HandleProgressInput represents the input for the handle progress use case.
// Fields come from Slack's structured modal submission payload.
type HandleProgressInput struct {
	SlackUserID string
	TeamID      string
	ChannelID   string
	Phase       entities.ProgressPhase
	SOS         bool
	Comment     string
}

// HandleProgressUseCase orchestrates the handling of a progress command.
type HandleProgressUseCase struct {
	repo   entities.ProgressRepository
	poster *service.SlackPoster
}

// NewHandleProgressUseCase creates a new HandleProgressUseCase with the given dependencies.
func NewHandleProgressUseCase(repo entities.ProgressRepository, poster *service.SlackPoster) *HandleProgressUseCase {
	return &HandleProgressUseCase{
		repo:   repo,
		poster: poster,
	}
}

// Execute runs the handle progress use case.
// It builds the entity, saves it, and posts to Slack.
func (uc *HandleProgressUseCase) Execute(ctx context.Context, input HandleProgressInput) error {
	if input.SlackUserID == "" {
		return fmt.Errorf("slack_user_id is required")
	}
	if input.ChannelID == "" {
		return fmt.Errorf("channel_id is required")
	}

	now := time.Now().UTC()

	progressLog := &entities.ProgressLog{
		ID:            entities.ProgressLogID(uuid.New().String()),
		ParticipantID: entities.ParticipantID(input.SlackUserID),
		ProgressBodies: []entities.ProgressBody{
			{
				Phase:       input.Phase,
				SOS:         input.SOS,
				Comment:     input.Comment,
				SubmittedAt: now,
			},
		},
	}

	if err := uc.repo.Save(ctx, progressLog); err != nil {
		return fmt.Errorf("failed to save progress log: %w", err)
	}

	if err := uc.poster.PostProgress(ctx, input.ChannelID, input.TeamID, progressLog); err != nil {
		return err
	}

	return nil
}
