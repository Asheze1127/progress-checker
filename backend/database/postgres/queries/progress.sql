-- name: InsertProgressLog :one
INSERT INTO progress_logs (id, participant_id) VALUES ($1, $2) RETURNING *;

-- name: InsertProgressBody :one
INSERT INTO progress_bodies (id, progress_log_id, phase, sos, comment, submitted_at)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetLatestProgressByTeam :many
WITH latest_logs AS (
    SELECT DISTINCT ON (p.team_id)
           pl.id, pl.participant_id, pl.created_at, p.team_id
    FROM progress_logs pl
    JOIN participants p ON p.user_id = pl.participant_id
    ORDER BY p.team_id, pl.created_at DESC
)
SELECT t.id AS team_id, t.name AS team_name,
       ll.id AS progress_log_id, ll.participant_id, ll.created_at AS log_created_at,
       pb.id AS body_id, pb.phase, pb.sos, pb.comment, pb.submitted_at
FROM teams t
LEFT JOIN latest_logs ll ON ll.team_id = t.id
LEFT JOIN progress_bodies pb ON pb.progress_log_id = ll.id
ORDER BY t.name, pb.submitted_at DESC;

-- name: GetLatestProgressByTeamID :many
WITH latest_logs AS (
    SELECT DISTINCT ON (p.team_id)
           pl.id, pl.participant_id, pl.created_at, p.team_id
    FROM progress_logs pl
    JOIN participants p ON p.user_id = pl.participant_id
    ORDER BY p.team_id, pl.created_at DESC
)
SELECT t.id AS team_id, t.name AS team_name,
       ll.id AS progress_log_id, ll.participant_id, ll.created_at AS log_created_at,
       pb.id AS body_id, pb.phase, pb.sos, pb.comment, pb.submitted_at
FROM teams t
LEFT JOIN latest_logs ll ON ll.team_id = t.id
LEFT JOIN progress_bodies pb ON pb.progress_log_id = ll.id
WHERE t.id = $1 ORDER BY pb.submitted_at DESC;
