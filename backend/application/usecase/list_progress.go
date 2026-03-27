package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Asheze1127/progress-checker/backend/api/middleware"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// ListProgressUseCase orchestrates fetching the latest progress data,
// optionally filtered by team ID.
type ListProgressUseCase struct {
	repo       entities.ProgressQueryRepository
	mentorRepo entities.MentorRepository
}

// NewListProgressUseCase creates a new ListProgressUseCase.
func NewListProgressUseCase(repo entities.ProgressQueryRepository, mentorRepo entities.MentorRepository) *ListProgressUseCase {
	return &ListProgressUseCase{repo: repo, mentorRepo: mentorRepo}
}

// Execute returns the latest progress for the mentor's assigned teams,
// or for a specific team when teamID is non-empty.
func (uc *ListProgressUseCase) Execute(ctx context.Context, teamID string) (result []entities.TeamProgress, err error) {
	defer func() {
		attrs := []slog.Attr{slog.String("team_id", teamID)}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		} else {
			attrs = append(attrs, slog.Int("count", len(result)))
		}
		slog.LogAttrs(ctx, slog.LevelDebug, "ListProgressUseCase.Execute", attrs...)
	}()

	mentorUser := middleware.UserFromContext(ctx)
	if mentorUser == nil {
		return nil, fmt.Errorf("not authorized: authentication required")
	}

	mentor, err := uc.mentorRepo.GetByUserID(ctx, mentorUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentor: %w", err)
	}

	if teamID != "" {
		if !mentor.BelongsToTeam(entities.TeamID(teamID)) {
			return nil, fmt.Errorf("not authorized for this team")
		}
		return uc.repo.ListLatestByTeamID(ctx, entities.TeamID(teamID))
	}

	// Fetch progress for each of the mentor's assigned teams
	var all []entities.TeamProgress
	for _, tid := range mentor.TeamIDs {
		progress, err := uc.repo.ListLatestByTeamID(ctx, tid)
		if err != nil {
			return nil, fmt.Errorf("failed to list progress for team %s: %w", tid, err)
		}
		all = append(all, progress...)
	}
	return all, nil
}
