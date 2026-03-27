-- name: CreateMentor :exec
INSERT INTO mentors (user_id) VALUES ($1);
