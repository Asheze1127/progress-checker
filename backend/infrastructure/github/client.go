package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-github/v84/github"

	githubissuecreator "github.com/Asheze1127/progress-checker/backend/application/service/github_issue_creator"
)

// Compile-time check that Client implements githubissuecreator.GitHubIssueCreator.
var _ githubissuecreator.GitHubIssueCreator = (*Client)(nil)

// Client interacts with the GitHub API using the google/go-github library.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new GitHub API client.
// baseURL is optional; pass empty string to use the default GitHub API.
func NewClient(baseURL string) *Client {
	normalized, _ := normalizeBaseURL(baseURL)
	return &Client{
		baseURL: normalized,
	}
}

// CreateIssue creates a GitHub Issue and returns the Issue URL.
func (c *Client) CreateIssue(ctx context.Context, owner, repo, token, title, body string) (string, error) {
	if strings.TrimSpace(owner) == "" {
		return "", fmt.Errorf("owner is required")
	}
	if strings.TrimSpace(repo) == "" {
		return "", fmt.Errorf("repo is required")
	}
	if strings.TrimSpace(token) == "" {
		return "", fmt.Errorf("token is required")
	}
	if strings.TrimSpace(title) == "" {
		return "", fmt.Errorf("title is required")
	}

	ghClient, err := c.newGitHubClient(token)
	if err != nil {
		return "", err
	}

	issueRequest := &github.IssueRequest{
		Title: github.Ptr(title),
	}
	if body != "" {
		issueRequest.Body = github.Ptr(body)
	}

	issue, _, err := ghClient.Issues.Create(ctx, owner, repo, issueRequest)
	if err != nil {
		return "", fmt.Errorf("create github issue: %w", err)
	}

	issueURL := issue.GetHTMLURL()
	if issueURL == "" {
		return "", fmt.Errorf("github api returned empty issue url")
	}

	return issueURL, nil
}

func (c *Client) newGitHubClient(token string) (*github.Client, error) {
	ghClient := github.NewClient(c.httpClient).WithAuthToken(token)

	if c.baseURL == "" {
		return ghClient, nil
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse github base url: %w", err)
	}

	ghClient.BaseURL = baseURL
	ghClient.UploadURL = baseURL

	return ghClient, nil
}

func normalizeBaseURL(rawURL string) (string, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return "", nil
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("parse github base url: %w", err)
	}

	if !strings.HasSuffix(parsed.Path, "/") {
		parsed.Path += "/"
	}

	return parsed.String(), nil
}
