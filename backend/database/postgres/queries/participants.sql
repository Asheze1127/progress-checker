-- name: CreateParticipant :exec
INSERT INTO participants (user_id, team_id) VALUES ($1, $2);
