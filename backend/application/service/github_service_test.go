package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// --- Test doubles ---

type mockGitHubRepoRepository struct {
	saved     []*entities.GitHubRepo
	repos     map[string][]entities.GitHubRepo
	byID      map[string]*entities.GitHubRepo
	byChannel map[string]*entities.GitHubRepo
	deleteErr error
}

func newMockRepo() *mockGitHubRepoRepository {
	return &mockGitHubRepoRepository{
		repos:     make(map[string][]entities.GitHubRepo),
		byID:      make(map[string]*entities.GitHubRepo),
		byChannel: make(map[string]*entities.GitHubRepo),
	}
}

func (m *mockGitHubRepoRepository) Save(_ context.Context, repo *entities.GitHubRepo) error {
	m.saved = append(m.saved, repo)
	m.byID[string(repo.ID)] = repo
	teamID := string(repo.TeamID)
	m.repos[teamID] = append(m.repos[teamID], *repo)
	return nil
}

func (m *mockGitHubRepoRepository) FindByTeamID(_ context.Context, teamID string) ([]entities.GitHubRepo, error) {
	return m.repos[teamID], nil
}

func (m *mockGitHubRepoRepository) FindByID(_ context.Context, repoID string) (*entities.GitHubRepo, error) {
	repo, ok := m.byID[repoID]
	if !ok {
		return nil, fmt.Errorf("github repo not found: %s", repoID)
	}
	return repo, nil
}

func (m *mockGitHubRepoRepository) FindByChannelID(_ context.Context, channelID string) (*entities.GitHubRepo, error) {
	repo, ok := m.byChannel[channelID]
	if !ok {
		return nil, fmt.Errorf("github repo not found for channel: %s", channelID)
	}
	return repo, nil
}

func (m *mockGitHubRepoRepository) Delete(_ context.Context, repoID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.byID, repoID)
	return nil
}

func (m *mockGitHubRepoRepository) UpdateToken(_ context.Context, repoID string, encryptedToken string) error {
	repo, ok := m.byID[repoID]
	if !ok {
		return fmt.Errorf("github repo not found: %s", repoID)
	}
	repo.EncryptedToken = encryptedToken
	return nil
}

type mockEncryptor struct{}

func (m *mockEncryptor) Encrypt(plaintext string) (string, error) {
	return "enc:" + plaintext, nil
}

func (m *mockEncryptor) Decrypt(ciphertext string) (string, error) {
	if len(ciphertext) < 4 {
		return "", fmt.Errorf("invalid ciphertext")
	}
	return ciphertext[4:], nil // strip "enc:" prefix
}

type mockGitHubIssueCreator struct {
	issueURL string
	err      error
}

func (m *mockGitHubIssueCreator) CreateIssue(_ context.Context, _, _, _, _, _ string) (string, error) {
	return m.issueURL, m.err
}

func newTestService(repo *mockGitHubRepoRepository, ghClient *mockGitHubIssueCreator) *service.GitHubService {
	idCounter := 0
	return service.NewGitHubService(
		repo,
		&mockEncryptor{},
		ghClient,
		func() string {
			idCounter++
			return fmt.Sprintf("repo-%d", idCounter)
		},
	)
}

// --- Tests ---

func TestRegisterRepository(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := newTestService(repo, &mockGitHubIssueCreator{})

	err := svc.RegisterRepository(context.Background(), "team-1", "https://github.com/org/repo", "ghp_test123")
	if err != nil {
		t.Fatalf("RegisterRepository() error = %v", err)
	}

	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 saved repo, got %d", len(repo.saved))
	}

	saved := repo.saved[0]
	if saved.Owner != "org" {
		t.Errorf("expected owner 'org', got %s", saved.Owner)
	}
	if saved.RepoName != "repo" {
		t.Errorf("expected repo_name 'repo', got %s", saved.RepoName)
	}
	if saved.EncryptedToken == "ghp_test123" {
		t.Error("token should be encrypted, not stored as plaintext")
	}
}

