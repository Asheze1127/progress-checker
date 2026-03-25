package entities

import "context"

// QuestionRepository defines the persistence interface for Question entities.
type QuestionRepository interface {
	Save(ctx context.Context, q *Question) error
	FindByThreadTS(ctx context.Context, channelID, threadTS string) (*Question, error)
}
