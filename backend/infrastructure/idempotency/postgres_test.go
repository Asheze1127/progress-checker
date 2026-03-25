package idempotency

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	idempotencysvc "github.com/Asheze1127/progress-checker/backend/service/idempotency"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/sqlcgen"
)

func testDBTX(t *testing.T) sqlcgen.DBTX {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func cleanupKeys(t *testing.T, db sqlcgen.DBTX) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), "DELETE FROM idempotency_keys")
	if err != nil {
		t.Fatalf("failed to cleanup: %v", err)
	}
}

func TestPostgresStoreImplementsInterface(t *testing.T) {
	var _ idempotencysvc.Store = (*PostgresStore)(nil)
}

func TestPostgresStore_Exists_returns_false_for_missing_key(t *testing.T) {
	db := testDBTX(t)
	cleanupKeys(t, db)
	store := NewPostgresStore(db)

	exists, err := store.Exists(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatal("expected false for missing key")
	}
}

func TestPostgresStore_Set_and_Exists(t *testing.T) {
	db := testDBTX(t)
	cleanupKeys(t, db)
	store := NewPostgresStore(db)
	ctx := context.Background()

	err := store.Set(ctx, "test-key", time.Hour)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	exists, err := store.Exists(ctx, "test-key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected key to exist after Set")
	}
}

func TestPostgresStore_Set_duplicate_does_not_error(t *testing.T) {
	db := testDBTX(t)
	cleanupKeys(t, db)
	store := NewPostgresStore(db)
	ctx := context.Background()

	if err := store.Set(ctx, "dup-key", time.Hour); err != nil {
		t.Fatalf("first Set failed: %v", err)
	}
	if err := store.Set(ctx, "dup-key", time.Hour); err != nil {
		t.Fatalf("duplicate Set should not error: %v", err)
	}
}

func TestPostgresStore_Exists_returns_false_for_expired_key(t *testing.T) {
	db := testDBTX(t)
	cleanupKeys(t, db)
	store := NewPostgresStore(db)
	ctx := context.Background()

	// Insert an already-expired key via sqlc queries
	q := sqlcgen.New(db)
	err := q.InsertIdempotencyKey(ctx, sqlcgen.InsertIdempotencyKeyParams{
		Key:         "expired-key",
		TtlInterval: -1_000_000, // -1 second in microseconds
	})
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	exists, err := store.Exists(ctx, "expired-key")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("expected false for expired key")
	}
}
