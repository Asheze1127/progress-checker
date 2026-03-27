package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	githubsvc "github.com/Asheze1127/progress-checker/backend/application/service/github"
)

// InternalHandler handles internal API requests (accessible via Internal ALB).
type InternalHandler struct {
	service *githubsvc.GitHubService
}

// NewInternalHandler creates a new InternalHandler.
func NewInternalHandler(service *githubsvc.GitHubService) *InternalHandler {
	return &InternalHandler{service: service}
}

// CreateIssue handles POST /internal/issues.
// Looks up team by channel_id, retrieves the GitHub repo config, and creates an Issue.
func (h *InternalHandler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, openapi.ErrorResponse{Error: "invalid request body"})
		return
	}

	issueURL, err := h.service.CreateIssue(r.Context(), req.ChannelId, req.Title, req.Body)
	if err != nil {
		slog.Error("failed to create issue", slog.String("error", err.Error()))
		WriteError(w, http.StatusInternalServerError, "failed to create issue")
		return
	}

	WriteJSON(w, http.StatusCreated, openapi.CreateIssueResponse{IssueUrl: issueURL})
}
