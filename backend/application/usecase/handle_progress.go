package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// HandleProgressInput represents the input for the handle progress use case.
type HandleProgressInput struct {
	SlackUserID string
	TeamID      string
	ChannelID   string
	Phase       string
	SOS         bool
	Comment     string
}

// HandleProgressUseCase orchestrates the handling of a progress command.
type HandleProgressUseCase struct {
	repo   entities.ProgressRepository
	poster *service.SlackPoster
	idGen  service.IDGenerator
}

// NewHandleProgressUseCase creates a new HandleProgressUseCase with the given dependencies.
func NewHandleProgressUseCase(repo entities.ProgressRepository, poster *service.SlackPoster, idGen service.IDGenerator) *HandleProgressUseCase {
	return &HandleProgressUseCase{
		repo:   repo,
		poster: poster,
		idGen:  idGen,
	}
}

// Execute runs the handle progress use case.
// It builds the entity, validates it, saves it, and posts to Slack.
func (uc *HandleProgressUseCase) Execute(ctx context.Context, input HandleProgressInput) error {
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

	// 1. Build entity
	progressLog := &entities.ProgressLog{
		ID:            entities.ProgressLogID(uc.idGen.Generate()),
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

	// 2. Validate entity
	if err := progressLog.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 3. Save to repository
	if err := uc.repo.Save(ctx, progressLog); err != nil {
		return fmt.Errorf("failed to save progress log: %w", err)
	}

	// 4. Post to Slack via service
	if err := uc.poster.PostProgress(ctx, input.ChannelID, input.TeamID, progressLog); err != nil {
		return err
	}

	return nil
}
