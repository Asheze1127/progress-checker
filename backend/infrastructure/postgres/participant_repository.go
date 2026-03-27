package postgres

import (
	"context"
	"database/sql"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

var _ entities.ParticipantRepository = (*ParticipantRepository)(nil)

// ParticipantRepository manages participant records in PostgreSQL.
type ParticipantRepository struct {
	queries *db.Queries
}

// NewParticipantRepository creates a new ParticipantRepository.
func NewParticipantRepository(database *sql.DB) *ParticipantRepository {
	return &ParticipantRepository{queries: db.New(database)}
}

// Create creates a new participant record linking a user to a team.
func (r *ParticipantRepository) Create(ctx context.Context, userID entities.UserID, teamID entities.TeamID) error {
	uid, err := uuid.Parse(string(userID))
	if err != nil {
		return err
	}
	tid, err := uuid.Parse(string(teamID))
	if err != nil {
		return err
	}
	return r.queries.CreateParticipant(ctx, db.CreateParticipantParams{
		UserID: uid,
		TeamID: tid,
	})
}