func TestRegisterRepository_ValidationErrors(t *testing.T) {
	t.Parallel()

	svc := newTestService(newMockRepo(), &mockGitHubIssueCreator{})

	tests := []struct {
		name    string
		teamID  string
		repoURL string
		pat     string
	}{
		{name: "empty team_id", teamID: "", repoURL: "https://github.com/org/repo", pat: "ghp_test"},
		{name: "empty repo_url", teamID: "team-1", repoURL: "", pat: "ghp_test"},
		{name: "empty pat", teamID: "team-1", repoURL: "https://github.com/org/repo", pat: ""},
		{name: "invalid repo_url", teamID: "team-1", repoURL: "not-a-url", pat: "ghp_test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := svc.RegisterRepository(context.Background(), tt.teamID, tt.repoURL, tt.pat)
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestListRepositories(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := newTestService(repo, &mockGitHubIssueCreator{})

	err := svc.RegisterRepository(context.Background(), "team-1", "https://github.com/org/repo", "ghp_test")
	if err != nil {
		t.Fatalf("RegisterRepository() error = %v", err)
	}

	repos, err := svc.ListRepositories(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("ListRepositories() error = %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
}

func TestListRepositories_EmptyTeamID(t *testing.T) {
	t.Parallel()

	svc := newTestService(newMockRepo(), &mockGitHubIssueCreator{})

	_, err := svc.ListRepositories(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty team_id")
	}
}

func TestRemoveRepository(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := newTestService(repo, &mockGitHubIssueCreator{})

	err := svc.RegisterRepository(context.Background(), "team-1", "https://github.com/org/repo", "ghp_test")
	if err != nil {
		t.Fatalf("RegisterRepository() error = %v", err)
	}

	repoID := string(repo.saved[0].ID)

	err = svc.RemoveRepository(context.Background(), "team-1", repoID)
	if err != nil {
		t.Fatalf("RemoveRepository() error = %v", err)
	}
}

func TestRemoveRepository_WrongTeam(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := newTestService(repo, &mockGitHubIssueCreator{})

	err := svc.RegisterRepository(context.Background(), "team-1", "https://github.com/org/repo", "ghp_test")
	if err != nil {
		t.Fatalf("RegisterRepository() error = %v", err)
	}

	repoID := string(repo.saved[0].ID)

	err = svc.RemoveRepository(context.Background(), "team-2", repoID)
	if err == nil {
		t.Error("expected error when removing repo from wrong team")
	}
}

func TestUpdateToken(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := newTestService(repo, &mockGitHubIssueCreator{})

	err := svc.RegisterRepository(context.Background(), "team-1", "https://github.com/org/repo", "ghp_old")
	if err != nil {
		t.Fatalf("RegisterRepository() error = %v", err)
	}

	repoID := string(repo.saved[0].ID)

	err = svc.UpdateToken(context.Background(), "team-1", repoID, "ghp_new")
	if err != nil {
		t.Fatalf("UpdateToken() error = %v", err)
	}

	updated := repo.byID[repoID]
	if updated.EncryptedToken == "ghp_new" {
		t.Error("token should be encrypted, not stored as plaintext")
	}
}

func TestCreateIssue(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	repo.byChannel["C123"] = &entities.GitHubRepo{
		ID:             "repo-1",
		TeamID:         "team-1",
		Owner:          "org",
		RepoName:       "repo",
		EncryptedToken: "enc:ghp_test",
	}

	ghClient := &mockGitHubIssueCreator{
		issueURL: "https://github.com/org/repo/issues/42",
	}

	svc := newTestService(repo, ghClient)

	issueURL, err := svc.CreateIssue(context.Background(), "C123", "Bug fix", "Fix the bug")
	if err != nil {
		t.Fatalf("CreateIssue() error = %v", err)
	}

	expected := "https://github.com/org/repo/issues/42"
	if issueURL != expected {
		t.Errorf("CreateIssue() = %v, want %v", issueURL, expected)
	}
}

func TestCreateIssue_ChannelNotFound(t *testing.T) {
	t.Parallel()

	svc := newTestService(newMockRepo(), &mockGitHubIssueCreator{})

	_, err := svc.CreateIssue(context.Background(), "C_UNKNOWN", "Title", "Body")
	if err == nil {
		t.Error("expected error for unknown channel")
	}
}

func TestCreateIssue_ValidationErrors(t *testing.T) {
	t.Parallel()

	svc := newTestService(newMockRepo(), &mockGitHubIssueCreator{})

	tests := []struct {
		name      string
		channelID string
		title     string
	}{
		{name: "empty channel_id", channelID: "", title: "Title"},
		{name: "empty title", channelID: "C123", title: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := svc.CreateIssue(context.Background(), tt.channelID, tt.title, "Body")
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}
