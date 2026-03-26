package entities

import "context"

// UserRepository defines the interface for querying users.
type UserRepository interface {
	GetByID(ctx context.Context, id UserID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetBySlackUserID(ctx context.Context, slackUserID SlackUserID) (*User, error)
}
