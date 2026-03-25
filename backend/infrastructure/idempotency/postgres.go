// Package idempotency provides implementations of the idempotency Store interface.
package idempotency

import (
	"context"
	"database/sql"
	"time"

	idempotencysvc "github.com/Asheze1127/progress-checker/backend/service/idempotency"
)

// Compile-time check that PostgresStore implements idempotencysvc.Store.
var _ idempotencysvc.Store = (*PostgresStore)(nil)

// PostgresStore is a PostgreSQL-backed implementation of idempotencysvc.Store.
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
