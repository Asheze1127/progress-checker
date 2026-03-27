package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do/v2"

	"github.com/Asheze1127/progress-checker/backend/application/port"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// GitHubService handles GitHub repository management and Issue creation.
type GitHubService struct {
	repo       port.GitHubRepoRepository
	encryptor  port.TokenEncryptor
	ghClient   port.GitHubIssueCreator
	idProvider func() string
}

// NewGitHubService creates a new GitHubService via DI container.
func NewGitHubService(i do.Injector) (*GitHubService, error) {
	repo := do.MustInvoke[port.GitHubRepoRepository](i)
	encryptor := do.MustInvoke[port.TokenEncryptor](i)
	ghClient := do.MustInvoke[port.GitHubIssueCreator](i)
	return &GitHubService{
		repo:       repo,
		encryptor:  encryptor,
		ghClient:   ghClient,
		idProvider: func() string { return uuid.New().String() },
	}, nil
}

// RegisterRepository registers a GitHub repository for a team.
func (s *GitHubService) RegisterRepository(ctx context.Context, teamID string, repoURL string, pat string) error {
	if strings.TrimSpace(teamID) == "" {
		return fmt.Errorf("team_id is required")
	}

	if strings.TrimSpace(pat) == "" {
		return fmt.Errorf("personal_access_token is required")
	}

	owner, repoName, err := entities.ParseGitHubRepoURL(repoURL)
	if err != nil {
		return err
	}

	encryptedToken, err := s.encryptor.Encrypt(pat)
	if err != nil {
		return fmt.Errorf("encrypt token: %w", err)
	}

	now := time.Now()
	ghRepo := &entities.GitHubRepo{
		ID:             entities.GitHubRepoID(s.idProvider()),
		TeamID:         entities.TeamID(teamID),
		Owner:          owner,
		RepoName:       repoName,
		EncryptedToken: encryptedToken,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Save(ctx, ghRepo); err != nil {
		return fmt.Errorf("save github repo: %w", err)
	}

	return nil
}

// ListRepositories returns all GitHub repositories registered for a team.
func (s *GitHubService) ListRepositories(ctx context.Context, teamID string) ([]entities.GitHubRepo, error) {
	if strings.TrimSpace(teamID) == "" {
		return nil, fmt.Errorf("team_id is required")
	}

	repos, err := s.repo.FindByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("find github repos: %w", err)
	}

	return repos, nil
}

// RemoveRepository removes a GitHub repository registration.
func (s *GitHubService) RemoveRepository(ctx context.Context, teamID string, repoID string) error {
	if strings.TrimSpace(teamID) == "" {
		return fmt.Errorf("team_id is required")
	}

	if strings.TrimSpace(repoID) == "" {
		return fmt.Errorf("repo_id is required")
	}

	existing, err := s.repo.FindByID(ctx, repoID)
	if err != nil {
		return fmt.Errorf("find github repo: %w", err)
	}

	if string(existing.TeamID) != teamID {
		return fmt.Errorf("github repo %s does not belong to team %s", repoID, teamID)
	}

	if err := s.repo.Delete(ctx, repoID); err != nil {
		return fmt.Errorf("delete github repo: %w", err)
	}

	return nil
}

// UpdateToken updates the personal access token for a registered GitHub repository.
func (s *GitHubService) UpdateToken(ctx context.Context, teamID string, repoID string, newPAT string) error {
	if strings.TrimSpace(teamID) == "" {
		return fmt.Errorf("team_id is required")
	}

	if strings.TrimSpace(repoID) == "" {
		return fmt.Errorf("repo_id is required")
	}

	if strings.TrimSpace(newPAT) == "" {
		return fmt.Errorf("personal_access_token is required")
	}

	existing, err := s.repo.FindByID(ctx, repoID)
	if err != nil {
		return fmt.Errorf("find github repo: %w", err)
	}

	if string(existing.TeamID) != teamID {
		return fmt.Errorf("github repo %s does not belong to team %s", repoID, teamID)
	}

	encryptedToken, err := s.encryptor.Encrypt(newPAT)
	if err != nil {
		return fmt.Errorf("encrypt token: %w", err)
	}

	if err := s.repo.UpdateToken(ctx, repoID, encryptedToken); err != nil {
		return fmt.Errorf("update token: %w", err)
	}

	return nil
}

// CreateIssue looks up the GitHub repository config by channel ID, decrypts the token,
// and creates a GitHub Issue.
func (s *GitHubService) CreateIssue(ctx context.Context, channelID string, title string, body string) (string, error) {
	if strings.TrimSpace(channelID) == "" {
		return "", fmt.Errorf("channel_id is required")
	}

	if strings.TrimSpace(title) == "" {
		return "", fmt.Errorf("title is required")
	}

	ghRepo, err := s.repo.FindByChannelID(ctx, channelID)
	if err != nil {
		return "", fmt.Errorf("find github repo by channel: %w", err)
	}

	token, err := s.encryptor.Decrypt(ghRepo.EncryptedToken)
	if err != nil {
		return "", fmt.Errorf("decrypt token: %w", err)
	}

	issueURL, err := s.ghClient.CreateIssue(ctx, ghRepo.Owner, ghRepo.RepoName, token, title, body)
	if err != nil {
		return "", fmt.Errorf("create github issue: %w", err)
	}

	return issueURL, nil
}
