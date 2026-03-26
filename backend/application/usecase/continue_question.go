package usecase

import (
	"context"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// ContinueQuestionUsecase sets a question status to in_progress to trigger
// the follow-up conversation flow.
type ContinueQuestionUsecase struct {
	questionRepo entities.QuestionRepository
}

// NewContinueQuestionUsecase creates a new ContinueQuestionUsecase.
func NewContinueQuestionUsecase(repo entities.QuestionRepository) *ContinueQuestionUsecase {
	return &ContinueQuestionUsecase{questionRepo: repo}
}

// Execute sets the given question to in_progress. It is idempotent: if the
// question is already in progress, no update is performed.
func (u *ContinueQuestionUsecase) Execute(ctx context.Context, questionID entities.QuestionID) error {
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
