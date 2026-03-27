package usecase

import (
	"context"
	"fmt"

	"github.com/samber/do/v2"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// ResolveQuestionUseCase marks a question as resolved.
type ResolveQuestionUseCase struct {
	questionRepo entities.QuestionRepository
}

// NewResolveQuestionUseCase creates a new ResolveQuestionUseCase via DI container.
func NewResolveQuestionUseCase(i do.Injector) (*ResolveQuestionUseCase, error) {
	repo := do.MustInvoke[entities.QuestionRepository](i)
	return &ResolveQuestionUseCase{questionRepo: repo}, nil
}

// Execute marks the given question as resolved. It is idempotent: if the
// question is already resolved, no update is performed.
func (u *ResolveQuestionUseCase) Execute(ctx context.Context, questionID entities.QuestionID) error {
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
