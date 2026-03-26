package usecase

import (
	"context"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// ResolveQuestionUsecase marks a question as resolved.
type ResolveQuestionUsecase struct {
	questionRepo entities.QuestionRepository
}

// NewResolveQuestionUsecase creates a new ResolveQuestionUsecase.
func NewResolveQuestionUsecase(repo entities.QuestionRepository) *ResolveQuestionUsecase {
	return &ResolveQuestionUsecase{questionRepo: repo}
}

// Execute marks the given question as resolved. It is idempotent: if the
// question is already resolved, no update is performed.
func (u *ResolveQuestionUsecase) Execute(ctx context.Context, questionID entities.QuestionID) error {
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
