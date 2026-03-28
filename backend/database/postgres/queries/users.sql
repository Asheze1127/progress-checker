-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserBySlackUserID :one
SELECT * FROM users WHERE slack_user_id = $1;

-- name: GetUserWithPasswordByEmail :one
SELECT id, slack_user_id, name, email, role, password_hash FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (slack_user_id, name, email, role, password_hash)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateUserPasswordHash :exec
UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2;

-- name: ListUsers :many
SELECT * FROM users ORDER BY name;
