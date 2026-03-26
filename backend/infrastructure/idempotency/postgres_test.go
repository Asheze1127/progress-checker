package idempotency

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
)

func testDBTX(t *testing.T) db.DBTX {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := conn.PingContext(context.Background()); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func cleanupKey(t *testing.T, dbtx db.DBTX, key string) {
	t.Helper()
	_, err := dbtx.ExecContext(context.Background(), "DELETE FROM idempotency_keys WHERE key = $1", key)
	if err != nil {
		t.Fatalf("failed to cleanup key %q: %v", key, err)
	}
}

func TestPostgresStore_Exists_returns_false_for_missing_key(t *testing.T) {
	dbtx := testDBTX(t)
	store := NewPostgresStore(dbtx)

	exists, err := store.Exists(context.Background(), "test-missing-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatal("expected false for missing key")
	}
}

func TestPostgresStore_Set_and_Exists(t *testing.T) {
	dbtx := testDBTX(t)
	store := NewPostgresStore(dbtx)
	ctx := context.Background()
	key := "test-set-exists-key"
	t.Cleanup(func() { cleanupKey(t, dbtx, key) })

	err := store.Set(ctx, key, time.Hour)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	exists, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected key to exist after Set")
	}
}

func TestPostgresStore_Set_duplicate_does_not_error(t *testing.T) {
	dbtx := testDBTX(t)
	store := NewPostgresStore(dbtx)
	ctx := context.Background()
	key := "test-dup-key"
	t.Cleanup(func() { cleanupKey(t, dbtx, key) })

	if err := store.Set(ctx, key, time.Hour); err != nil {
		t.Fatalf("first Set failed: %v", err)
	}
	if err := store.Set(ctx, key, time.Hour); err != nil {
		t.Fatalf("duplicate Set should not error: %v", err)
	}

	exists, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists after duplicate Set failed: %v", err)
	}
	if !exists {
		t.Fatal("expected key to still exist after duplicate Set")
	}
}

func TestPostgresStore_Exists_returns_false_for_expired_key(t *testing.T) {
	dbtx := testDBTX(t)
	store := NewPostgresStore(dbtx)
	ctx := context.Background()
	key := "test-expired-key"
	t.Cleanup(func() { cleanupKey(t, dbtx, key) })

	// Insert an already-expired key via sqlc queries directly
	q := db.New(dbtx)
	err := q.SetIdempotencyKey(ctx, db.SetIdempotencyKeyParams{
		Key:       key,
		ExpiresAt: time.Now().Add(-1 * time.Second),
	})
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	exists, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("expected false for expired key")
	}
}
