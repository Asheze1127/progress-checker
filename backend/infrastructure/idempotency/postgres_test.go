package idempotency

import (
	"database/sql"
	"testing"

	idempotencysvc "github.com/Asheze1127/progress-checker/backend/service/idempotency"
)

// TestPostgresStoreImplementsInterface verifies that PostgresStore satisfies
// the Store interface at compile time.
func TestPostgresStoreImplementsInterface(t *testing.T) {
	var _ idempotencysvc.Store = (*PostgresStore)(nil)
}

// TestNewPostgresStoreReturnsNonNil verifies that the constructor returns a
// non-nil store when given a database handle.
func TestNewPostgresStoreReturnsNonNil(t *testing.T) {
	store := NewPostgresStore(&sql.DB{})
	if store == nil {
		t.Fatal("expected non-nil PostgresStore")
	}
	if store.db == nil {
		t.Fatal("expected db field to be set")
	}
}
