package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// Compile-time interface check.
var _ entities.UserRepository = (*UserRepository)(nil)

// UserRepository implements entities.UserRepository using PostgreSQL.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL-backed UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByEmail retrieves a user with their password hash by email address.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.UserWithPassword, error) {
	query := `SELECT id, slack_user_id, name, email, role, password_hash FROM users WHERE email = $1`

	var user entities.UserWithPassword
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.SlackUserID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
	)
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return &user, nil
}

// FindByID retrieves a user by their ID.
func (r *UserRepository) FindByID(ctx context.Context, id entities.UserID) (*entities.User, error) {
	query := `SELECT id, slack_user_id, name, email, role FROM users WHERE id = $1`

	var user entities.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.SlackUserID,
		&user.Name,
		&user.Email,
		&user.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return &user, nil
}
