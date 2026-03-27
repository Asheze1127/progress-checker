package github_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	githubclient "github.com/Asheze1127/progress-checker/backend/infrastructure/github"
)

func TestClient_CreateIssue_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		expectedPath := "/repos/test-org/test-repo/issues"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") && !strings.HasPrefix(authHeader, "token ") {
			t.Errorf("expected auth header, got %s", authHeader)
		}

		w.WriteHeader(http.StatusCreated)
		resp := map[string]any{
			"html_url": "https://github.com/test-org/test-repo/issues/1",
			"number":   1,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := githubclient.NewClient(server.URL)

	issueURL, err := client.CreateIssue(context.Background(), "test-org", "test-repo", "ghp_test", "Test Issue", "Test body")
	if err != nil {
		t.Fatalf("CreateIssue() error = %v", err)
	}

	expected := "https://github.com/test-org/test-repo/issues/1"
	if issueURL != expected {
		t.Errorf("CreateIssue() = %v, want %v", issueURL, expected)
	}
}

func TestClient_CreateIssue_APIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"message":"Validation Failed"}`))
	}))
	defer server.Close()

	client := githubclient.NewClient(server.URL)

	_, err := client.CreateIssue(context.Background(), "test-org", "test-repo", "ghp_test", "Test Issue", "Test body")
	if err == nil {
		t.Fatal("CreateIssue() expected error for API failure")
	}
}

func TestClient_CreateIssue_ValidationErrors(t *testing.T) {
	t.Parallel()

	client := githubclient.NewClient("")

	tests := []struct {
		name  string
		owner string
		repo  string
		token string
		title string
	}{
		{name: "empty owner", owner: "", repo: "repo", token: "tok", title: "title"},
		{name: "empty repo", owner: "owner", repo: "", token: "tok", title: "title"},
		{name: "empty token", owner: "owner", repo: "repo", token: "", title: "title"},
		{name: "empty title", owner: "owner", repo: "repo", token: "tok", title: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.CreateIssue(context.Background(), tt.owner, tt.repo, tt.token, tt.title, "body")
			if err == nil {
				t.Error("CreateIssue() expected validation error")
			}
		})
	}
}
