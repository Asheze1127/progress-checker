-- name: CreateSetupToken :one
INSERT INTO setup_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSetupTokenByHash :one
SELECT * FROM setup_tokens WHERE token_hash = $1;

-- name: MarkSetupTokenUsed :exec
UPDATE setup_tokens SET used_at = now() WHERE id = $1;
