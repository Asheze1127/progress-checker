-- name: InsertProgressLog :one
INSERT INTO progress_logs (id, participant_id) VALUES ($1, $2) RETURNING *;

-- name: InsertProgressBody :one
INSERT INTO progress_bodies (id, progress_log_id, phase, sos, comment, submitted_at)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetLatestProgressByTeam :many
SELECT t.id AS team_id, t.name AS team_name,
       pl.id AS progress_log_id, pl.participant_id, pl.created_at AS log_created_at,
       pb.id AS body_id, pb.phase, pb.sos, pb.comment, pb.submitted_at
FROM teams t
LEFT JOIN LATERAL (
    SELECT pl2.* FROM progress_logs pl2
    JOIN participants p ON p.user_id = pl2.participant_id
    WHERE p.team_id = t.id ORDER BY pl2.created_at DESC LIMIT 1
) pl ON true
LEFT JOIN progress_bodies pb ON pb.progress_log_id = pl.id
ORDER BY t.name, pb.submitted_at DESC;

-- name: GetLatestProgressByTeamID :many
SELECT t.id AS team_id, t.name AS team_name,
       pl.id AS progress_log_id, pl.participant_id, pl.created_at AS log_created_at,
       pb.id AS body_id, pb.phase, pb.sos, pb.comment, pb.submitted_at
FROM teams t
LEFT JOIN LATERAL (
    SELECT pl2.* FROM progress_logs pl2
    JOIN participants p ON p.user_id = pl2.participant_id
    WHERE p.team_id = t.id ORDER BY pl2.created_at DESC LIMIT 1
) pl ON true
LEFT JOIN progress_bodies pb ON pb.progress_log_id = pl.id
WHERE t.id = $1 ORDER BY pb.submitted_at DESC;
