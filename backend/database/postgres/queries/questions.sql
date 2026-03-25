-- name: InsertQuestion :one
INSERT INTO questions (id, participant_id, title, slack_channel_id, status, slack_thread_ts)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetQuestionByID :one
SELECT * FROM questions WHERE id = $1;

-- name: GetQuestionByThreadTS :one
SELECT * FROM questions WHERE slack_channel_id = $1 AND slack_thread_ts = $2;

-- name: GetAwaitingQuestionByChannelAndThread :one
SELECT * FROM questions WHERE slack_channel_id = $1 AND slack_thread_ts = $2 AND status = 'awaiting_user';

-- name: UpdateQuestionStatus :exec
UPDATE questions SET status = $1, updated_at = now() WHERE id = $2;

-- name: InsertQuestionMentorAssignment :exec
INSERT INTO question_mentor_assignments (question_id, mentor_user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;
