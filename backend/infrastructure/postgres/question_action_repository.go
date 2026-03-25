package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/Asheze1127/progress-checker/backend/infrastructure/sqlcgen"
	"github.com/google/uuid"
)

// Compile-time interface check.
var _ entities.QuestionRepository = (*QuestionActionRepository)(nil)

// QuestionActionRepository implements entities.QuestionRepository using PostgreSQL.
type QuestionActionRepository struct {
	queries *sqlcgen.Queries
}

// NewQuestionActionRepository creates a new QuestionActionRepository.
func NewQuestionActionRepository(db sqlcgen.DBTX) *QuestionActionRepository {
	return &QuestionActionRepository{queries: sqlcgen.New(db)}
}

// FindByID retrieves a question by its ID from the questions table.
func (r *QuestionActionRepository) FindByID(ctx context.Context, id entities.QuestionID) (*entities.Question, error) {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return nil, fmt.Errorf("invalid question ID %q: %w", id, err)
	}

	row, err := r.queries.FindQuestionByID(ctx, uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("question %q not found: %w", id, err)
		}
		return nil, fmt.Errorf("querying question %q: %w", id, err)
	}

	mentorUUIDs, err := r.queries.FindMentorIDsByQuestionID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("querying mentors for question %q: %w", id, err)
	}

	mentorIDs := make([]entities.MentorID, len(mentorUUIDs))
	for i, m := range mentorUUIDs {
		mentorIDs[i] = entities.MentorID(m.String())
	}

	return &entities.Question{
		ID:             entities.QuestionID(row.ID.String()),
		ParticipantID:  entities.ParticipantID(row.ParticipantID.String()),
		MentorIDs:      mentorIDs,
		Title:          row.Title,
		SlackChannelID: entities.SlackChannelID(row.SlackChannelID),
		Status:         entities.QuestionStatus(row.Status),
		SlackThreadTS:  row.SlackThreadTs,
	}, nil
}

// UpdateStatus updates the status of a question and sets updated_at to now().
func (r *QuestionActionRepository) UpdateStatus(ctx context.Context, id entities.QuestionID, status entities.QuestionStatus) error {
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return fmt.Errorf("invalid question ID %q: %w", id, err)
	}

	result, err := r.queries.UpdateQuestionStatus(ctx, sqlcgen.UpdateQuestionStatusParams{
		Status: sqlcgen.QuestionStatus(status),
		ID:     uid,
	})
	if err != nil {
		return fmt.Errorf("updating status of question %q: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected for question %q: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("question %q not found: %w", id, sql.ErrNoRows)
	}

	return nil
}
