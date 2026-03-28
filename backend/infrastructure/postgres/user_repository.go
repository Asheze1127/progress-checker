package postgres

import (
	"context"
	"database/sql"
	"fmt"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

var _ entities.UserRepository = (*UserRepository)(nil)

// UserRepository queries users from PostgreSQL.
type UserRepository struct {
	db      *sql.DB
	queries *db.Queries
}

// NewUserRepository creates a new UserRepository backed by the given database connection.
func NewUserRepository(database *sql.DB) *UserRepository {
	return &UserRepository{
		db:      database,
		queries: db.New(database),
	}
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

// FindByEmail retrieves a user with their password hash by email address.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.UserWithPassword, error) {
	row, err := r.queries.GetUserWithPasswordByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return &entities.UserWithPassword{
		User: entities.User{
			ID:          entities.UserID(row.ID.String()),
			SlackUserID: entities.SlackUserID(row.SlackUserID),
			Name:        row.Name,
			Email:       row.Email,
			Role:        entities.UserRole(row.Role),
		},
		PasswordHash: row.PasswordHash,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User, passwordHash string) (*entities.User, error) {
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		SlackUserID:  string(user.SlackUserID),
		Name:         user.Name,
		Email:        user.Email,
		Role:         string(user.Role),
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return toUserEntity(row), nil
}

func (r *UserRepository) UpdatePasswordHash(ctx context.Context, id entities.UserID, passwordHash string) error {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return err
	}
	return r.queries.UpdateUserPasswordHash(ctx, db.UpdateUserPasswordHashParams{
		PasswordHash: passwordHash,
		ID:           uid,
	})
}

func (r *UserRepository) List(ctx context.Context) ([]*entities.User, error) {
	rows, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]*entities.User, len(rows))
	for i, row := range rows {
		users[i] = toUserEntity(row)
	}
	return users, nil
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
