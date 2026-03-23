package entities

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

