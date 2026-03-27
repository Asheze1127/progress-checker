package postgres

import (
	"context"
	"database/sql"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

var _ entities.MentorRepository = (*MentorRepository)(nil)

// MentorRepository manages mentor records in PostgreSQL.
type MentorRepository struct {
	queries *db.Queries
}

// NewMentorRepository creates a new MentorRepository.
func NewMentorRepository(database *sql.DB) *MentorRepository {
	return &MentorRepository{queries: db.New(database)}
}

// Create creates a new mentor record for a user.
func (r *MentorRepository) Create(ctx context.Context, userID entities.UserID) error {
	uid, err := uuid.Parse(string(userID))
	if err != nil {
		return err
	}
	return r.queries.CreateMentor(ctx, uid)
}

func (r *MentorRepository) AssignTeam(ctx context.Context, userID entities.UserID, teamID entities.TeamID) error {
	uid, err := uuid.Parse(string(userID))
	if err != nil {
		return err
	}
	tid, err := uuid.Parse(string(teamID))
	if err != nil {
		return err
	}
	return r.queries.CreateMentorTeamAssignment(ctx, db.CreateMentorTeamAssignmentParams{
		MentorUserID: uid,
		TeamID:       tid,
	})
}

func (r *MentorRepository) GetTeamIDs(ctx context.Context, userID entities.UserID) ([]entities.TeamID, error) {
	uid, err := uuid.Parse(string(userID))
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.GetMentorTeamIDs(ctx, uid)
	if err != nil {
		return nil, err
	}
	ids := make([]entities.TeamID, len(rows))
	for i, row := range rows {
		ids[i] = entities.TeamID(row.String())
	}
	return ids, nil
}
