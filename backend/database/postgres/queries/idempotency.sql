-- name: IdempotencyKeyExists :one
SELECT EXISTS(
    SELECT 1 FROM idempotency_keys WHERE key = @key AND expires_at > now()
);

-- name: InsertIdempotencyKey :exec
INSERT INTO idempotency_keys (key, expires_at)
VALUES (@key, now() + CAST(@ttl_interval AS interval))
ON CONFLICT (key) DO NOTHING;
