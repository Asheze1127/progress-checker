package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	githubsvc "github.com/Asheze1127/progress-checker/backend/application/service/github"
)

// GitHubHandler handles REST API requests for GitHub repository management.
type GitHubHandler struct {
	service *githubsvc.GitHubService
}

// NewGitHubHandler creates a new GitHubHandler.
func NewGitHubHandler(service *githubsvc.GitHubService) *GitHubHandler {
	return &GitHubHandler{service: service}
}

// RegisterRepository handles POST /api/v1/teams/{teamId}/github-repos.
func (h *GitHubHandler) RegisterRepository(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("teamId")
	if teamID == "" {
		WriteError(w, http.StatusBadRequest, "team_id is required in path")
		return
	}

	var req openapi.RegisterRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.RegisterRepository(r.Context(), teamID, req.GithubRepoUrl, req.PersonalAccessToken); err != nil {
		slog.Error("failed to register repository", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "failed to register repository")
		return
	}

	WriteJSON(w, http.StatusCreated, openapi.MessageResponse{Message: "repository registered successfully"})
}

// ListRepositories handles GET /api/v1/teams/{teamId}/github-repos.
func (h *GitHubHandler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("teamId")
	if teamID == "" {
		WriteError(w, http.StatusBadRequest, "team_id is required in path")
		return
	}

	repos, err := h.service.ListRepositories(r.Context(), teamID)
	if err != nil {
		slog.Error("failed to list repositories", slog.String("error", err.Error()))
		WriteError(w, http.StatusInternalServerError, "failed to list repositories")
		return
	}

	items := make([]openapi.RepoItem, 0, len(repos))
	for _, repo := range repos {
		items = append(items, openapi.RepoItem{
			Id:       string(repo.ID),
			Owner:    repo.Owner,
			RepoName: repo.RepoName,
		})
	}

	WriteJSON(w, http.StatusOK, openapi.ListReposResponse{Repos: items})
}

// RemoveRepository handles DELETE /api/v1/teams/{teamId}/github-repos/{repoId}.
func (h *GitHubHandler) RemoveRepository(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("teamId")
	repoID := r.PathValue("repoId")

	if teamID == "" {
		WriteError(w, http.StatusBadRequest, "team_id is required in path")
		return
	}
	if repoID == "" {
		WriteError(w, http.StatusBadRequest, "repo_id is required in path")
		return
	}

	if err := h.service.RemoveRepository(r.Context(), teamID, repoID); err != nil {
		slog.Error("failed to remove repository", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "failed to remove repository")
		return
	}

	WriteJSON(w, http.StatusOK, openapi.MessageResponse{Message: "repository removed successfully"})
}

// UpdateToken handles PUT /api/v1/teams/{teamId}/github-repos/{repoId}/token.
func (h *GitHubHandler) UpdateToken(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("teamId")
	repoID := r.PathValue("repoId")

	if teamID == "" {
		WriteError(w, http.StatusBadRequest, "team_id is required in path")
		return
	}
	if repoID == "" {
		WriteError(w, http.StatusBadRequest, "repo_id is required in path")
		return
	}

	var req openapi.UpdateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.UpdateToken(r.Context(), teamID, repoID, req.PersonalAccessToken); err != nil {
		slog.Error("failed to update token", slog.String("error", err.Error()))
		WriteError(w, http.StatusBadRequest, "failed to update token")
		return
	}

	WriteJSON(w, http.StatusOK, openapi.MessageResponse{Message: "token updated successfully"})
}
