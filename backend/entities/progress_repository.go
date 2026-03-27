package entities

import "context"

// ProgressRepository defines the interface for persisting progress logs.
type ProgressRepository interface {
  Save(ctx context.Context, log *ProgressLog) error
}

// TeamProgress holds team information along with its latest progress log.
type TeamProgress struct {
  TeamID         TeamID
  TeamName       string
  LatestProgress *ProgressLog
}

// ProgressQueryRepository defines the interface for querying progress data.
type ProgressQueryRepository interface {
  // ListLatestByTeam returns the latest progress for all teams.
  ListLatestByTeam(ctx context.Context) ([]TeamProgress, error)
  // ListLatestByTeamID returns the latest progress filtered by team ID.
  ListLatestByTeamID(ctx context.Context, teamID TeamID) ([]TeamProgress, error)
}
