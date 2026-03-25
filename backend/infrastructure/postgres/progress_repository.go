package postgres

import (
	"context"
	"database/sql"
	"fmt"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

var _ entities.ProgressRepository = (*ProgressRepository)(nil)

// ProgressRepository persists progress logs to PostgreSQL.
type ProgressRepository struct {
	db      *sql.DB
	queries *db.Queries
}

// NewProgressRepository creates a new ProgressRepository backed by the given database connection.
func NewProgressRepository(database *sql.DB) *ProgressRepository {
	return &ProgressRepository{
		db:      database,
		queries: db.New(database),
	}
}

// Save persists a progress log and its bodies within a transaction.
// Validation is performed before persisting.
func (r *ProgressRepository) Save(ctx context.Context, log *entities.ProgressLog) error {
	if err := log.Validate(); err != nil {
		return err
	}

	logID, err := uuid.Parse(string(log.ID))
	if err != nil {
		return fmt.Errorf("invalid progress log ID %q: %w", log.ID, err)
	}
	participantID, err := uuid.Parse(string(log.ParticipantID))
	if err != nil {
		return fmt.Errorf("invalid participant ID %q: %w", log.ParticipantID, err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	_, err = qtx.InsertProgressLog(ctx, db.InsertProgressLogParams{
		ID:            logID,
		ParticipantID: participantID,
	})
	if err != nil {
		return fmt.Errorf("inserting progress log %q: %w", log.ID, err)
	}

	for _, body := range log.ProgressBodies {
		bodyID := uuid.New()
		_, err = qtx.InsertProgressBody(ctx, db.InsertProgressBodyParams{
			ID:            bodyID,
			ProgressLogID: logID,
			Phase:         string(body.Phase),
			Sos:           body.SOS,
			Comment:       sql.NullString{String: body.Comment, Valid: body.Comment != ""},
			SubmittedAt:   body.SubmittedAt,
		})
		if err != nil {
			return fmt.Errorf("inserting progress body (phase=%s) for log %q: %w", body.Phase, log.ID, err)
		}
	}

	return tx.Commit()
}
