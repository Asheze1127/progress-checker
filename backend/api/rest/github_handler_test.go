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

// --- Test doubles ---

type stubGitHubRepoRepository struct {
	saved     []*entities.GitHubRepo
	repos     map[string][]entities.GitHubRepo
	byID      map[string]*entities.GitHubRepo
	byChannel map[string]*entities.GitHubRepo
}

func newStubRepo() *stubGitHubRepoRepository {
	return &stubGitHubRepoRepository{
		repos:     make(map[string][]entities.GitHubRepo),
		byID:      make(map[string]*entities.GitHubRepo),
		byChannel: make(map[string]*entities.GitHubRepo),
	}
}

func (s *stubGitHubRepoRepository) Save(_ context.Context, repo *entities.GitHubRepo) error {
	s.saved = append(s.saved, repo)
	s.byID[string(repo.ID)] = repo
	teamID := string(repo.TeamID)
	s.repos[teamID] = append(s.repos[teamID], *repo)
	return nil
}

func (s *stubGitHubRepoRepository) FindByTeamID(_ context.Context, teamID string) ([]entities.GitHubRepo, error) {
	return s.repos[teamID], nil
}

func (s *stubGitHubRepoRepository) FindByID(_ context.Context, repoID string) (*entities.GitHubRepo, error) {
	repo, ok := s.byID[repoID]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return repo, nil
}

func (s *stubGitHubRepoRepository) FindByChannelID(_ context.Context, channelID string) (*entities.GitHubRepo, error) {
	repo, ok := s.byChannel[channelID]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return repo, nil
}

func (s *stubGitHubRepoRepository) Delete(_ context.Context, repoID string) error {
	delete(s.byID, repoID)
	return nil
}

func (s *stubGitHubRepoRepository) UpdateToken(_ context.Context, repoID string, encryptedToken string) error {
	repo, ok := s.byID[repoID]
	if !ok {
		return fmt.Errorf("not found")
	}
	repo.EncryptedToken = encryptedToken
	return nil
}

type stubEncryptor struct{}

func (s *stubEncryptor) Encrypt(plaintext string) (string, error) { return "enc:" + plaintext, nil }
func (s *stubEncryptor) Decrypt(ciphertext string) (string, error) {
	return ciphertext[4:], nil
}

type stubGitHubIssueCreator struct{}

func (s *stubGitHubIssueCreator) CreateIssue(_ context.Context, _, _, _, _, _ string) (string, error) {
	return "https://github.com/org/repo/issues/1", nil
}

func newTestHandler() (*rest.GitHubHandler, *stubGitHubRepoRepository) {
	repo := newStubRepo()
	idCounter := 0
	svc := application.NewGitHubService(repo, &stubEncryptor{}, &stubGitHubIssueCreator{}, func() string {
		idCounter++
		return fmt.Sprintf("repo-%d", idCounter)
	})
	return rest.NewGitHubHandler(svc), repo
}

// --- Tests ---

func TestGitHubHandler_RegisterRepository(t *testing.T) {
	t.Parallel()

	handler, _ := newTestHandler()

	body := `{"github_repo_url":"https://github.com/org/repo","personal_access_token":"ghp_test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/team-1/github-repos", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.RegisterRepository(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["message"] != "repository registered successfully" {
		t.Errorf("unexpected response message: %s", resp["message"])
	}
}

func TestGitHubHandler_RegisterRepository_InvalidBody(t *testing.T) {
	t.Parallel()

	handler, _ := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/team-1/github-repos", strings.NewReader("invalid"))
	rec := httptest.NewRecorder()

	handler.RegisterRepository(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGitHubHandler_RegisterRepository_MissingTeamID(t *testing.T) {
	t.Parallel()

	handler, _ := newTestHandler()

	body := `{"github_repo_url":"https://github.com/org/repo","personal_access_token":"ghp_test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams//github-repos", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.RegisterRepository(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGitHubHandler_ListRepositories(t *testing.T) {
	t.Parallel()

	handler, _ := newTestHandler()

	// Register a repo first
	body := `{"github_repo_url":"https://github.com/org/repo","personal_access_token":"ghp_test"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams/team-1/github-repos", strings.NewReader(body))
	regRec := httptest.NewRecorder()
	handler.RegisterRepository(regRec, regReq)

	// List repos
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/team-1/github-repos", nil)
	rec := httptest.NewRecorder()

	handler.ListRepositories(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp struct {
		Repos []struct {
			ID       string `json:"id"`
			Owner    string `json:"owner"`
			RepoName string `json:"repo_name"`
		} `json:"repos"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(resp.Repos))
	}

	if resp.Repos[0].Owner != "org" {
		t.Errorf("expected owner 'org', got %s", resp.Repos[0].Owner)
	}
}

func TestGitHubHandler_RemoveRepository(t *testing.T) {
	t.Parallel()

	handler, repo := newTestHandler()

	// Register a repo first
	body := `{"github_repo_url":"https://github.com/org/repo","personal_access_token":"ghp_test"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams/team-1/github-repos", strings.NewReader(body))
	regRec := httptest.NewRecorder()
	handler.RegisterRepository(regRec, regReq)

	repoID := string(repo.saved[0].ID)

	// Remove repo
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/team-1/github-repos/"+repoID, nil)
	rec := httptest.NewRecorder()

	handler.RemoveRepository(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestGitHubHandler_UpdateToken(t *testing.T) {
	t.Parallel()

	handler, repo := newTestHandler()

	// Register a repo first
	body := `{"github_repo_url":"https://github.com/org/repo","personal_access_token":"ghp_old"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams/team-1/github-repos", strings.NewReader(body))
	regRec := httptest.NewRecorder()
	handler.RegisterRepository(regRec, regReq)

	repoID := string(repo.saved[0].ID)

	// Update token
	updateBody := `{"personal_access_token":"ghp_new"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/team-1/github-repos/"+repoID, strings.NewReader(updateBody))
	rec := httptest.NewRecorder()

	handler.UpdateToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestGitHubHandler_UpdateToken_InvalidBody(t *testing.T) {
	t.Parallel()

	handler, _ := newTestHandler()

	req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/team-1/github-repos/repo-1", strings.NewReader("invalid"))
	rec := httptest.NewRecorder()

	handler.UpdateToken(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
