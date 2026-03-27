// Package idempotency provides a service for checking request idempotency.
package idempotency

import (
	"context"
	"time"
)

// defaultTTL is the default time-to-live for idempotency keys.
const defaultTTL = 1 * time.Hour

// Store defines the interface for checking and storing idempotency keys.
type Store interface {
	// Exists checks whether the given key has already been processed.
	Exists(ctx context.Context, key string) (bool, error)

	// Set stores the key with the specified TTL.
	Set(ctx context.Context, key string, ttl time.Duration) error
}

// Service handles idempotency checking logic.
type Service struct {
	store Store
}

// NewService creates a new idempotency Service with the given store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// IsDuplicate checks if the given key has already been processed.
// If not, it marks the key as processed and returns false.
// Returns true if the request is a duplicate.
// On store errors, returns false to avoid blocking legitimate requests.
func (s *Service) IsDuplicate(ctx context.Context, key string) bool {
	exists, err := s.store.Exists(ctx, key)
	if err != nil {
		return false
	}

	if exists {
		return true
	}

	_ = s.store.Set(ctx, key, defaultTTL)
	return false
}
