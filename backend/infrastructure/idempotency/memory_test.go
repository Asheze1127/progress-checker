package idempotency

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStore(t *testing.T) {
	t.Run("set and exists returns true", func(t *testing.T) {
		store := NewMemoryStore(1 * time.Hour)
		defer store.Close()

		ctx := context.Background()
		key := "test-key"

		if err := store.Set(ctx, key, 10*time.Minute); err != nil {
			t.Fatalf("unexpected error on Set: %v", err)
		}

		exists, err := store.Exists(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error on Exists: %v", err)
		}
		if !exists {
			t.Fatal("expected key to exist")
		}
	})

	t.Run("non-existent key returns false", func(t *testing.T) {
		store := NewMemoryStore(1 * time.Hour)
		defer store.Close()

		exists, err := store.Exists(context.Background(), "missing-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exists {
			t.Fatal("expected key to not exist")
		}
	})

	t.Run("expired key returns false", func(t *testing.T) {
		store := NewMemoryStore(1 * time.Hour)
		defer store.Close()

		now := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
		store.nowFunc = func() time.Time { return now }

		ctx := context.Background()
		key := "expiring-key"

		if err := store.Set(ctx, key, 5*time.Minute); err != nil {
			t.Fatalf("unexpected error on Set: %v", err)
		}

		// Advance time past expiration.
		store.nowFunc = func() time.Time { return now.Add(6 * time.Minute) }

		exists, err := store.Exists(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error on Exists: %v", err)
		}
		if exists {
			t.Fatal("expected expired key to not exist")
		}
	})

	t.Run("cleanup removes expired entries", func(t *testing.T) {
		store := NewMemoryStore(1 * time.Hour)
		defer store.Close()

		now := time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
		store.nowFunc = func() time.Time { return now }

		ctx := context.Background()

		// Store one short-lived and one long-lived key.
		if err := store.Set(ctx, "short", 1*time.Minute); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := store.Set(ctx, "long", 1*time.Hour); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Advance time past the short key's TTL but before the long key's.
		store.nowFunc = func() time.Time { return now.Add(2 * time.Minute) }
		store.removeExpired()

		shortExists, _ := store.Exists(ctx, "short")
		longExists, _ := store.Exists(ctx, "long")

		if shortExists {
			t.Error("expected short-lived key to be cleaned up")
		}
		if !longExists {
			t.Error("expected long-lived key to still exist")
		}
	})
}
