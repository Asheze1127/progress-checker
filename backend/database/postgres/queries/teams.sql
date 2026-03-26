-- name: GetTeamByID :one
SELECT * FROM teams WHERE id = $1;

-- name: ListTeams :many
SELECT * FROM teams ORDER BY name;
