-- name: CreateParticipant :exec
INSERT INTO participants (user_id, team_id) VALUES ($1, $2);

-- name: ListParticipantsByTeamID :many
SELECT u.id, u.slack_user_id, u.name, u.email, p.created_at
FROM participants p
JOIN users u ON u.id = p.user_id
WHERE p.team_id = $1
ORDER BY p.created_at DESC;
