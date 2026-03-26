package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	githubAPIBaseURL     = "https://api.github.com"
	defaultClientTimeout = 30 * time.Second
)

// Client interacts with the GitHub REST API to create Issues.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new GitHub API client.
// An optional baseURL can be provided for testing; pass empty string for production.
func NewClient(baseURL string) *Client {
	effectiveBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if effectiveBaseURL == "" {
		effectiveBaseURL = githubAPIBaseURL
	}

	return &Client{
		httpClient: &http.Client{Timeout: defaultClientTimeout},
		baseURL:    effectiveBaseURL,
	}
}

type createIssueRequest struct {
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
}

type createIssueResponse struct {
	HTMLURL string `json:"html_url"`
}

// CreateIssue creates a GitHub Issue in the specified repository and returns the Issue URL.
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

	requestURL := fmt.Sprintf("%s/repos/%s/%s/issues", c.baseURL, owner, repo)

	reqBody := createIssueRequest{
		Title: title,
		Body:  body,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("github api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var issueResp createIssueResponse
	if err := json.Unmarshal(respBody, &issueResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if issueResp.HTMLURL == "" {
		return "", fmt.Errorf("github api returned empty issue url")
	}

	return issueResp.HTMLURL, nil
}
