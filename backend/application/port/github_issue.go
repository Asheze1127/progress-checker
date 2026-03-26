package port

import "context"

// GitHubIssueCreator defines the operation for creating GitHub Issues via the GitHub API.
type GitHubIssueCreator interface {
	CreateIssue(ctx context.Context, owner, repo, token, title, body string) (issueURL string, err error)
}
