package entities

import "context"

// UserRepository defines the interface for querying users.
type UserRepository interface {
  GetByID(ctx context.Context, id UserID) (*User, error)
  GetByEmail(ctx context.Context, email string) (*User, error)
  GetBySlackUserID(ctx context.Context, slackUserID SlackUserID) (*User, error)
  // FindByEmail retrieves a user with their password hash by email address.
  FindByEmail(ctx context.Context, email string) (*UserWithPassword, error)
}

// UserWithPassword extends User with a hashed password for credential verification.
type UserWithPassword struct {
  User
  PasswordHash string
}
