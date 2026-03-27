package entities

import "context"

// MentorRepository defines the interface for managing mentors.
type MentorRepository interface {
	Create(ctx context.Context, userID UserID) error
}
