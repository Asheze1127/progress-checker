package usecase

import (
  "context"
  "errors"
  "testing"

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

func TestListProgressUseCase_Execute(t *testing.T) {
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

  tests := []struct {
    name          string
    teamID        string
    mockRepo      *mockProgressQueryRepository
    expectedCount int
    expectedError bool
  }{
    {
      name:   "list all teams when team_id is empty",
      teamID: "",
      mockRepo: &mockProgressQueryRepository{
        listLatestByTeamFunc: func(ctx context.Context) ([]entities.TeamProgress, error) {
          return sampleProgress, nil
        },
      },
      expectedCount: 1,
      expectedError: false,
    },
    {
      name:   "filter by team_id when provided",
      teamID: "team-1",
      mockRepo: &mockProgressQueryRepository{
        listLatestByTeamIDFunc: func(ctx context.Context, teamID entities.TeamID) ([]entities.TeamProgress, error) {
          if teamID != "team-1" {
            t.Errorf("expected teamID %q, got %q", "team-1", teamID)
          }
          return sampleProgress, nil
        },
      },
      expectedCount: 1,
      expectedError: false,
    },
    {
      name:   "return error when repository fails for all teams",
      teamID: "",
      mockRepo: &mockProgressQueryRepository{
        listLatestByTeamFunc: func(ctx context.Context) ([]entities.TeamProgress, error) {
          return nil, errors.New("database connection failed")
        },
      },
      expectedCount: 0,
      expectedError: true,
    },
    {
      name:   "return error when repository fails for filtered team",
      teamID: "team-1",
      mockRepo: &mockProgressQueryRepository{
        listLatestByTeamIDFunc: func(ctx context.Context, teamID entities.TeamID) ([]entities.TeamProgress, error) {
          return nil, errors.New("database connection failed")
        },
      },
      expectedCount: 0,
      expectedError: true,
    },
    {
      name:   "return empty list when no progress exists",
      teamID: "",
      mockRepo: &mockProgressQueryRepository{
        listLatestByTeamFunc: func(ctx context.Context) ([]entities.TeamProgress, error) {
          return []entities.TeamProgress{}, nil
        },
      },
      expectedCount: 0,
      expectedError: false,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      uc := NewListProgressUseCase(tt.mockRepo)
      results, err := uc.Execute(context.Background(), tt.teamID)

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
