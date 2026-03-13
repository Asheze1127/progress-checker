package entities

import "time"

// SessionToken represents an opaque web session token stored as a hash.
type SessionToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}
