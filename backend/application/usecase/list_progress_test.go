package usecase

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Asheze1127/progress-checker/backend/api/middleware"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// mockProgressQueryRepository is a mock implementation of ProgressQueryRepository.
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

type mockMentorRepository struct {
	getByUserIDFunc func(ctx context.Context, userID entities.UserID) (*entities.Mentor, error)
}

func (m *mockMentorRepository) Create(_ context.Context, _ entities.UserID) error {
	return errors.New("not implemented")
}

func (m *mockMentorRepository) AssignTeam(_ context.Context, _ entities.UserID, _ entities.TeamID) error {
	return errors.New("not implemented")
}

func (m *mockMentorRepository) GetByUserID(ctx context.Context, userID entities.UserID) (*entities.Mentor, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, userID)
	}
	return nil, errors.New("not found")
}

func (m *mockMentorRepository) GetTeamIDs(_ context.Context, _ entities.UserID) ([]entities.TeamID, error) {
	return nil, errors.New("not implemented")
}

// contextWithMentorUser creates a Gin context with an authenticated mentor user.
func contextWithMentorUser(userID entities.UserID) context.Context {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(middleware.GinKeyUser, &entities.User{
		ID:   userID,
		Name: "Test Mentor",
		Role: entities.UserRoleMentor,
	})
	return c
}

func TestListProgressUseCase_Execute_WithAuth(t *testing.T) {
	sampleProgress := []entities.TeamProgress{
		{
			TeamID:   "team-1",
			TeamName: "Team Alpha",
			LatestProgress: &entities.ProgressLog{
				ID:            "log-1",
				ParticipantID: "participant-1",
			},
		},
	}

	mentor := entities.NewMentor("user-1", []entities.TeamID{"team-1"})

	tests := []struct {
		name          string
		teamID        string
		mockRepo      *mockProgressQueryRepository
		mockMentor    *mockMentorRepository
		ctx           context.Context
		expectedCount int
		expectedError bool
	}{
		{
			name:   "list progress for mentor's teams",
			teamID: "",
			mockRepo: &mockProgressQueryRepository{
				listLatestByTeamIDFunc: func(_ context.Context, _ entities.TeamID) ([]entities.TeamProgress, error) {
					return sampleProgress, nil
				},
			},
			mockMentor: &mockMentorRepository{
				getByUserIDFunc: func(_ context.Context, _ entities.UserID) (*entities.Mentor, error) {
					return mentor, nil
				},
			},
			ctx:           contextWithMentorUser("user-1"),
			expectedCount: 1,
			expectedError: false,
		},
		{
			name:   "filter by team_id when provided",
			teamID: "team-1",
			mockRepo: &mockProgressQueryRepository{
				listLatestByTeamIDFunc: func(_ context.Context, teamID entities.TeamID) ([]entities.TeamProgress, error) {
					if teamID != "team-1" {
						t.Errorf("expected teamID %q, got %q", "team-1", teamID)
					}
					return sampleProgress, nil
				},
			},
			mockMentor: &mockMentorRepository{
				getByUserIDFunc: func(_ context.Context, _ entities.UserID) (*entities.Mentor, error) {
					return mentor, nil
				},
			},
			ctx:           contextWithMentorUser("user-1"),
			expectedCount: 1,
			expectedError: false,
		},
		{
			name:   "error when not authenticated",
			teamID: "",
			mockRepo: &mockProgressQueryRepository{},
			mockMentor: &mockMentorRepository{},
			ctx:           context.Background(),
			expectedCount: 0,
			expectedError: true,
		},
		{
			name:   "error when repository fails",
			teamID: "team-1",
			mockRepo: &mockProgressQueryRepository{
				listLatestByTeamIDFunc: func(_ context.Context, _ entities.TeamID) ([]entities.TeamProgress, error) {
					return nil, errors.New("database connection failed")
				},
			},
			mockMentor: &mockMentorRepository{
				getByUserIDFunc: func(_ context.Context, _ entities.UserID) (*entities.Mentor, error) {
					return mentor, nil
				},
			},
			ctx:           contextWithMentorUser("user-1"),
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewListProgressUseCase(tt.mockRepo, tt.mockMentor)
			results, err := uc.Execute(tt.ctx, tt.teamID)

			if tt.expectedError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectedError && len(results) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(results))
			}
		})
	}
}
