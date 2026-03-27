package usecase

import (
  "context"

  "github.com/Asheze1127/progress-checker/backend/entities"
)

// ListProgressUseCase orchestrates fetching the latest progress data,
// optionally filtered by team ID.
type ListProgressUseCase struct {
  repo entities.ProgressQueryRepository
}

// NewListProgressUseCase creates a new ListProgressUseCase.
func NewListProgressUseCase(repo entities.ProgressQueryRepository) *ListProgressUseCase {
  return &ListProgressUseCase{repo: repo}
}

// Execute returns the latest progress for all teams, or for a specific team
// when teamID is non-empty.
func (uc *ListProgressUseCase) Execute(ctx context.Context, teamID string) ([]entities.TeamProgress, error) {
  if teamID != "" {
    return uc.repo.ListLatestByTeamID(ctx, entities.TeamID(teamID))
  }
  return uc.repo.ListLatestByTeam(ctx)
}
