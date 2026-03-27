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
