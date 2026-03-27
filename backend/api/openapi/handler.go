package openapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// APIHandler implements the generated ServerInterface.
type APIHandler struct {
	loginUC        *usecase.LoginUseCase
	listProgressUC *usecase.ListProgressUseCase
	ghService      *service.GitHubService
	corsOrigins    []string
}

// NewAPIHandler creates a new APIHandler with the given dependencies.
func NewAPIHandler(
	loginUC *usecase.LoginUseCase,
	listProgressUC *usecase.ListProgressUseCase,
	ghService *service.GitHubService,
	corsOrigins []string,
) *APIHandler {
	return &APIHandler{
		loginUC:        loginUC,
		listProgressUC: listProgressUC,
		ghService:      ghService,
		corsOrigins:    corsOrigins,
	}
}

// Compile-time check that APIHandler implements ServerInterface.
var _ ServerInterface = (*APIHandler)(nil)

// Login handles POST /api/v1/auth/login.
func (h *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.loginUC.Execute(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		if errors.Is(err, usecase.ErrUserNotMentor) {
			writeError(w, http.StatusForbidden, "only mentors can log in")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, LoginResponse{
		Token: result.Token,
		User:  toUserResponse(result.User),
	})
}

// ListProgress handles GET /api/v1/progress.
func (h *APIHandler) ListProgress(w http.ResponseWriter, r *http.Request, params ListProgressParams) {
	h.setCORSHeaders(w, r)

	var teamID string
	if params.TeamId != nil {
		teamID = *params.TeamId
	}

	results, err := h.listProgressUC.Execute(r.Context(), teamID)
	if err != nil {
		slog.Error("failed to list latest progress", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, toProgressListResponse(results))
}

// RegisterRepository handles POST /api/v1/teams/{teamId}/github-repos.
func (h *APIHandler) RegisterRepository(w http.ResponseWriter, r *http.Request, teamId string) {
	if teamId == "" {
		writeError(w, http.StatusBadRequest, "team_id is required in path")
		return
	}

	var req RegisterRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.ghService.RegisterRepository(r.Context(), teamId, req.GithubRepoUrl, req.PersonalAccessToken); err != nil {
		slog.Error("failed to register repository", slog.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, "failed to register repository")
		return
	}

	writeJSON(w, http.StatusCreated, MessageResponse{Message: "repository registered successfully"})
}

// ListRepositories handles GET /api/v1/teams/{teamId}/github-repos.
func (h *APIHandler) ListRepositories(w http.ResponseWriter, r *http.Request, teamId string) {
	if teamId == "" {
		writeError(w, http.StatusBadRequest, "team_id is required in path")
		return
	}

	repos, err := h.ghService.ListRepositories(r.Context(), teamId)
	if err != nil {
		slog.Error("failed to list repositories", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed to list repositories")
		return
	}

	items := make([]RepoItem, 0, len(repos))
	for _, repo := range repos {
		items = append(items, RepoItem{
			Id:       string(repo.ID),
			Owner:    repo.Owner,
			RepoName: repo.RepoName,
		})
	}

	writeJSON(w, http.StatusOK, ListReposResponse{Repos: items})
}

// RemoveRepository handles DELETE /api/v1/teams/{teamId}/github-repos/{repoId}.
func (h *APIHandler) RemoveRepository(w http.ResponseWriter, r *http.Request, teamId string, repoId string) {
	if teamId == "" {
		writeError(w, http.StatusBadRequest, "team_id is required in path")
		return
	}
	if repoId == "" {
		writeError(w, http.StatusBadRequest, "repo_id is required in path")
		return
	}

	if err := h.ghService.RemoveRepository(r.Context(), teamId, repoId); err != nil {
		slog.Error("failed to remove repository", slog.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, "failed to remove repository")
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "repository removed successfully"})
}

// UpdateToken handles PUT /api/v1/teams/{teamId}/github-repos/{repoId}/token.
func (h *APIHandler) UpdateToken(w http.ResponseWriter, r *http.Request, teamId string, repoId string) {
	if teamId == "" {
		writeError(w, http.StatusBadRequest, "team_id is required in path")
		return
	}
	if repoId == "" {
		writeError(w, http.StatusBadRequest, "repo_id is required in path")
		return
	}

	var req UpdateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.ghService.UpdateToken(r.Context(), teamId, repoId, req.PersonalAccessToken); err != nil {
		slog.Error("failed to update token", slog.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, "failed to update token")
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "token updated successfully"})
}

// CreateIssue handles POST /internal/issues.
func (h *APIHandler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	var req CreateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	issueURL, err := h.ghService.CreateIssue(r.Context(), req.ChannelId, req.Title, req.Body)
	if err != nil {
		slog.Error("failed to create issue", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed to create issue")
		return
	}

	writeJSON(w, http.StatusCreated, CreateIssueResponse{IssueUrl: issueURL})
}

// HandleProgressPreflight handles CORS preflight requests for the progress endpoint.
func (h *APIHandler) HandleProgressPreflight(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w, r)
	w.WriteHeader(http.StatusNoContent)
}

// setCORSHeaders adds CORS headers to the response for cross-origin access.
func (h *APIHandler) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}

	for _, allowed := range h.corsOrigins {
		if origin == allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			return
		}
	}
}

// toUserResponse converts an entity User to a UserResponse.
func toUserResponse(user entities.User) UserResponse {
	return UserResponse{
		Id:   string(user.ID),
		Name: user.Name,
		Role: string(user.Role),
	}
}

// toProgressListResponse converts domain objects to the API response format.
func toProgressListResponse(teamProgresses []entities.TeamProgress) ProgressListResponse {
	data := make([]TeamProgressResponse, 0, len(teamProgresses))

	for _, tp := range teamProgresses {
		item := TeamProgressResponse{
			TeamId:   string(tp.TeamID),
			TeamName: tp.TeamName,
		}

		if tp.LatestProgress != nil {
			item.LatestProgress = toLatestProgressResponse(tp.LatestProgress)
		}

		data = append(data, item)
	}

	return ProgressListResponse{Data: data}
}

// toLatestProgressResponse converts a ProgressLog entity to its API response format.
func toLatestProgressResponse(log *entities.ProgressLog) *LatestProgressResponse {
	bodies := make([]ProgressBodyResponse, 0, len(log.ProgressBodies))
	for _, b := range log.ProgressBodies {
		bodies = append(bodies, ProgressBodyResponse{
			Phase:       string(b.Phase),
			Sos:         b.SOS,
			Comment:     b.Comment,
			SubmittedAt: b.SubmittedAt.UTC().Truncate(time.Second),
		})
	}

	return &LatestProgressResponse{
		Id:             string(log.ID),
		ParticipantId:  string(log.ParticipantID),
		ProgressBodies: bodies,
	}
}

// writeJSON writes a JSON response with the given status code and payload.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON response", slog.String("error", err.Error()))
	}
}

// writeError writes a JSON error response with the given status code and message.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
