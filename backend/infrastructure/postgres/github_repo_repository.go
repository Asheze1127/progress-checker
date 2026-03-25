package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Asheze1127/progress-checker/backend/application"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// Compile-time interface check.
var _ application.GitHubRepoRepository = (*GitHubRepoRepository)(nil)

// GitHubRepoRepository implements application.GitHubRepoRepository using PostgreSQL.
type GitHubRepoRepository struct {
	db *sql.DB
}

// NewGitHubRepoRepository creates a new GitHubRepoRepository.
func NewGitHubRepoRepository(db *sql.DB) *GitHubRepoRepository {
	return &GitHubRepoRepository{db: db}
}

// Save validates the entity and inserts it into team_github_repositories.
func (r *GitHubRepoRepository) Save(ctx context.Context, repo *entities.GitHubRepo) error {
	if err := repo.Validate(); err != nil {
		return fmt.Errorf("validate github repo: %w", err)
	}

	const query = `
		INSERT INTO team_github_repositories (id, team_id, owner, repo_name, encrypted_pat, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		string(repo.ID),
		string(repo.TeamID),
		repo.Owner,
		repo.RepoName,
		repo.EncryptedToken,
		repo.CreatedAt,
		repo.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert github repo: %w", err)
	}

	return nil
}

// FindByTeamID returns all GitHub repositories for the given team.
func (r *GitHubRepoRepository) FindByTeamID(ctx context.Context, teamID string) ([]entities.GitHubRepo, error) {
	const query = `
		SELECT id, team_id, owner, repo_name, encrypted_pat, created_at, updated_at
		FROM team_github_repositories
		WHERE team_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("query github repos by team_id: %w", err)
	}
	defer rows.Close()

	var repos []entities.GitHubRepo
	for rows.Next() {
		var repo entities.GitHubRepo
		var id, teamIDVal string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &teamIDVal, &repo.Owner, &repo.RepoName, &repo.EncryptedToken, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan github repo row: %w", err)
		}

		repo.ID = entities.GitHubRepoID(id)
		repo.TeamID = entities.TeamID(teamIDVal)
		repo.CreatedAt = createdAt
		repo.UpdatedAt = updatedAt
		repos = append(repos, repo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate github repo rows: %w", err)
	}

	return repos, nil
}

// FindByID returns a GitHub repository by its ID.
func (r *GitHubRepoRepository) FindByID(ctx context.Context, repoID string) (*entities.GitHubRepo, error) {
	const query = `
		SELECT id, team_id, owner, repo_name, encrypted_pat, created_at, updated_at
		FROM team_github_repositories
		WHERE id = $1
	`

	var repo entities.GitHubRepo
	var id, teamID string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, repoID).Scan(
		&id, &teamID, &repo.Owner, &repo.RepoName, &repo.EncryptedToken, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("query github repo by id: %w", err)
	}

	repo.ID = entities.GitHubRepoID(id)
	repo.TeamID = entities.TeamID(teamID)
	repo.CreatedAt = createdAt
	repo.UpdatedAt = updatedAt

	return &repo, nil
}

// FindByChannelID returns the GitHub repository associated with a Slack channel.
// It joins team_github_repositories with slack_channels to find the repo.
func (r *GitHubRepoRepository) FindByChannelID(ctx context.Context, channelID string) (*entities.GitHubRepo, error) {
	const query = `
		SELECT tgr.id, tgr.team_id, tgr.owner, tgr.repo_name, tgr.encrypted_pat, tgr.created_at, tgr.updated_at
		FROM team_github_repositories tgr
		JOIN slack_channels sc ON sc.team_id = tgr.team_id
		WHERE sc.id = $1
		LIMIT 1
	`

	var repo entities.GitHubRepo
	var id, teamID string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, channelID).Scan(
		&id, &teamID, &repo.Owner, &repo.RepoName, &repo.EncryptedToken, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("query github repo by channel_id: %w", err)
	}

	repo.ID = entities.GitHubRepoID(id)
	repo.TeamID = entities.TeamID(teamID)
	repo.CreatedAt = createdAt
	repo.UpdatedAt = updatedAt

	return &repo, nil
}

// Delete removes a GitHub repository by its ID.
func (r *GitHubRepoRepository) Delete(ctx context.Context, repoID string) error {
	const query = `DELETE FROM team_github_repositories WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, repoID)
	if err != nil {
		return fmt.Errorf("delete github repo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("github repo not found: %s", repoID)
	}

	return nil
}

// UpdateToken updates the encrypted personal access token for a GitHub repository.
func (r *GitHubRepoRepository) UpdateToken(ctx context.Context, repoID string, encryptedToken string) error {
	const query = `
		UPDATE team_github_repositories
		SET encrypted_pat = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, encryptedToken, time.Now(), repoID)
	if err != nil {
		return fmt.Errorf("update github repo token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("github repo not found: %s", repoID)
	}

	return nil
}
