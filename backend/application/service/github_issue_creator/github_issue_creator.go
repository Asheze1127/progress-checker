package githubissuecreator

import "context"

// CreateIssueInput holds the parameters for creating a GitHub issue.
type CreateIssueInput struct {
	Owner  string
	Repo   string
	Title  string
	Body   string
	Labels []string
}

// CreatedIssue holds the result of a successfully created GitHub issue.
type CreatedIssue struct {
	Number int
	URL    string
}

// GitHubIssueCreator creates GitHub issues.
type GitHubIssueCreator interface {
	CreateIssue(ctx context.Context, input CreateIssueInput) (*CreatedIssue, error)
}
