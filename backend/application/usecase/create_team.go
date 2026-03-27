package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

type CreateTeamUseCase struct {
	teamRepo entities.TeamRepository
}

func NewCreateTeamUseCase(teamRepo entities.TeamRepository) *CreateTeamUseCase {
	return &CreateTeamUseCase{teamRepo: teamRepo}
}

func (uc *CreateTeamUseCase) Execute(ctx context.Context, name string) (*entities.Team, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("team name is required")
	}

	team, err := uc.teamRepo.Create(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}
	return team, nil
}
