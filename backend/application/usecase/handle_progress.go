package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// HandleProgressInput represents the input for the handle progress use case.
// Text contains the raw text from the Slack slash command.
type HandleProgressInput struct {
	SlackUserID string
	TeamID      string
	ChannelID   string
	Text        string
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
// It builds the entity, saves it (validation happens in the repository), and posts to Slack.
func (uc *HandleProgressUseCase) Execute(ctx context.Context, input HandleProgressInput) error {
	if input.SlackUserID == "" {
		return fmt.Errorf("slack_user_id is required")
	}
	if input.ChannelID == "" {
		return fmt.Errorf("channel_id is required")
	}

	now := time.Now().UTC()

	// Build entity from raw text input
	progressLog := &entities.ProgressLog{
		ID:            entities.ProgressLogID(uc.idGen.Generate()),
		ParticipantID: entities.ParticipantID(input.SlackUserID),
		ProgressBodies: []entities.ProgressBody{
			{
				Phase:       entities.ProgressPhase(input.Text),
				SOS:         false,
				Comment:     input.Text,
				SubmittedAt: now,
			},
		},
	}

	// Save to repository (validation is done in the repository layer)
	if err := uc.repo.Save(ctx, progressLog); err != nil {
		return fmt.Errorf("failed to save progress log: %w", err)
	}

	// Post to Slack via service
	if err := uc.poster.PostProgress(ctx, input.ChannelID, input.TeamID, progressLog); err != nil {
		return err
	}

	return nil
}
