package entities

import "context"

// TeamParticipant holds participant info joined with user data.
type TeamParticipant struct {
	ID          UserID
	SlackUserID SlackUserID
	Name        string
	Email       string
}

// ParticipantRepository defines the interface for managing participants.
type ParticipantRepository interface {
	Create(ctx context.Context, userID UserID, teamID TeamID) error
	ListByTeamID(ctx context.Context, teamID TeamID) ([]TeamParticipant, error)
}
