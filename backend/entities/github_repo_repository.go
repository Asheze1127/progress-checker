package entities

import "context"

// GitHubRepoRepository defines the persistence operations for GitHub repository configurations.
type GitHubRepoRepository interface {
	Save(ctx context.Context, repo *GitHubRepo) error
	FindByTeamID(ctx context.Context, teamID string) ([]GitHubRepo, error)
	FindByID(ctx context.Context, repoID string) (*GitHubRepo, error)
	FindByChannelID(ctx context.Context, channelID string) (*GitHubRepo, error)
	Delete(ctx context.Context, repoID string) error
	UpdateToken(ctx context.Context, repoID string, encryptedToken string) error
}
