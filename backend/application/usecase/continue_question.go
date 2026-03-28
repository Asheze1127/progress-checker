package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// ContinueQuestionUseCase sets a question status to in_progress to trigger
// the follow-up conversation flow.
type ContinueQuestionUseCase struct {
	questionRepo entities.QuestionRepository
}

// NewContinueQuestionUseCase creates a new ContinueQuestionUseCase.
func NewContinueQuestionUseCase(repo entities.QuestionRepository) *ContinueQuestionUseCase {
	return &ContinueQuestionUseCase{questionRepo: repo}
}

// Execute sets the given question to in_progress. It is idempotent: if the
// question is already in progress, no update is performed.
func (u *ContinueQuestionUseCase) Execute(ctx context.Context, questionID entities.QuestionID) (err error) {
	defer func() {
		attrs := []slog.Attr{slog.String("question_id", string(questionID))}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
		slog.LogAttrs(ctx, slog.LevelDebug, "ContinueQuestionUseCase.Execute", attrs...)
	}()

	question, err := u.questionRepo.GetByID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("finding question %q: %w", questionID, err)
	}

	if question.Status == entities.QuestionStatusInProgress {
		return nil
	}

	if !question.CanTransitionTo(entities.QuestionStatusInProgress) {
		return fmt.Errorf("question %q cannot transition from %q to %q", questionID, question.Status, entities.QuestionStatusInProgress)
	}

	if err := u.questionRepo.UpdateStatus(ctx, questionID, entities.QuestionStatusInProgress); err != nil {
		return fmt.Errorf("updating question %q to in_progress: %w", questionID, err)
	}

	return nil
}
