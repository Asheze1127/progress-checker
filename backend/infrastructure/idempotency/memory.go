// Package idempotency provides implementations of the IdempotencyStore interface.
package idempotency

import (
	"context"
	"sync"
	"time"
)

// entry holds a stored key with its expiration time.
type entry struct {
	expiresAt time.Time
}

// MemoryStore is an in-memory implementation of middleware.IdempotencyStore.
// It uses sync.Map for concurrent access and periodically cleans up expired entries.
type MemoryStore struct {
	entries sync.Map
	nowFunc func() time.Time // for testing
	stop    chan struct{}
	once    sync.Once
}

// NewMemoryStore creates a new MemoryStore and starts a background cleanup goroutine
// that runs at the specified interval.
func NewMemoryStore(cleanupInterval time.Duration) *MemoryStore {
	store := &MemoryStore{
		nowFunc: time.Now,
		stop:    make(chan struct{}),
	}

	go store.cleanupLoop(cleanupInterval)
	return store
}

// Exists checks whether the given key exists and has not expired.
func (s *MemoryStore) Exists(_ context.Context, key string) (bool, error) {
	val, ok := s.entries.Load(key)
	if !ok {
		return false, nil
	}

	e := val.(entry)
	if s.nowFunc().After(e.expiresAt) {
		s.entries.Delete(key)
		return false, nil
	}

	return true, nil
}

// Set stores the key with the specified TTL.
func (s *MemoryStore) Set(_ context.Context, key string, ttl time.Duration) error {
	s.entries.Store(key, entry{
		expiresAt: s.nowFunc().Add(ttl),
	})
	return nil
}

// Close stops the background cleanup goroutine.
func (s *MemoryStore) Close() {
	s.once.Do(func() {
		close(s.stop)
	})
}

// cleanupLoop periodically removes expired entries.
func (s *MemoryStore) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.removeExpired()
		case <-s.stop:
			return
		}
	}
}

// removeExpired iterates over all entries and deletes those that have expired.
func (s *MemoryStore) removeExpired() {
	now := s.nowFunc()
	s.entries.Range(func(key, value any) bool {
		e := value.(entry)
		if now.After(e.expiresAt) {
			s.entries.Delete(key)
		}
		return true
	})
}
