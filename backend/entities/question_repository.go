package entities

import "context"

// QuestionRepository defines the persistence operations for questions.
type QuestionRepository interface {
	// FindByID retrieves a question by its ID.
	FindByID(ctx context.Context, id QuestionID) (*Question, error)
	// UpdateStatus persists the updated question status.
	UpdateStatus(ctx context.Context, id QuestionID, status QuestionStatus) error
}
