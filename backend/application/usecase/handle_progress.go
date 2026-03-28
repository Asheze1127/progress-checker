package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Asheze1127/progress-checker/backend/application/service/slack_poster"
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
	repo     entities.ProgressRepository
	userRepo entities.UserRepository
	poster   *slackposter.SlackPoster
}

// NewHandleProgressUseCase creates a new HandleProgressUseCase with the given dependencies.
func NewHandleProgressUseCase(repo entities.ProgressRepository, userRepo entities.UserRepository, poster *slackposter.SlackPoster) *HandleProgressUseCase {
	return &HandleProgressUseCase{
		repo:     repo,
		userRepo: userRepo,
		poster:   poster,
	}
}

// Execute runs the handle progress use case.
// It builds the entity, saves it, and posts to Slack.
func (uc *HandleProgressUseCase) Execute(ctx context.Context, input HandleProgressInput) (err error) {
	defer func() {
		attrs := []slog.Attr{slog.String("slack_user_id", input.SlackUserID), slog.String("team_id", input.TeamID), slog.String("phase", string(input.Phase))}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
		slog.LogAttrs(ctx, slog.LevelDebug, "HandleProgressUseCase.Execute", attrs...)
	}()

	if input.SlackUserID == "" {
		return fmt.Errorf("slack_user_id is required")
	}
	if input.ChannelID == "" {
		return fmt.Errorf("channel_id is required")
	}

	if !input.Phase.IsValid() {
		return fmt.Errorf("invalid phase: %s", input.Phase)
	}

	// Resolve Slack User ID to database User ID
	user, err := uc.userRepo.GetBySlackUserID(ctx, entities.SlackUserID(input.SlackUserID))
	if err != nil {
		return fmt.Errorf("failed to find user for slack ID %s: %w", input.SlackUserID, err)
	}

	now := time.Now().UTC()

	progressLog := &entities.ProgressLog{
		ID:            entities.ProgressLogID(uuid.New().String()),
		ParticipantID: entities.ParticipantID(user.ID),
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
