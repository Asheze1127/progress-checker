-- name: CheckIdempotencyKeyExists :one
SELECT EXISTS(SELECT 1 FROM idempotency_keys WHERE key = $1 AND expires_at > now());

-- name: SetIdempotencyKey :exec
INSERT INTO idempotency_keys (key, expires_at) VALUES ($1, $2) ON CONFLICT (key) DO NOTHING;

-- name: DeleteExpiredIdempotencyKeys :exec
DELETE FROM idempotency_keys WHERE expires_at <= now();
