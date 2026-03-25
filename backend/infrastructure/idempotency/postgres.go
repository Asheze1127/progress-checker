// Package idempotency provides implementations of the idempotency Store interface.
package idempotency

import (
	"context"
	"time"

	idempotencysvc "github.com/Asheze1127/progress-checker/backend/service/idempotency"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/sqlcgen"
)

// Compile-time check that PostgresStore implements idempotencysvc.Store.
var _ idempotencysvc.Store = (*PostgresStore)(nil)

// PostgresStore is a PostgreSQL-backed implementation of idempotencysvc.Store
// using sqlc-generated queries.
type PostgresStore struct {
	queries *sqlcgen.Queries
}

// NewPostgresStore creates a new PostgresStore backed by the given sqlc DBTX.
func NewPostgresStore(db sqlcgen.DBTX) *PostgresStore {
	return &PostgresStore{queries: sqlcgen.New(db)}
}

// Exists checks whether the given key exists and has not expired.
func (s *PostgresStore) Exists(ctx context.Context, key string) (bool, error) {
	return s.queries.IdempotencyKeyExists(ctx, key)
}

// Set stores the key with the specified TTL.
func (s *PostgresStore) Set(ctx context.Context, key string, ttl time.Duration) error {
	return s.queries.InsertIdempotencyKey(ctx, sqlcgen.InsertIdempotencyKeyParams{
		Key:         key,
		TtlInterval: ttl.Microseconds(),
	})
}
