package githubissuecreator

import "context"

type GitHubIssueCreator interface {
	CreateIssue(ctx context.Context, owner, repo, token, title, body string) (issueURL string, err error)
}
