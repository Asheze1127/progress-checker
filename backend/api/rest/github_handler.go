package rest

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/application"
)

// GitHubHandler handles REST API requests for GitHub repository management.
type GitHubHandler struct {
	service *application.GitHubService
}

// NewGitHubHandler creates a new GitHubHandler.
func NewGitHubHandler(service *application.GitHubService) *GitHubHandler {
	return &GitHubHandler{service: service}
}

type registerRepoRequest struct {
	GitHubRepoURL       string `json:"github_repo_url"`
	PersonalAccessToken string `json:"personal_access_token"`
}

type registerRepoResponse struct {
	Message string `json:"message"`
}

type listReposResponseItem struct {
	ID       string `json:"id"`
	Owner    string `json:"owner"`
	RepoName string `json:"repo_name"`
}

type listReposResponse struct {
	Repos []listReposResponseItem `json:"repos"`
}

type updateTokenRequest struct {
	PersonalAccessToken string `json:"personal_access_token"`
}

// RegisterRepository handles POST /api/v1/teams/:teamId/github-repos
func (h *GitHubHandler) RegisterRepository(w http.ResponseWriter, r *http.Request) {
	teamID := extractPathParam(r.URL.Path, "teams")
	if teamID == "" {
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "team_id is required in path"})
		return
	}

	var req registerRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if err := h.service.RegisterRepository(r.Context(), teamID, req.GitHubRepoURL, req.PersonalAccessToken); err != nil {
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusCreated, registerRepoResponse{Message: "repository registered successfully"})
}

// ListRepositories handles GET /api/v1/teams/:teamId/github-repos
func (h *GitHubHandler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	teamID := extractPathParam(r.URL.Path, "teams")
	if teamID == "" {
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "team_id is required in path"})
		return
	}

	repos, err := h.service.ListRepositories(r.Context(), teamID)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	items := make([]listReposResponseItem, 0, len(repos))
	for _, repo := range repos {
		items = append(items, listReposResponseItem{
			ID:       string(repo.ID),
			Owner:    repo.Owner,
			RepoName: repo.RepoName,
		})
	}

	WriteJSON(w, http.StatusOK, listReposResponse{Repos: items})
}

// RemoveRepository handles DELETE /api/v1/teams/:teamId/github-repos/:repoId
func (h *GitHubHandler) RemoveRepository(w http.ResponseWriter, r *http.Request) {
	teamID := extractPathParam(r.URL.Path, "teams")
	repoID := extractPathParam(r.URL.Path, "github-repos")

	if teamID == "" {
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "team_id is required in path"})
		return
	}
	if repoID == "" {
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "repo_id is required in path"})
		return
	}

	if err := h.service.RemoveRepository(r.Context(), teamID, repoID); err != nil {
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "repository removed successfully"})
}

// UpdateToken handles PUT /api/v1/teams/:teamId/github-repos/:repoId
func (h *GitHubHandler) UpdateToken(w http.ResponseWriter, r *http.Request) {
	teamID := extractPathParam(r.URL.Path, "teams")
	repoID := extractPathParam(r.URL.Path, "github-repos")

	if teamID == "" {
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "team_id is required in path"})
		return
	}
	if repoID == "" {
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "repo_id is required in path"})
		return
	}

	var req updateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if err := h.service.UpdateToken(r.Context(), teamID, repoID, req.PersonalAccessToken); err != nil {
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"message": "token updated successfully"})
}

// extractPathParam extracts the value after the given segment in a URL path.
// For example, extractPathParam("/api/v1/teams/team-1/github-repos", "teams") returns "team-1".
func extractPathParam(urlPath string, segment string) string {
	parts := strings.Split(strings.Trim(urlPath, "/"), "/")
	for i, part := range parts {
		if part == segment && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
