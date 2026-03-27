package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

var _ entities.SetupTokenRepository = (*SetupTokenRepository)(nil)

// SetupTokenRepository manages setup tokens in PostgreSQL.
type SetupTokenRepository struct {
	queries *db.Queries
}

// NewSetupTokenRepository creates a new SetupTokenRepository backed by the given database connection.
func NewSetupTokenRepository(database *sql.DB) *SetupTokenRepository {
	return &SetupTokenRepository{queries: db.New(database)}
}

func (r *SetupTokenRepository) Create(ctx context.Context, userID entities.UserID, tokenHash string, expiresAt time.Time) (*entities.SetupToken, error) {
	uid, err := uuid.Parse(string(userID))
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	row, err := r.queries.CreateSetupToken(ctx, db.CreateSetupTokenParams{
		UserID:    uid,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("create setup token: %w", err)
	}
	return toSetupTokenEntity(row), nil
}

func (r *SetupTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*entities.SetupToken, error) {
	row, err := r.queries.GetSetupTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("find setup token by hash: %w", err)
	}
	return toSetupTokenEntity(row), nil
}

func (r *SetupTokenRepository) MarkUsed(ctx context.Context, id entities.SetupTokenID) error {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return fmt.Errorf("parse setup token id: %w", err)
	}
	return r.queries.MarkSetupTokenUsed(ctx, uid)
}

func toSetupTokenEntity(row db.SetupTokens) *entities.SetupToken {
	var usedAt *time.Time
	if row.UsedAt.Valid {
		usedAt = &row.UsedAt.Time
	}
	return &entities.SetupToken{
		ID:        entities.SetupTokenID(row.ID.String()),
		UserID:    entities.UserID(row.UserID.String()),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		UsedAt:    usedAt,
	}
}
