package rest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/api/rest"
	"github.com/Asheze1127/progress-checker/backend/application"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

func newTestInternalHandler() *rest.InternalHandler {
	repo := newStubRepo()
	repo.byChannel["C123"] = &entities.GitHubRepo{
		ID:             "repo-1",
		TeamID:         "team-1",
		Owner:          "org",
		RepoName:       "repo",
		EncryptedToken: "enc:ghp_test",
	}

	idCounter := 0
	svc := application.NewGitHubService(repo, &stubEncryptor{}, &stubGitHubIssueCreator{}, func() string {
		idCounter++
		return fmt.Sprintf("repo-%d", idCounter)
	})

	return rest.NewInternalHandler(svc)
}

func TestInternalHandler_CreateIssue(t *testing.T) {
	t.Parallel()

	handler := newTestInternalHandler()

	body := `{"channel_id":"C123","title":"Bug fix","body":"Fix the bug"}`
	req := httptest.NewRequest(http.MethodPost, "/internal/issues", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.CreateIssue(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp struct {
		IssueURL string `json:"issue_url"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.IssueURL != "https://github.com/org/repo/issues/1" {
		t.Errorf("unexpected issue_url: %s", resp.IssueURL)
	}
}

func TestInternalHandler_CreateIssue_InvalidBody(t *testing.T) {
	t.Parallel()

	handler := newTestInternalHandler()

	req := httptest.NewRequest(http.MethodPost, "/internal/issues", strings.NewReader("invalid"))
	rec := httptest.NewRecorder()

	handler.CreateIssue(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestInternalHandler_CreateIssue_MissingChannelID(t *testing.T) {
	t.Parallel()

	handler := newTestInternalHandler()

	body := `{"channel_id":"","title":"Bug fix","body":"Fix the bug"}`
	req := httptest.NewRequest(http.MethodPost, "/internal/issues", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.CreateIssue(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestInternalHandler_CreateIssue_UnknownChannel(t *testing.T) {
	t.Parallel()

	// Use a repo with no channel mappings
	repo := newStubRepo()
	svc := application.NewGitHubService(repo, &stubEncryptor{}, &stubGitHubIssueCreator{}, func() string { return "id" })
	handler := rest.NewInternalHandler(svc)

	body := `{"channel_id":"C_UNKNOWN","title":"Bug fix","body":"Fix the bug"}`
	req := httptest.NewRequest(http.MethodPost, "/internal/issues", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.CreateIssue(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// Ensure the stub implements the interface at compile time.
var _ application.GitHubRepoRepository = (*stubGitHubRepoRepository)(nil)
var _ application.TokenEncryptor = (*stubEncryptor)(nil)
var _ application.GitHubIssueCreator = (*stubGitHubIssueCreator)(nil)

// Suppress unused import warning for context in this file.
var _ = context.Background
