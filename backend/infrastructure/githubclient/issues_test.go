package githubclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	githubissuecreator "github.com/Asheze1127/progress-checker/backend/application/service/github_issue_creator"
)

// newTestClient creates a Client pointing at a test server, bypassing HTTPS validation.
func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	client := &Client{
		httpClient: server.Client(),
		baseURL:    server.URL + "/",
		token:      "ghp_test",
	}
	return client
}

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

		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if reqBody["title"] != "Test Issue" {
			t.Errorf("expected title 'Test Issue', got %v", reqBody["title"])
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

	client := newTestClient(t, server)

	result, err := client.CreateIssue(context.Background(), githubissuecreator.CreateIssueInput{
		Owner: "test-org",
		Repo:  "test-repo",
		Title: "Test Issue",
		Body:  "Test body",
	})
	if err != nil {
		t.Fatalf("CreateIssue() error = %v", err)
	}

	expectedURL := "https://github.com/test-org/test-repo/issues/1"
	if result.URL != expectedURL {
		t.Errorf("CreateIssue() URL = %v, want %v", result.URL, expectedURL)
	}
	if result.Number != 1 {
		t.Errorf("CreateIssue() Number = %v, want %v", result.Number, 1)
	}
}

func TestClient_CreateIssue_WithLabels(t *testing.T) {
	t.Parallel()

	var receivedLabels []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if labels, ok := reqBody["labels"].([]any); ok {
			for _, l := range labels {
				receivedLabels = append(receivedLabels, l.(string))
			}
		}

		w.WriteHeader(http.StatusCreated)
		resp := map[string]any{
			"html_url": "https://github.com/o/r/issues/2",
			"number":   2,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server)

	_, err := client.CreateIssue(context.Background(), githubissuecreator.CreateIssueInput{
		Owner:  "o",
		Repo:   "r",
		Title:  "Labeled Issue",
		Labels: []string{"bug", " enhancement ", ""},
	})
	if err != nil {
		t.Fatalf("CreateIssue() error = %v", err)
	}

	if len(receivedLabels) != 2 {
		t.Errorf("expected 2 labels, got %d: %v", len(receivedLabels), receivedLabels)
	}
}

func TestClient_CreateIssue_APIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		resp := map[string]string{"message": "Validation Failed"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server)

	_, err := client.CreateIssue(context.Background(), githubissuecreator.CreateIssueInput{
		Owner: "test-org",
		Repo:  "test-repo",
		Title: "Test Issue",
	})
	if err == nil {
		t.Fatal("CreateIssue() expected error for API failure")
	}
}

func TestClient_CreateIssue_ValidationErrors(t *testing.T) {
	t.Parallel()

	client := &Client{token: "ghp_test"}

	tests := []struct {
		name  string
		input githubissuecreator.CreateIssueInput
	}{
		{name: "empty owner", input: githubissuecreator.CreateIssueInput{Owner: "", Repo: "repo", Title: "title"}},
		{name: "empty repo", input: githubissuecreator.CreateIssueInput{Owner: "owner", Repo: "", Title: "title"}},
		{name: "empty title", input: githubissuecreator.CreateIssueInput{Owner: "owner", Repo: "repo", Title: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.CreateIssue(context.Background(), tt.input)
			if err == nil {
				t.Error("CreateIssue() expected validation error")
			}
		})
	}
}

func TestNewClient_RequiresAuth(t *testing.T) {
	t.Parallel()

	_, err := NewClient(Config{})
	if err == nil {
		t.Error("NewClient() expected error when no auth configured")
	}
}

func TestNewClient_RejectsHTTPBaseURL(t *testing.T) {
	t.Parallel()

	_, err := NewClient(Config{
		Token:   "ghp_test",
		BaseURL: "http://example.com",
	})
	if err == nil {
		t.Error("NewClient() expected error for non-HTTPS base URL")
	}
}

func TestSanitizeLabels(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "removes empty labels",
			input:    []string{"bug", "", " ", "fix"},
			expected: []string{"bug", "fix"},
		},
		{
			name:     "trims whitespace",
			input:    []string{" bug ", " fix "},
			expected: []string{"bug", "fix"},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeLabels(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("sanitizeLabels() length = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("sanitizeLabels()[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}
