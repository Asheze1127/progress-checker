package entities

import "context"

// TeamRepository defines the interface for querying and managing teams.
type TeamRepository interface {
	GetByID(ctx context.Context, id TeamID) (*Team, error)
	List(ctx context.Context) ([]*Team, error)
	Create(ctx context.Context, name string) (*Team, error)
}
