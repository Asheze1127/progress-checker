package entities

import (
	"errors"
	"fmt"
	"strings"
)

type MentorID string

type Mentor struct {
	id      UserID
	TeamIDs []TeamID
}

func NewMentor(userID UserID, teamIDs []TeamID) *Mentor {
	return &Mentor{id: userID, TeamIDs: teamIDs}
}

func (m *Mentor) ID() MentorID {
	return MentorID(m.id)
}

func (m Mentor) Validate() error {
	var errs []error

	if strings.TrimSpace(string(m.id)) == "" {
		errs = append(errs, fmt.Errorf("mentor.id is required"))
	}

	seenTeamIDs := make(map[TeamID]struct{}, len(m.TeamIDs))

	for i, teamID := range m.TeamIDs {
		if strings.TrimSpace(string(teamID)) == "" {
			errs = append(errs, fmt.Errorf("mentor.team_ids[%d] is required", i))
		}

		if _, ok := seenTeamIDs[teamID]; ok {
			errs = append(errs, fmt.Errorf("mentor.team_ids contains duplicate value %q", teamID))
			continue
		}

		seenTeamIDs[teamID] = struct{}{}
	}

	return errors.Join(errs...)
}
