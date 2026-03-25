-- name: InsertProgressLog :exec
INSERT INTO progress_logs (id, participant_id)
VALUES (@id, @participant_id);

-- name: InsertProgressBody :exec
INSERT INTO progress_bodies (progress_log_id, phase, sos, comment, submitted_at)
VALUES (@progress_log_id, @phase, @sos, @comment, @submitted_at);
