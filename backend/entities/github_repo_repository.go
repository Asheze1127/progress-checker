package entities

import "context"

// GitHubRepoID is the unique identifier for a team GitHub repository.
type GitHubRepoID string

// GitHubRepo represents a team's linked GitHub repository.
type GitHubRepo struct {
	ID           GitHubRepoID
	TeamID       TeamID
	GitHubRepoURL string
	EncryptedPAT string
}

// GitHubRepoRepository defines the interface for persisting and querying GitHub repositories.
type GitHubRepoRepository interface {
	Save(ctx context.Context, repo *GitHubRepo) (*GitHubRepo, error)
	GetByID(ctx context.Context, id GitHubRepoID) (*GitHubRepo, error)
	GetByTeamID(ctx context.Context, teamID TeamID) ([]*GitHubRepo, error)
	GetByChannelID(ctx context.Context, channelID SlackChannelID) (*GitHubRepo, error)
	Delete(ctx context.Context, id GitHubRepoID) error
	UpdateToken(ctx context.Context, id GitHubRepoID, encryptedPAT string) error
}
