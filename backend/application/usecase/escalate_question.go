package usecase

import (
	"context"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/Asheze1127/progress-checker/backend/application/service"
)

// EscalateQuestionUseCase sets a question status to assigned_mentor and
// notifies the mentor channel.
type EscalateQuestionUseCase struct {
	questionRepo  entities.QuestionRepository
	slackNotifier service.SlackNotifier
}

// NewEscalateQuestionUseCase creates a new EscalateQuestionUseCase.
func NewEscalateQuestionUseCase(repo entities.QuestionRepository, notifier service.SlackNotifier) *EscalateQuestionUseCase {
	return &EscalateQuestionUseCase{
		questionRepo:  repo,
		slackNotifier: notifier,
	}
}

// Execute escalates the given question to a mentor. It is idempotent: if the
// question is already assigned to a mentor, no update is performed.
func (u *EscalateQuestionUseCase) Execute(ctx context.Context, questionID entities.QuestionID) error {
	question, err := u.questionRepo.GetByID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("finding question %q: %w", questionID, err)
	}

	if question.Status == entities.QuestionStatusAssignedMentor {
		return nil
	}

	if !question.CanTransitionTo(entities.QuestionStatusAssignedMentor) {
		return fmt.Errorf("question %q cannot transition from %q to %q", questionID, question.Status, entities.QuestionStatusAssignedMentor)
	}

	// Notify first, then persist — if notification fails, we don't leave
	// the question in assigned_mentor status without a notification.
	if err := u.slackNotifier.PostToMentorChannel(ctx, question); err != nil {
		return fmt.Errorf("posting question %q to mentor channel: %w", questionID, err)
	}

	if err := u.questionRepo.UpdateStatus(ctx, questionID, entities.QuestionStatusAssignedMentor); err != nil {
		return fmt.Errorf("updating question %q to assigned_mentor: %w", questionID, err)
	}

	return nil
}
