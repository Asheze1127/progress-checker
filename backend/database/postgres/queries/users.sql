-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserBySlackUserID :one
SELECT * FROM users WHERE slack_user_id = $1;

-- name: GetUserWithPasswordByEmail :one
SELECT id, slack_user_id, name, email, role, password_hash FROM users WHERE email = $1;
