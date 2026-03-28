-- name: GetStaffByEmail :one
SELECT id, slack_user_id, name, email, password_hash FROM staff WHERE email = $1;

-- name: GetStaffByID :one
SELECT * FROM staff WHERE id = $1;

-- name: GetStaffBySlackUserID :one
SELECT * FROM staff WHERE slack_user_id = $1;

-- name: UpdateStaffSlackUserID :exec
UPDATE staff SET slack_user_id = $2, updated_at = now() WHERE id = $1;
