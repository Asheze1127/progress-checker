package postgres

import (
	"context"
	"database/sql"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

var _ entities.UserRepository = (*UserRepository)(nil)

// UserRepository queries users from PostgreSQL.
type UserRepository struct {
	queries *db.Queries
}

// NewUserRepository creates a new UserRepository backed by the given database connection.
func NewUserRepository(database *sql.DB) *UserRepository {
	return &UserRepository{queries: db.New(database)}
}

func (r *UserRepository) GetByID(ctx context.Context, id entities.UserID) (*entities.User, error) {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return nil, err
	}
	row, err := r.queries.GetUserByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return toUserEntity(row), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return toUserEntity(row), nil
}

func (r *UserRepository) GetBySlackUserID(ctx context.Context, slackUserID entities.SlackUserID) (*entities.User, error) {
	row, err := r.queries.GetUserBySlackUserID(ctx, string(slackUserID))
	if err != nil {
		return nil, err
	}
	return toUserEntity(row), nil
}

func toUserEntity(row db.Users) *entities.User {
	return &entities.User{
		ID:          entities.UserID(row.ID.String()),
		SlackUserID: entities.SlackUserID(row.SlackUserID),
		Name:        row.Name,
		Email:       row.Email,
		Role:        entities.UserRole(row.Role),
	}
}
