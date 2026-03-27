package entities

import (
	"errors"
	"fmt"
	"strings"
)

type ParticipantID string

type Participant struct {
	id     UserID
	TeamID TeamID
}

func NewParticipant(userID UserID, teamID TeamID) *Participant {
	return &Participant{id: userID, TeamID: teamID}
}

func (p *Participant) ID() ParticipantID {
	return ParticipantID(p.id)
}

func (p Participant) Validate() error {
	var errs []error

	if strings.TrimSpace(string(p.id)) == "" {
		errs = append(errs, fmt.Errorf("participant.id is required"))
	}

	if strings.TrimSpace(string(p.TeamID)) == "" {
		errs = append(errs, fmt.Errorf("participant.team_id is required"))
	}

	return errors.Join(errs...)
}
