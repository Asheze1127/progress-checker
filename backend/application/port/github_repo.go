package port

import (
	"context"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// GitHubRepoRepository defines the persistence operations for GitHub repository configurations.
type GitHubRepoRepository interface {
	Save(ctx context.Context, repo *entities.GitHubRepo) error
	FindByTeamID(ctx context.Context, teamID string) ([]entities.GitHubRepo, error)
	FindByID(ctx context.Context, repoID string) (*entities.GitHubRepo, error)
	FindByChannelID(ctx context.Context, channelID string) (*entities.GitHubRepo, error)
	Delete(ctx context.Context, repoID string) error
	UpdateToken(ctx context.Context, repoID string, encryptedToken string) error
}
