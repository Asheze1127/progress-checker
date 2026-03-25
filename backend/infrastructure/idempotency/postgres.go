// Package idempotency provides implementations of the IdempotencyStore interface.
package idempotency

import (
	"context"
	"database/sql"
	"time"

	"github.com/Asheze1127/progress-checker/backend/api/middleware"
)

// Compile-time check that PostgresStore implements middleware.IdempotencyStore.
var _ middleware.IdempotencyStore = (*PostgresStore)(nil)

// PostgresStore is a PostgreSQL-backed implementation of middleware.IdempotencyStore.
// It uses the idempotency_keys table for persistence across multiple instances.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new PostgresStore backed by the given database connection.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// Exists checks whether the given key exists and has not expired.
func (s *PostgresStore) Exists(ctx context.Context, key string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM idempotency_keys WHERE key = $1 AND expires_at > now())",
		key,
	).Scan(&exists)
	return exists, err
}

// Set stores the key with the specified TTL using an atomic upsert.
func (s *PostgresStore) Set(ctx context.Context, key string, ttl time.Duration) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO idempotency_keys (key, expires_at) VALUES ($1, now() + $2::interval) ON CONFLICT (key) DO NOTHING",
		key, ttl.String(),
	)
	return err
}
