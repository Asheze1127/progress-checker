package idempotency

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	idempotencysvc "github.com/Asheze1127/progress-checker/backend/service/idempotency"
)

// TestPostgresStoreImplementsInterface verifies compile-time interface compliance.
func TestPostgresStoreImplementsInterface(t *testing.T) {
	var _ idempotencysvc.Store = (*PostgresStore)(nil)
}

// TestNewPostgresStoreReturnsNonNil verifies the constructor.
func TestNewPostgresStoreReturnsNonNil(t *testing.T) {
	store := NewPostgresStore(&sql.DB{})
	if store == nil {
		t.Fatal("expected non-nil PostgresStore")
	}
	if store.queries == nil {
		t.Fatal("expected queries to be initialized")
	}
}

// mockRow implements sql.Row-like scanning for testing.
type mockRow struct {
	exists bool
	err    error
}

func (r *mockRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) > 0 {
		if p, ok := dest[0].(*bool); ok {
			*p = r.exists
		}
	}
	return nil
}

// mockDB is a test implementation of sqlcgen.DBTX.
type mockDB struct {
	queryRowResult *mockRow
	execErr        error
}

func (m *mockDB) ExecContext(_ context.Context, _ string, _ ...interface{}) (sql.Result, error) {
	return nil, m.execErr
}

func (m *mockDB) PrepareContext(_ context.Context, _ string) (*sql.Stmt, error) {
	return nil, nil
}

func (m *mockDB) QueryContext(_ context.Context, _ string, _ ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (m *mockDB) QueryRowContext(_ context.Context, _ string, _ ...interface{}) *sql.Row {
	// sql.Row cannot be easily mocked, so we test through the store's full flow
	// using a real database or integration tests.
	// For unit tests, we verify construction and interface compliance.
	return nil
}

func TestPostgresStore_Set_success(t *testing.T) {
	db := &mockDB{}
	store := NewPostgresStore(db)

	err := store.Set(context.Background(), "test-key", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresStore_Set_error(t *testing.T) {
	db := &mockDB{execErr: errors.New("db error")}
	store := NewPostgresStore(db)

	err := store.Set(context.Background(), "test-key", time.Hour)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPostgresStore_Set_ttl_conversion(t *testing.T) {
	var capturedArgs []interface{}
	db := &capturingDB{capturedArgs: &capturedArgs}
	store := NewPostgresStore(db)

	ttl := 30 * time.Minute
	_ = store.Set(context.Background(), "my-key", ttl)

	// The second arg after key should be microseconds
	if len(capturedArgs) < 2 {
		t.Fatalf("expected at least 2 args, got %d", len(capturedArgs))
	}
	got, ok := capturedArgs[1].(int64)
	if !ok {
		t.Fatalf("expected int64, got %T", capturedArgs[1])
	}
	want := ttl.Microseconds()
	if got != want {
		t.Errorf("TTL microseconds = %d, want %d", got, want)
	}
}

// capturingDB captures ExecContext args for verification.
type capturingDB struct {
	capturedArgs *[]interface{}
}

func (c *capturingDB) ExecContext(_ context.Context, _ string, args ...interface{}) (sql.Result, error) {
	*c.capturedArgs = args
	return nil, nil
}

func (c *capturingDB) PrepareContext(_ context.Context, _ string) (*sql.Stmt, error) {
	return nil, nil
}

func (c *capturingDB) QueryContext(_ context.Context, _ string, _ ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

func (c *capturingDB) QueryRowContext(_ context.Context, _ string, _ ...interface{}) *sql.Row {
	return nil
}
