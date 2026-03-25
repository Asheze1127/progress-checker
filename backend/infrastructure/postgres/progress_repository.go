package postgres

import (
	"context"
	"database/sql"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

var _ entities.ProgressRepository = (*ProgressRepository)(nil)

// ProgressRepository persists progress logs to PostgreSQL.
type ProgressRepository struct {
	db *sql.DB
}

// NewProgressRepository creates a new ProgressRepository backed by the given database connection.
func NewProgressRepository(db *sql.DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

// Save persists a progress log and its bodies within a transaction.
// Validation is performed before persisting.
func (r *ProgressRepository) Save(ctx context.Context, log *entities.ProgressLog) error {
	if err := log.Validate(); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"INSERT INTO progress_logs (id, participant_id) VALUES ($1, $2)",
		string(log.ID), string(log.ParticipantID),
	)
	if err != nil {
		return err
	}

	for _, body := range log.ProgressBodies {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO progress_bodies (progress_log_id, phase, sos, comment, submitted_at) VALUES ($1, $2, $3, $4, $5)",
			string(log.ID), string(body.Phase), body.SOS, body.Comment, body.SubmittedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
