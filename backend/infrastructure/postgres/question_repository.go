package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// Compile-time check that QuestionRepository implements entities.QuestionRepository.
var _ entities.QuestionRepository = (*QuestionRepository)(nil)

// QuestionRepository is a PostgreSQL-backed implementation of entities.QuestionRepository.
type QuestionRepository struct {
	db *sql.DB
}

// NewQuestionRepository creates a new QuestionRepository with the given database connection.
func NewQuestionRepository(db *sql.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

// Save validates and persists a Question entity using a transaction.
func (r *QuestionRepository) Save(ctx context.Context, q *entities.Question) error {
	if err := q.Validate(); err != nil {
		return fmt.Errorf("invalid question: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	mentorIDs := make([]string, len(q.MentorIDs))
	for i, id := range q.MentorIDs {
		mentorIDs[i] = string(id)
	}

	const query = `INSERT INTO questions (id, participant_id, mentor_ids, title, slack_channel_id, status, slack_thread_ts)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = tx.ExecContext(ctx, query,
		string(q.ID),
		string(q.ParticipantID),
		"{"+strings.Join(mentorIDs, ",")+"}",
		q.Title,
		string(q.SlackChannelID),
		string(q.Status),
		q.SlackThreadTS,
	)
	if err != nil {
		return fmt.Errorf("failed to insert question: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByThreadTS retrieves a Question by its Slack channel ID and thread timestamp.
func (r *QuestionRepository) FindByThreadTS(ctx context.Context, channelID, threadTS string) (*entities.Question, error) {
	const query = `SELECT id, participant_id, mentor_ids, title, slack_channel_id, status, slack_thread_ts
		FROM questions
		WHERE slack_channel_id = $1 AND slack_thread_ts = $2`

	var (
		id            string
		participantID string
		mentorIDsRaw  string
		title         string
		slackChannel  string
		status        string
		slackThreadTS string
	)

	err := r.db.QueryRowContext(ctx, query, channelID, threadTS).Scan(
		&id,
		&participantID,
		&mentorIDsRaw,
		&title,
		&slackChannel,
		&status,
		&slackThreadTS,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query question: %w", err)
	}

	mentorIDs := parseMentorIDs(mentorIDsRaw)

	return &entities.Question{
		ID:             entities.QuestionID(id),
		ParticipantID:  entities.ParticipantID(participantID),
		MentorIDs:      mentorIDs,
		Title:          title,
		SlackChannelID: entities.SlackChannelID(slackChannel),
		Status:         entities.QuestionStatus(status),
		SlackThreadTS:  slackThreadTS,
	}, nil
}

// parseMentorIDs converts a PostgreSQL array string like "{a,b,c}" into a slice of MentorID.
func parseMentorIDs(raw string) []entities.MentorID {
	trimmed := strings.Trim(raw, "{}")
	if trimmed == "" {
		return nil
	}

	parts := strings.Split(trimmed, ",")
	mentorIDs := make([]entities.MentorID, 0, len(parts))
	for _, p := range parts {
		cleaned := strings.TrimSpace(p)
		if cleaned != "" {
			mentorIDs = append(mentorIDs, entities.MentorID(cleaned))
		}
	}

	return mentorIDs
}
