package usecase

import (
	"context"

	"github.com/samber/do/v2"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// ListProgressUseCase orchestrates fetching the latest progress data,
// optionally filtered by team ID.
type ListProgressUseCase struct {
	repo entities.ProgressQueryRepository
}

// NewListProgressUseCase creates a new ListProgressUseCase via DI container.
func NewListProgressUseCase(i do.Injector) (*ListProgressUseCase, error) {
	repo := do.MustInvoke[entities.ProgressQueryRepository](i)
	return &ListProgressUseCase{repo: repo}, nil
}

// Execute returns the latest progress for all teams, or for a specific team
// when teamID is non-empty.
func (uc *ListProgressUseCase) Execute(ctx context.Context, teamID string) ([]entities.TeamProgress, error) {
	if teamID != "" {
		return uc.repo.ListLatestByTeamID(ctx, entities.TeamID(teamID))
	}
	return uc.repo.ListLatestByTeam(ctx)
}
