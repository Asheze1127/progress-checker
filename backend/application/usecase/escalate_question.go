package usecase

import (
	"context"
	"fmt"

	"github.com/samber/do/v2"

	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// EscalateQuestionUseCase sets a question status to assigned_mentor and
// notifies the mentor channel.
type EscalateQuestionUseCase struct {
	questionRepo  entities.QuestionRepository
	slackNotifier service.SlackNotifier
}

// NewEscalateQuestionUseCase creates a new EscalateQuestionUseCase via DI container.
func NewEscalateQuestionUseCase(i do.Injector) (*EscalateQuestionUseCase, error) {
	repo := do.MustInvoke[entities.QuestionRepository](i)
	notifier := do.MustInvoke[service.SlackNotifier](i)
	return &EscalateQuestionUseCase{
		questionRepo:  repo,
		slackNotifier: notifier,
	}, nil
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

	if err := u.questionRepo.UpdateStatus(ctx, questionID, entities.QuestionStatusAssignedMentor); err != nil {
		return fmt.Errorf("updating question %q to assigned_mentor: %w", questionID, err)
	}

	if err := u.slackNotifier.PostToMentorChannel(ctx, question); err != nil {
		return fmt.Errorf("posting question %q to mentor channel: %w", questionID, err)
	}

	return nil
}
