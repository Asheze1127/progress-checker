package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	githubsvc "github.com/Asheze1127/progress-checker/backend/application/service/github"
)

// GitHubHandler handles GitHub repository management endpoints.
type GitHubHandler struct {
	service *githubsvc.GitHubService
}

// NewGitHubHandler creates a new GitHubHandler.
func NewGitHubHandler(service *githubsvc.GitHubService) *GitHubHandler {
	return &GitHubHandler{service: service}
}

// isAuthError returns true if the error is an authorization failure.
// NOTE: Ideally these should map to 401/403 responses, but the OpenAPI spec
// does not define those response types for GitHub endpoints yet. Using 400
// until the spec is updated and the server code is regenerated.
func isAuthError(err error) bool {
	return errors.Is(err, githubsvc.ErrGitHubNotAuthorized) || errors.Is(err, githubsvc.ErrGitHubNotAuthorizedForTeam)
}

// ListRepositories handles GET /api/v1/teams/{teamId}/github-repos.
func (h *GitHubHandler) ListRepositories(ctx context.Context, request openapi.ListRepositoriesRequestObject) (openapi.ListRepositoriesResponseObject, error) {
	repos, err := h.service.ListRepositories(ctx, request.TeamId)
	if err != nil {
		if isAuthError(err) {
			return openapi.ListRepositories400JSONResponse{Error: "not authorized for this team"}, nil
		}
		slog.Error("failed to list repositories", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	items := make([]openapi.RepoItem, 0, len(repos))
	for _, repo := range repos {
		items = append(items, openapi.RepoItem{
			Id:       string(repo.ID),
			Owner:    repo.Owner,
			RepoName: repo.RepoName,
		})
	}

	return openapi.ListRepositories200JSONResponse{Repos: items}, nil
}

// RegisterRepository handles POST /api/v1/teams/{teamId}/github-repos.
func (h *GitHubHandler) RegisterRepository(ctx context.Context, request openapi.RegisterRepositoryRequestObject) (openapi.RegisterRepositoryResponseObject, error) {
	if err := h.service.RegisterRepository(ctx, request.TeamId, request.Body.GithubRepoUrl, request.Body.PersonalAccessToken); err != nil {
		if isAuthError(err) {
			return openapi.RegisterRepository400JSONResponse{Error: "not authorized for this team"}, nil
		}
		slog.Error("failed to register repository", slog.String("error", err.Error()))
		return openapi.RegisterRepository400JSONResponse{Error: "failed to register repository"}, nil
	}

	return openapi.RegisterRepository201JSONResponse{Message: "repository registered successfully"}, nil
}

// RemoveRepository handles DELETE /api/v1/teams/{teamId}/github-repos/{repoId}.
func (h *GitHubHandler) RemoveRepository(ctx context.Context, request openapi.RemoveRepositoryRequestObject) (openapi.RemoveRepositoryResponseObject, error) {
	if err := h.service.RemoveRepository(ctx, request.TeamId, request.RepoId); err != nil {
		if isAuthError(err) {
			return openapi.RemoveRepository400JSONResponse{Error: "not authorized for this team"}, nil
		}
		slog.Error("failed to remove repository", slog.String("error", err.Error()))
		return openapi.RemoveRepository400JSONResponse{Error: "failed to remove repository"}, nil
	}

	return openapi.RemoveRepository200JSONResponse{Message: "repository removed successfully"}, nil
}

// UpdateToken handles PUT /api/v1/teams/{teamId}/github-repos/{repoId}/token.
func (h *GitHubHandler) UpdateToken(ctx context.Context, request openapi.UpdateTokenRequestObject) (openapi.UpdateTokenResponseObject, error) {
	if err := h.service.UpdateToken(ctx, request.TeamId, request.RepoId, request.Body.PersonalAccessToken); err != nil {
		if isAuthError(err) {
			return openapi.UpdateToken400JSONResponse{Error: "not authorized for this team"}, nil
		}
		slog.Error("failed to update token", slog.String("error", err.Error()))
		return openapi.UpdateToken400JSONResponse{Error: "failed to update token"}, nil
	}

	return openapi.UpdateToken200JSONResponse{Message: "token updated successfully"}, nil
}
