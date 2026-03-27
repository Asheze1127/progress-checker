-- name: CreateMentor :exec
INSERT INTO mentors (user_id) VALUES ($1);

-- name: CreateMentorTeamAssignment :exec
INSERT INTO mentor_team_assignments (mentor_user_id, team_id) VALUES ($1, $2);

-- name: GetMentorTeamIDs :many
SELECT team_id FROM mentor_team_assignments WHERE mentor_user_id = $1;
