package entities

import "context"

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	// FindByEmail retrieves a user with their password hash by email address.
	FindByEmail(ctx context.Context, email string) (*UserWithPassword, error)
	// FindByID retrieves a user by their ID.
	FindByID(ctx context.Context, id UserID) (*User, error)
}

// UserWithPassword extends User with a hashed password for credential verification.
type UserWithPassword struct {
	User
	PasswordHash string
}
