package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Asheze1127/progress-checker/backend/application/appcontext"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

type ListMentorTeamsUseCase struct {
	mentorRepo entities.MentorRepository
	teamRepo   entities.TeamRepository
}

func NewListMentorTeamsUseCase(mentorRepo entities.MentorRepository, teamRepo entities.TeamRepository) *ListMentorTeamsUseCase {
	return &ListMentorTeamsUseCase{mentorRepo: mentorRepo, teamRepo: teamRepo}
}

func (uc *ListMentorTeamsUseCase) Execute(ctx context.Context) (result []*entities.Team, err error) {
	defer func() {
		attrs := []slog.Attr{slog.Int("count", len(result))}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
		slog.LogAttrs(ctx, slog.LevelDebug, "ListMentorTeamsUseCase.Execute", attrs...)
	}()

	mentorUser := appcontext.UserFromContext(ctx)
	if mentorUser == nil {
		return nil, fmt.Errorf("not authorized: authentication required")
	}

	mentor, err := uc.mentorRepo.GetByUserID(ctx, mentorUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentor: %w", err)
	}

	teams := make([]*entities.Team, 0, len(mentor.TeamIDs))
	for _, tid := range mentor.TeamIDs {
		team, err := uc.teamRepo.GetByID(ctx, tid)
		if err != nil {
			return nil, fmt.Errorf("failed to get team %s: %w", tid, err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}
