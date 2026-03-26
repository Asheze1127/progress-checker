-- name: InsertGitHubRepo :one
INSERT INTO team_github_repositories (id, team_id, github_repo_url, encrypted_pat)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetGitHubReposByTeamID :many
SELECT * FROM team_github_repositories WHERE team_id = $1;

-- name: GetGitHubRepoByID :one
SELECT * FROM team_github_repositories WHERE id = $1;

-- name: GetGitHubRepoByChannelID :one
SELECT tgr.* FROM team_github_repositories tgr
JOIN slack_channels sc ON sc.team_id = tgr.team_id WHERE sc.id = $1 LIMIT 1;

-- name: DeleteGitHubRepo :exec
DELETE FROM team_github_repositories WHERE id = $1;

-- name: UpdateGitHubRepoToken :exec
UPDATE team_github_repositories SET encrypted_pat = $1, updated_at = now() WHERE id = $2;
