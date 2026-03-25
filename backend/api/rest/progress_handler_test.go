package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// mockProgressQueryRepository is a mock for testing handlers.
type mockProgressQueryRepository struct {
	listLatestByTeamFunc   func(ctx context.Context) ([]entities.TeamProgress, error)
	listLatestByTeamIDFunc func(ctx context.Context, teamID entities.TeamID) ([]entities.TeamProgress, error)
}

func (m *mockProgressQueryRepository) ListLatestByTeam(ctx context.Context) ([]entities.TeamProgress, error) {
	return m.listLatestByTeamFunc(ctx)
}

func (m *mockProgressQueryRepository) ListLatestByTeamID(ctx context.Context, teamID entities.TeamID) ([]entities.TeamProgress, error) {
	return m.listLatestByTeamIDFunc(ctx, teamID)
}

func newTestHandler(repo entities.ProgressQueryRepository) *ProgressHandler {
	uc := usecase.NewListProgressUseCase(repo)
	return NewProgressHandler(uc)
}

func TestHandleListProgress(t *testing.T) {
	fixedTime := time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC)

	sampleData := []entities.TeamProgress{
		{
			TeamID:   "team-uuid-1",
			TeamName: "Team Alpha",
			LatestProgress: &entities.ProgressLog{
				ID:            "log-uuid-1",
				ParticipantID: "participant-uuid-1",
				ProgressBodies: []entities.ProgressBody{
					{
						Phase:       entities.ProgressPhaseCoding,
						SOS:         false,
						Comment:     "Working on API",
						SubmittedAt: fixedTime,
					},
				},
			},
		},
	}

	tests := []struct {
		name               string
		method             string
		queryParams        string
		mockRepo           *mockProgressQueryRepository
		expectedStatusCode int
		validateBody       func(t *testing.T, body []byte)
	}{
		{
			name:   "successful response with all teams",
			method: http.MethodGet,
			mockRepo: &mockProgressQueryRepository{
				listLatestByTeamFunc: func(ctx context.Context) ([]entities.TeamProgress, error) {
					return sampleData, nil
				},
			},
			expectedStatusCode: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var resp progressListResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(resp.Data) != 1 {
					t.Fatalf("expected 1 team, got %d", len(resp.Data))
				}
				if resp.Data[0].TeamID != "team-uuid-1" {
					t.Errorf("expected team_id %q, got %q", "team-uuid-1", resp.Data[0].TeamID)
				}
				if resp.Data[0].TeamName != "Team Alpha" {
					t.Errorf("expected team_name %q, got %q", "Team Alpha", resp.Data[0].TeamName)
				}
				if resp.Data[0].LatestProgress == nil {
					t.Fatal("expected latest_progress to be present")
				}
				if resp.Data[0].LatestProgress.ID != "log-uuid-1" {
					t.Errorf("expected progress id %q, got %q", "log-uuid-1", resp.Data[0].LatestProgress.ID)
				}
				if len(resp.Data[0].LatestProgress.ProgressBodies) != 1 {
					t.Fatalf("expected 1 progress body, got %d", len(resp.Data[0].LatestProgress.ProgressBodies))
				}
				pb := resp.Data[0].LatestProgress.ProgressBodies[0]
				if pb.Phase != "coding" {
					t.Errorf("expected phase %q, got %q", "coding", pb.Phase)
				}
				if pb.SOS {
					t.Error("expected sos to be false")
				}
				if pb.Comment != "Working on API" {
					t.Errorf("expected comment %q, got %q", "Working on API", pb.Comment)
				}
				expectedTime := "2026-03-25T10:00:00Z"
				if pb.SubmittedAt != expectedTime {
					t.Errorf("expected submitted_at %q, got %q", expectedTime, pb.SubmittedAt)
				}
			},
		},
		{
			name:        "successful response with team_id filter",
			method:      http.MethodGet,
			queryParams: "team_id=team-uuid-1",
			mockRepo: &mockProgressQueryRepository{
				listLatestByTeamIDFunc: func(ctx context.Context, teamID entities.TeamID) ([]entities.TeamProgress, error) {
					if teamID != "team-uuid-1" {
						t.Errorf("expected teamID %q, got %q", "team-uuid-1", teamID)
					}
					return sampleData, nil
				},
			},
			expectedStatusCode: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var resp progressListResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(resp.Data) != 1 {
					t.Fatalf("expected 1 team, got %d", len(resp.Data))
				}
			},
		},
		{
			name:   "returns empty data array when no results",
			method: http.MethodGet,
			mockRepo: &mockProgressQueryRepository{
				listLatestByTeamFunc: func(ctx context.Context) ([]entities.TeamProgress, error) {
					return []entities.TeamProgress{}, nil
				},
			},
			expectedStatusCode: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var resp progressListResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(resp.Data) != 0 {
					t.Errorf("expected 0 teams, got %d", len(resp.Data))
				}
			},
		},
		{
			name:   "returns 500 on repository error",
			method: http.MethodGet,
			mockRepo: &mockProgressQueryRepository{
				listLatestByTeamFunc: func(ctx context.Context) ([]entities.TeamProgress, error) {
					return nil, errors.New("database error")
				},
			},
			expectedStatusCode: http.StatusInternalServerError,
			validateBody: func(t *testing.T, body []byte) {
				var resp map[string]string
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("failed to unmarshal error response: %v", err)
				}
				if resp["error"] != "internal server error" {
					t.Errorf("expected error message %q, got %q", "internal server error", resp["error"])
				}
			},
		},
		{
			name:   "returns 405 for non-GET method",
			method: http.MethodPost,
			mockRepo: &mockProgressQueryRepository{
				listLatestByTeamFunc: func(ctx context.Context) ([]entities.TeamProgress, error) {
					return nil, nil
				},
			},
			expectedStatusCode: http.StatusMethodNotAllowed,
			validateBody: func(t *testing.T, body []byte) {
				var resp map[string]string
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("failed to unmarshal error response: %v", err)
				}
				if resp["error"] != "method not allowed" {
					t.Errorf("expected error message %q, got %q", "method not allowed", resp["error"])
				}
			},
		},
		{
			name:   "team with nil latest_progress",
			method: http.MethodGet,
			mockRepo: &mockProgressQueryRepository{
				listLatestByTeamFunc: func(ctx context.Context) ([]entities.TeamProgress, error) {
					return []entities.TeamProgress{
						{
							TeamID:         "team-uuid-2",
							TeamName:       "Team Beta",
							LatestProgress: nil,
						},
					}, nil
				},
			},
			expectedStatusCode: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var resp progressListResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(resp.Data) != 1 {
					t.Fatalf("expected 1 team, got %d", len(resp.Data))
				}
				if resp.Data[0].LatestProgress != nil {
					t.Error("expected latest_progress to be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler(tt.mockRepo)

			path := "/api/v1/progress"
			if tt.queryParams != "" {
				path += "?" + tt.queryParams
			}

			req := httptest.NewRequest(tt.method, path, nil)
			rec := httptest.NewRecorder()

			handler.HandleListProgress(rec, req)

			if rec.Code != tt.expectedStatusCode {
				t.Errorf("expected status %d, got %d", tt.expectedStatusCode, rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json; charset=utf-8" {
				t.Errorf("expected Content-Type %q, got %q", "application/json; charset=utf-8", contentType)
			}

			if tt.validateBody != nil {
				tt.validateBody(t, rec.Body.Bytes())
			}
		})
	}
}

func TestHandleListProgress_CORSHeaders(t *testing.T) {
	repo := &mockProgressQueryRepository{
		listLatestByTeamFunc: func(ctx context.Context) ([]entities.TeamProgress, error) {
			return []entities.TeamProgress{}, nil
		},
	}
	handler := newTestHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/progress", nil)
	rec := httptest.NewRecorder()

	handler.HandleListProgress(rec, req)

	corsOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if corsOrigin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin %q, got %q", "*", corsOrigin)
	}

	corsMethods := rec.Header().Get("Access-Control-Allow-Methods")
	if corsMethods != "GET, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods %q, got %q", "GET, OPTIONS", corsMethods)
	}

	corsHeaders := rec.Header().Get("Access-Control-Allow-Headers")
	if corsHeaders != "Content-Type, Authorization" {
		t.Errorf("expected Access-Control-Allow-Headers %q, got %q", "Content-Type, Authorization", corsHeaders)
	}
}
