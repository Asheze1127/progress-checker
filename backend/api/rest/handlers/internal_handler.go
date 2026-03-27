package handlers

import (
	"context"
	"log/slog"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	githubsvc "github.com/Asheze1127/progress-checker/backend/application/service/github"
)

// InternalHandler handles internal API endpoints.
type InternalHandler struct {
	service *githubsvc.GitHubService
}

// NewInternalHandler creates a new InternalHandler.
func NewInternalHandler(service *githubsvc.GitHubService) *InternalHandler {
	return &InternalHandler{service: service}
}

// CreateIssue handles POST /internal/issues.
func (h *InternalHandler) CreateIssue(ctx context.Context, request openapi.CreateIssueRequestObject) (openapi.CreateIssueResponseObject, error) {
	issueURL, err := h.service.CreateIssue(ctx, request.Body.ChannelId, request.Body.Title, request.Body.Body)
	if err != nil {
		slog.Error("failed to create issue", slog.String("error", err.Error()))
		return nil, err
	}

	return openapi.CreateIssue201JSONResponse{IssueUrl: issueURL}, nil
}
