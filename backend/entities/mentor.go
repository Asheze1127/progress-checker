package entities

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
