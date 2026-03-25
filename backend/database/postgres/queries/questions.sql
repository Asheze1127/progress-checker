-- name: FindQuestionByID :one
SELECT id, participant_id, title, slack_channel_id, status, slack_thread_ts
FROM questions
WHERE id = @id;

-- name: FindMentorIDsByQuestionID :many
SELECT mentor_user_id
FROM question_mentor_assignments
WHERE question_id = @question_id;

-- name: UpdateQuestionStatus :execresult
UPDATE questions
SET status = @status, updated_at = now()
WHERE id = @id;
