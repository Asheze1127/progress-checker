package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// Compile-time interface check.
var _ entities.QuestionRepository = (*QuestionActionRepository)(nil)

// QuestionActionRepository implements entities.QuestionRepository using PostgreSQL.
type QuestionActionRepository struct {
	db *sql.DB
}

// NewQuestionActionRepository creates a new QuestionActionRepository.
func NewQuestionActionRepository(db *sql.DB) *QuestionActionRepository {
	return &QuestionActionRepository{db: db}
}

// FindByID retrieves a question by its ID from the questions table.
func (r *QuestionActionRepository) FindByID(ctx context.Context, id entities.QuestionID) (*entities.Question, error) {
	const query = `SELECT id, participant_id, title, slack_channel_id, status, slack_thread_ts FROM questions WHERE id = $1`

	var question entities.Question
	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
		&question.ID,
		&question.ParticipantID,
		&question.Title,
		&question.SlackChannelID,
		&question.Status,
		&question.SlackThreadTS,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("question %q not found", id)
		}
		return nil, fmt.Errorf("querying question %q: %w", id, err)
	}

	mentorIDs, err := r.findMentorIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	question.MentorIDs = mentorIDs

	return &question, nil
}

// findMentorIDs retrieves the mentor IDs associated with a question.
func (r *QuestionActionRepository) findMentorIDs(ctx context.Context, questionID entities.QuestionID) ([]entities.MentorID, error) {
	const query = `SELECT mentor_id FROM question_mentors WHERE question_id = $1`

	rows, err := r.db.QueryContext(ctx, query, string(questionID))
	if err != nil {
		return nil, fmt.Errorf("querying mentors for question %q: %w", questionID, err)
	}
	defer rows.Close()

	var mentorIDs []entities.MentorID
	for rows.Next() {
		var mentorID entities.MentorID
		if err := rows.Scan(&mentorID); err != nil {
			return nil, fmt.Errorf("scanning mentor ID for question %q: %w", questionID, err)
		}
		mentorIDs = append(mentorIDs, mentorID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating mentors for question %q: %w", questionID, err)
	}

	return mentorIDs, nil
}

// UpdateStatus updates the status of a question and sets updated_at to now().
func (r *QuestionActionRepository) UpdateStatus(ctx context.Context, id entities.QuestionID, status entities.QuestionStatus) error {
	const query = `UPDATE questions SET status = $1, updated_at = now() WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, string(status), string(id))
	if err != nil {
		return fmt.Errorf("updating status of question %q: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected for question %q: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("question %q not found", id)
	}

	return nil
}
