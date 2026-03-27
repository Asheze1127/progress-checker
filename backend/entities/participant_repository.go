package entities

import "context"

// ParticipantRepository defines the interface for managing participants.
type ParticipantRepository interface {
	Create(ctx context.Context, userID UserID, teamID TeamID) error
}
