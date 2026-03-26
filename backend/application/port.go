package application

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

// TokenEncryptor defines operations for encrypting and decrypting tokens.
type TokenEncryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// GitHubIssueCreator defines the operation for creating GitHub Issues via the GitHub API.
type GitHubIssueCreator interface {
	CreateIssue(ctx context.Context, owner, repo, token, title, body string) (issueURL string, err error)
}
