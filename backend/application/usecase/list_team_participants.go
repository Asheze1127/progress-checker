package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Asheze1127/progress-checker/backend/application/appcontext"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

var (
	ErrNotAuthorized       = errors.New("not authorized")
	ErrNotAuthorizedForTeam = errors.New("not authorized for this team")
)

type ListTeamParticipantsUseCase struct {
	participantRepo entities.ParticipantRepository
	mentorRepo      entities.MentorRepository
}

func NewListTeamParticipantsUseCase(
	participantRepo entities.ParticipantRepository,
	mentorRepo entities.MentorRepository,
) *ListTeamParticipantsUseCase {
	return &ListTeamParticipantsUseCase{
		participantRepo: participantRepo,
		mentorRepo:      mentorRepo,
	}
}

func (uc *ListTeamParticipantsUseCase) Execute(ctx context.Context, teamID string) (result []entities.TeamParticipant, err error) {
	defer func() {
		attrs := []slog.Attr{slog.String("team_id", teamID), slog.Int("count", len(result))}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
		slog.LogAttrs(ctx, slog.LevelDebug, "ListTeamParticipantsUseCase.Execute", attrs...)
	}()

	mentorUser := appcontext.UserFromContext(ctx)
	if mentorUser == nil {
		return nil, ErrNotAuthorized
	}

	mentor, err := uc.mentorRepo.GetByUserID(ctx, mentorUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentor: %w", err)
	}

	if !mentor.BelongsToTeam(entities.TeamID(teamID)) {
		return nil, ErrNotAuthorizedForTeam
	}

	participants, err := uc.participantRepo.ListByTeamID(ctx, entities.TeamID(teamID))
	if err != nil {
		return nil, fmt.Errorf("failed to list participants: %w", err)
	}

	return participants, nil
}
