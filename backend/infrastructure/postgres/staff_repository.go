package postgres

import (
	"context"
	"database/sql"
	"fmt"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

var _ entities.StaffRepository = (*StaffRepository)(nil)

// StaffRepository queries staff from PostgreSQL.
type StaffRepository struct {
	queries *db.Queries
}

// NewStaffRepository creates a new StaffRepository backed by the given database connection.
func NewStaffRepository(database *sql.DB) *StaffRepository {
	return &StaffRepository{queries: db.New(database)}
}

func (r *StaffRepository) FindByEmail(ctx context.Context, email string) (*entities.StaffWithPassword, error) {
	row, err := r.queries.GetStaffByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("find staff by email: %w", err)
	}

	var slackUserID *entities.SlackUserID
	if row.SlackUserID.Valid {
		sid := entities.SlackUserID(row.SlackUserID.String)
		slackUserID = &sid
	}

	return &entities.StaffWithPassword{
		Staff: entities.Staff{
			ID:          entities.StaffID(row.ID.String()),
			SlackUserID: slackUserID,
			Name:        row.Name,
			Email:       row.Email,
		},
		PasswordHash: row.PasswordHash,
	}, nil
}

func (r *StaffRepository) GetByID(ctx context.Context, id entities.StaffID) (*entities.Staff, error) {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return nil, err
	}
	row, err := r.queries.GetStaffByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return toStaffEntity(row), nil
}

func toStaffEntity(row db.Staff) *entities.Staff {
	var slackUserID *entities.SlackUserID
	if row.SlackUserID.Valid {
		sid := entities.SlackUserID(row.SlackUserID.String)
		slackUserID = &sid
	}
	return &entities.Staff{
		ID:          entities.StaffID(row.ID.String()),
		SlackUserID: slackUserID,
		Name:        row.Name,
		Email:       row.Email,
	}
}
