package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/application"
	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

// Compile-time interface check.
var _ application.GitHubRepoRepository = (*GitHubRepoRepository)(nil)

// GitHubRepoRepository implements application.GitHubRepoRepository using sqlc-generated queries.
type GitHubRepoRepository struct {
	queries *db.Queries
}

// NewGitHubRepoRepository creates a new GitHubRepoRepository backed by the given database connection.
func NewGitHubRepoRepository(database *sql.DB) *GitHubRepoRepository {
	return &GitHubRepoRepository{queries: db.New(database)}
}

// Save validates the entity and inserts it into team_github_repositories.
func (r *GitHubRepoRepository) Save(ctx context.Context, repo *entities.GitHubRepo) error {
	if err := repo.Validate(); err != nil {
		return fmt.Errorf("validate github repo: %w", err)
	}

	repoID, err := uuid.Parse(string(repo.ID))
	if err != nil {
		return fmt.Errorf("parse repo id: %w", err)
	}

	teamID, err := uuid.Parse(string(repo.TeamID))
	if err != nil {
		return fmt.Errorf("parse team id: %w", err)
	}

	repoURL := buildGitHubRepoURL(repo.Owner, repo.RepoName)

	_, err = r.queries.InsertGitHubRepo(ctx, db.InsertGitHubRepoParams{
		ID:            repoID,
		TeamID:        teamID,
		GithubRepoUrl: repoURL,
		EncryptedPat:  repo.EncryptedToken,
	})
	if err != nil {
		return fmt.Errorf("insert github repo: %w", err)
	}

	return nil
}

// FindByTeamID returns all GitHub repositories for the given team.
func (r *GitHubRepoRepository) FindByTeamID(ctx context.Context, teamID string) ([]entities.GitHubRepo, error) {
	uid, err := uuid.Parse(teamID)
	if err != nil {
		return nil, fmt.Errorf("parse team id: %w", err)
	}

	rows, err := r.queries.GetGitHubReposByTeamID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("query github repos by team_id: %w", err)
	}

	repos := make([]entities.GitHubRepo, 0, len(rows))
	for _, row := range rows {
		repo, err := toGitHubRepoEntity(row)
		if err != nil {
			return nil, fmt.Errorf("convert github repo row: %w", err)
		}
		repos = append(repos, *repo)
	}

	return repos, nil
}

// FindByID returns a GitHub repository by its ID.
func (r *GitHubRepoRepository) FindByID(ctx context.Context, repoID string) (*entities.GitHubRepo, error) {
	uid, err := uuid.Parse(repoID)
	if err != nil {
		return nil, fmt.Errorf("parse repo id: %w", err)
	}

	row, err := r.queries.GetGitHubRepoByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("query github repo by id: %w", err)
	}

	return toGitHubRepoEntity(row)
}

// FindByChannelID returns the GitHub repository associated with a Slack channel.
func (r *GitHubRepoRepository) FindByChannelID(ctx context.Context, channelID string) (*entities.GitHubRepo, error) {
	row, err := r.queries.GetGitHubRepoByChannelID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("query github repo by channel_id: %w", err)
	}

	return toGitHubRepoEntity(row)
}

// Delete removes a GitHub repository by its ID.
func (r *GitHubRepoRepository) Delete(ctx context.Context, repoID string) error {
	uid, err := uuid.Parse(repoID)
	if err != nil {
		return fmt.Errorf("parse repo id: %w", err)
	}

	return r.queries.DeleteGitHubRepo(ctx, uid)
}

// UpdateToken updates the encrypted personal access token for a GitHub repository.
func (r *GitHubRepoRepository) UpdateToken(ctx context.Context, repoID string, encryptedToken string) error {
	uid, err := uuid.Parse(repoID)
	if err != nil {
		return fmt.Errorf("parse repo id: %w", err)
	}

	return r.queries.UpdateGitHubRepoToken(ctx, db.UpdateGitHubRepoTokenParams{
		EncryptedPat: encryptedToken,
		ID:           uid,
	})
}

// buildGitHubRepoURL constructs a GitHub repository URL from owner and repo name.
func buildGitHubRepoURL(owner, repoName string) string {
	return fmt.Sprintf("https://github.com/%s/%s", owner, repoName)
}

// toGitHubRepoEntity converts a sqlc-generated row to a domain entity.
func toGitHubRepoEntity(row db.TeamGithubRepositories) (*entities.GitHubRepo, error) {
	owner, repoName, err := entities.ParseGitHubRepoURL(row.GithubRepoUrl)
	if err != nil {
		return nil, fmt.Errorf("parse github repo url from db: %w", err)
	}

	return &entities.GitHubRepo{
		ID:             entities.GitHubRepoID(row.ID.String()),
		TeamID:         entities.TeamID(row.TeamID.String()),
		Owner:          owner,
		RepoName:       repoName,
		EncryptedToken: row.EncryptedPat,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}, nil
}
