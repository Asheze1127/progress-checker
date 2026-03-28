package entities

import (
	"context"
	"time"
)

// SetupTokenRepository defines the interface for managing setup tokens.
type SetupTokenRepository interface {
	Create(ctx context.Context, userID UserID, tokenHash string, expiresAt time.Time) (*SetupToken, error)
	FindByHash(ctx context.Context, tokenHash string) (*SetupToken, error)
	MarkUsed(ctx context.Context, id SetupTokenID) error
}
