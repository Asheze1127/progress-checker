package postgres

import (
	"context"
	"database/sql"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

var _ entities.TeamRepository = (*TeamRepository)(nil)

// TeamRepository queries teams from PostgreSQL.
type TeamRepository struct {
	queries *db.Queries
}

// NewTeamRepository creates a new TeamRepository backed by the given database connection.
func NewTeamRepository(database *sql.DB) *TeamRepository {
	return &TeamRepository{queries: db.New(database)}
}

func (r *TeamRepository) GetByID(ctx context.Context, id entities.TeamID) (*entities.Team, error) {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return nil, err
	}
	row, err := r.queries.GetTeamByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return toTeamEntity(row), nil
}

func (r *TeamRepository) List(ctx context.Context) ([]*entities.Team, error) {
	rows, err := r.queries.ListTeams(ctx)
	if err != nil {
		return nil, err
	}
	teams := make([]*entities.Team, len(rows))
	for i, row := range rows {
		teams[i] = toTeamEntity(row)
	}
	return teams, nil
}

func (r *TeamRepository) Create(ctx context.Context, name string) (*entities.Team, error) {
	row, err := r.queries.CreateTeam(ctx, name)
	if err != nil {
		return nil, err
	}
	return toTeamEntity(row), nil
}

func toTeamEntity(row db.Teams) *entities.Team {
	return &entities.Team{
		ID:   entities.TeamID(row.ID.String()),
		Name: row.Name,
	}
}
