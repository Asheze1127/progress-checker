package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/samber/do/v2"

	ghsvc "github.com/Asheze1127/progress-checker/backend/application/service"
)

// InternalHandler handles internal API requests (accessible via Internal ALB).
type InternalHandler struct {
	service *ghsvc.GitHubService
}

// NewInternalHandler creates a new InternalHandler via DI container.
func NewInternalHandler(i do.Injector) (*InternalHandler, error) {
	svc := do.MustInvoke[*ghsvc.GitHubService](i)
	return &InternalHandler{service: svc}, nil
}

type createIssueRequest struct {
	ChannelID string `json:"channel_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

type createIssueResponse struct {
	IssueURL string `json:"issue_url"`
}

// CreateIssue handles POST /internal/issues
// Looks up team by channel_id, retrieves the GitHub repo config, and creates an Issue.
func (h *InternalHandler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	var req createIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	issueURL, err := h.service.CreateIssue(r.Context(), req.ChannelID, req.Title, req.Body)
	if err != nil {
		slog.Error("failed to create issue", slog.String("error", err.Error()))
		WriteError(w, http.StatusInternalServerError, "failed to create issue")
		return
	}

	WriteJSON(w, http.StatusCreated, createIssueResponse{IssueURL: issueURL})
}
