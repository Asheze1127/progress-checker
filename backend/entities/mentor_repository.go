package entities

import "context"

// MentorRepository defines the interface for managing mentors.
type MentorRepository interface {
	Create(ctx context.Context, userID UserID) error
	AssignTeam(ctx context.Context, userID UserID, teamID TeamID) error
	GetTeamIDs(ctx context.Context, userID UserID) ([]TeamID, error)
}
