package entities

import (
  "errors"
  "fmt"
  "net/url"
  "strings"
  "time"
)

type GitHubRepoID string

type GitHubRepo struct {
  ID             GitHubRepoID
  TeamID         TeamID
  Owner          string
  RepoName       string
  EncryptedToken string
  CreatedAt      time.Time
  UpdatedAt      time.Time
}

func (g GitHubRepo) Validate() error {
  var errs []error

  if strings.TrimSpace(string(g.ID)) == "" {
    errs = append(errs, fmt.Errorf("github_repo.id is required"))
  }

  if strings.TrimSpace(string(g.TeamID)) == "" {
    errs = append(errs, fmt.Errorf("github_repo.team_id is required"))
  }

  if strings.TrimSpace(g.Owner) == "" {
    errs = append(errs, fmt.Errorf("github_repo.owner is required"))
  }

  if strings.TrimSpace(g.RepoName) == "" {
    errs = append(errs, fmt.Errorf("github_repo.repo_name is required"))
  }

  if strings.TrimSpace(g.EncryptedToken) == "" {
    errs = append(errs, fmt.Errorf("github_repo.encrypted_token is required"))
  }

  return errors.Join(errs...)
}

// ParseGitHubRepoURL extracts owner and repo name from a GitHub repository URL.
// Accepted formats: https://github.com/{owner}/{repo} or https://github.com/{owner}/{repo}.git
func ParseGitHubRepoURL(rawURL string) (owner string, repoName string, err error) {
  trimmed := strings.TrimSpace(rawURL)
  if trimmed == "" {
    return "", "", fmt.Errorf("github_repo_url is required")
  }

  parsed, err := url.Parse(trimmed)
  if err != nil {
    return "", "", fmt.Errorf("invalid github_repo_url: %w", err)
  }

  if parsed.Scheme != "https" {
    return "", "", fmt.Errorf("github_repo_url must use https scheme")
  }

  if !strings.EqualFold(parsed.Host, "github.com") {
    return "", "", fmt.Errorf("github_repo_url must be a github.com URL")
  }

  path := strings.Trim(parsed.Path, "/")
  path = strings.TrimSuffix(path, ".git")
  parts := strings.Split(path, "/")

  if len(parts) != 2 {
    return "", "", fmt.Errorf("github_repo_url must be in format https://github.com/{owner}/{repo}")
  }

  owner = strings.TrimSpace(parts[0])
  repoName = strings.TrimSpace(parts[1])

  if owner == "" || repoName == "" {
    return "", "", fmt.Errorf("github_repo_url must contain a valid owner and repo name")
  }

  return owner, repoName, nil
}
