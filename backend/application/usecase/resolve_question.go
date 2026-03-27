package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// ResolveQuestionUseCase marks a question as resolved.
type ResolveQuestionUseCase struct {
	questionRepo entities.QuestionRepository
}

// NewResolveQuestionUseCase creates a new ResolveQuestionUseCase.
func NewResolveQuestionUseCase(repo entities.QuestionRepository) *ResolveQuestionUseCase {
	return &ResolveQuestionUseCase{questionRepo: repo}
}

// Execute marks the given question as resolved. It is idempotent: if the
// question is already resolved, no update is performed.
func (u *ResolveQuestionUseCase) Execute(ctx context.Context, questionID entities.QuestionID) (err error) {
	defer func() {
		attrs := []slog.Attr{slog.String("question_id", string(questionID))}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
		slog.LogAttrs(ctx, slog.LevelDebug, "ResolveQuestionUseCase.Execute", attrs...)
	}()

	question, err := u.questionRepo.GetByID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("finding question %q: %w", questionID, err)
	}

	if question.Status == entities.QuestionStatusResolved {
		return nil
	}

	if !question.CanTransitionTo(entities.QuestionStatusResolved) {
		return fmt.Errorf("question %q cannot transition from %q to %q", questionID, question.Status, entities.QuestionStatusResolved)
	}

	if err := u.questionRepo.UpdateStatus(ctx, questionID, entities.QuestionStatusResolved); err != nil {
		return fmt.Errorf("updating question %q to resolved: %w", questionID, err)
	}

	return nil
}
