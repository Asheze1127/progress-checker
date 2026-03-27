package entities

import "time"

type SetupTokenID string

type SetupToken struct {
	ID        SetupTokenID
	UserID    UserID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

func (t SetupToken) IsExpired(now time.Time) bool {
	return now.After(t.ExpiresAt)
}

func (t SetupToken) IsUsed() bool {
	return t.UsedAt != nil
}
