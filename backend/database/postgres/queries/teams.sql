-- name: GetTeamByID :one
SELECT * FROM teams WHERE id = $1;

-- name: ListTeams :many
SELECT * FROM teams ORDER BY name;

-- name: CreateTeam :one
INSERT INTO teams (name) VALUES ($1) RETURNING *;

-- name: GetTeamByName :one
SELECT * FROM teams WHERE name = $1;
