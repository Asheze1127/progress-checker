package entities_test

import (
	"testing"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

func TestGitHubRepo_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		repo    entities.GitHubRepo
		wantErr bool
	}{
		{
			name: "valid repo",
			repo: entities.GitHubRepo{
				ID:             "repo-1",
				TeamID:         "team-1",
				Owner:          "org",
				RepoName:       "repo",
				EncryptedToken: "encrypted-token",
			},
			wantErr: false,
		},
		{
			name: "missing id",
			repo: entities.GitHubRepo{
				ID:             "",
				TeamID:         "team-1",
				Owner:          "org",
				RepoName:       "repo",
				EncryptedToken: "encrypted-token",
			},
			wantErr: true,
		},
		{
			name: "missing team_id",
			repo: entities.GitHubRepo{
				ID:             "repo-1",
				TeamID:         "",
				Owner:          "org",
				RepoName:       "repo",
				EncryptedToken: "encrypted-token",
			},
			wantErr: true,
		},
		{
			name: "missing owner",
			repo: entities.GitHubRepo{
				ID:             "repo-1",
				TeamID:         "team-1",
				Owner:          "",
				RepoName:       "repo",
				EncryptedToken: "encrypted-token",
			},
			wantErr: true,
		},
		{
			name: "missing repo_name",
			repo: entities.GitHubRepo{
				ID:             "repo-1",
				TeamID:         "team-1",
				Owner:          "org",
				RepoName:       "",
				EncryptedToken: "encrypted-token",
			},
			wantErr: true,
		},
		{
			name: "missing encrypted_token",
			repo: entities.GitHubRepo{
				ID:             "repo-1",
				TeamID:         "team-1",
				Owner:          "org",
				RepoName:       "repo",
				EncryptedToken: "",
			},
			wantErr: true,
		},
		{
			name:    "all fields missing",
			repo:    entities.GitHubRepo{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.repo.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseGitHubRepoURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawURL   string
		wantOwn  string
		wantRepo string
		wantErr  bool
	}{
		{
			name:     "valid URL",
			rawURL:   "https://github.com/org/repo",
			wantOwn:  "org",
			wantRepo: "repo",
			wantErr:  false,
		},
		{
			name:     "valid URL with .git suffix",
			rawURL:   "https://github.com/org/repo.git",
			wantOwn:  "org",
			wantRepo: "repo",
			wantErr:  false,
		},
		{
			name:     "valid URL with trailing slash",
			rawURL:   "https://github.com/org/repo/",
			wantOwn:  "org",
			wantRepo: "repo",
			wantErr:  false,
		},
		{
			name:    "empty URL",
			rawURL:  "",
			wantErr: true,
		},
		{
			name:    "http scheme rejected",
			rawURL:  "http://github.com/org/repo",
			wantErr: true,
		},
		{
			name:    "non-github host rejected",
			rawURL:  "https://gitlab.com/org/repo",
			wantErr: true,
		},
		{
			name:    "missing repo name",
			rawURL:  "https://github.com/org",
			wantErr: true,
		},
		{
			name:    "too many path segments",
			rawURL:  "https://github.com/org/repo/extra",
			wantErr: true,
		},
		{
			name:    "whitespace-only URL",
			rawURL:  "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			owner, repoName, err := entities.ParseGitHubRepoURL(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGitHubRepoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner != tt.wantOwn {
					t.Errorf("ParseGitHubRepoURL() owner = %v, want %v", owner, tt.wantOwn)
				}
				if repoName != tt.wantRepo {
					t.Errorf("ParseGitHubRepoURL() repoName = %v, want %v", repoName, tt.wantRepo)
				}
			}
		})
	}
}
