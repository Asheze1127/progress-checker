package idempotency

import (
	"database/sql"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/api/middleware"
)

// TestPostgresStoreImplementsInterface verifies that PostgresStore satisfies
// the IdempotencyStore interface at compile time.
func TestPostgresStoreImplementsInterface(t *testing.T) {
	var _ middleware.IdempotencyStore = (*PostgresStore)(nil)
}

// TestNewPostgresStoreReturnsNonNil verifies that the constructor returns a
// non-nil store when given a database handle.
func TestNewPostgresStoreReturnsNonNil(t *testing.T) {
	// Use a nil *sql.DB since we are only testing construction, not queries.
	store := NewPostgresStore(&sql.DB{})
	if store == nil {
		t.Fatal("expected non-nil PostgresStore")
	}
	if store.db == nil {
		t.Fatal("expected db field to be set")
	}
}
