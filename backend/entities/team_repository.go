package entities

import "context"

// TeamRepository defines the interface for querying teams.
type TeamRepository interface {
  GetByID(ctx context.Context, id TeamID) (*Team, error)
  List(ctx context.Context) ([]*Team, error)
}
