package postgres

import (
	"context"
	"database/sql"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

var _ entities.GitHubRepoRepository = (*GitHubRepoRepository)(nil)

// GitHubRepoRepository persists and queries GitHub repositories in PostgreSQL.
type GitHubRepoRepository struct {
	queries *db.Queries
}

// NewGitHubRepoRepository creates a new GitHubRepoRepository backed by the given database connection.
func NewGitHubRepoRepository(database *sql.DB) *GitHubRepoRepository {
	return &GitHubRepoRepository{queries: db.New(database)}
}

func (r *GitHubRepoRepository) Save(ctx context.Context, repo *entities.GitHubRepo) (*entities.GitHubRepo, error) {
	repoID, err := uuid.Parse(string(repo.ID))
	if err != nil {
		return nil, err
	}
	teamID, err := uuid.Parse(string(repo.TeamID))
	if err != nil {
		return nil, err
	}
	row, err := r.queries.InsertGitHubRepo(ctx, db.InsertGitHubRepoParams{
		ID:            repoID,
		TeamID:        teamID,
		GithubRepoUrl: repo.GitHubRepoURL,
		EncryptedPat:  repo.EncryptedPAT,
	})
	if err != nil {
		return nil, err
	}
	return toGitHubRepoEntity(row), nil
}

func (r *GitHubRepoRepository) GetByID(ctx context.Context, id entities.GitHubRepoID) (*entities.GitHubRepo, error) {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return nil, err
	}
	row, err := r.queries.GetGitHubRepoByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return toGitHubRepoEntity(row), nil
}

func (r *GitHubRepoRepository) GetByTeamID(ctx context.Context, teamID entities.TeamID) ([]*entities.GitHubRepo, error) {
	uid, err := uuid.Parse(string(teamID))
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.GetGitHubReposByTeamID(ctx, uid)
	if err != nil {
		return nil, err
	}
	repos := make([]*entities.GitHubRepo, len(rows))
	for i, row := range rows {
		repos[i] = toGitHubRepoEntity(row)
	}
	return repos, nil
}

func (r *GitHubRepoRepository) GetByChannelID(ctx context.Context, channelID entities.SlackChannelID) (*entities.GitHubRepo, error) {
	row, err := r.queries.GetGitHubRepoByChannelID(ctx, string(channelID))
	if err != nil {
		return nil, err
	}
	return toGitHubRepoEntity(row), nil
}

func (r *GitHubRepoRepository) Delete(ctx context.Context, id entities.GitHubRepoID) error {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return err
	}
	return r.queries.DeleteGitHubRepo(ctx, uid)
}

func (r *GitHubRepoRepository) UpdateToken(ctx context.Context, id entities.GitHubRepoID, encryptedPAT string) error {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return err
	}
	return r.queries.UpdateGitHubRepoToken(ctx, db.UpdateGitHubRepoTokenParams{
		EncryptedPat: encryptedPAT,
		ID:           uid,
	})
}

func toGitHubRepoEntity(row db.TeamGithubRepositories) *entities.GitHubRepo {
	return &entities.GitHubRepo{
		ID:            entities.GitHubRepoID(row.ID.String()),
		TeamID:        entities.TeamID(row.TeamID.String()),
		GitHubRepoURL: row.GithubRepoUrl,
		EncryptedPAT:  row.EncryptedPat,
	}
}
