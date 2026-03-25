package application

import (
	"context"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// QuestionRepository defines the persistence operations for questions.
type QuestionRepository interface {
	// FindByID retrieves a question by its ID.
	FindByID(ctx context.Context, id entities.QuestionID) (*entities.Question, error)
	// UpdateStatus persists the updated question status and mentor assignments.
	UpdateStatus(ctx context.Context, id entities.QuestionID, status entities.QuestionStatus) error
}

// SlackNotifier defines operations for sending Slack messages.
type SlackNotifier interface {
	// PostToMentorChannel posts a formatted question summary to the mentor channel.
	PostToMentorChannel(ctx context.Context, question *entities.Question) error
}

// QuestionActionService handles user actions on questions after AI responses.
type QuestionActionService struct {
	questionRepo  QuestionRepository
	slackNotifier SlackNotifier
}

// NewQuestionActionService creates a new QuestionActionService.
func NewQuestionActionService(repo QuestionRepository, notifier SlackNotifier) *QuestionActionService {
	return &QuestionActionService{
		questionRepo:  repo,
		slackNotifier: notifier,
	}
}

// ResolveQuestion marks a question as resolved.
func (s *QuestionActionService) ResolveQuestion(ctx context.Context, questionID entities.QuestionID) error {
	question, err := s.questionRepo.FindByID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("finding question %q: %w", questionID, err)
	}

	if question.Status == entities.QuestionStatusResolved {
		// Idempotent: already resolved, nothing to do.
		return nil
	}

	if err := s.questionRepo.UpdateStatus(ctx, questionID, entities.QuestionStatusResolved); err != nil {
		return fmt.Errorf("updating question %q to resolved: %w", questionID, err)
	}

	return nil
}

// ContinueQuestion sets a question status to in_progress to trigger the
// follow-up conversation flow.
func (s *QuestionActionService) ContinueQuestion(ctx context.Context, questionID entities.QuestionID) error {
	question, err := s.questionRepo.FindByID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("finding question %q: %w", questionID, err)
	}

	if question.Status == entities.QuestionStatusInProgress {
		// Idempotent: already in progress.
		return nil
	}

	if err := s.questionRepo.UpdateStatus(ctx, questionID, entities.QuestionStatusInProgress); err != nil {
		return fmt.Errorf("updating question %q to in_progress: %w", questionID, err)
	}

	return nil
}

// EscalateToMentor sets a question status to assigned_mentor and posts the
// question details to the mentor channel.
func (s *QuestionActionService) EscalateToMentor(ctx context.Context, questionID entities.QuestionID) error {
	question, err := s.questionRepo.FindByID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("finding question %q: %w", questionID, err)
	}

	if question.Status == entities.QuestionStatusAssignedMentor {
		// Idempotent: already escalated.
		return nil
	}

	if err := s.questionRepo.UpdateStatus(ctx, questionID, entities.QuestionStatusAssignedMentor); err != nil {
		return fmt.Errorf("updating question %q to assigned_mentor: %w", questionID, err)
	}

	if err := s.slackNotifier.PostToMentorChannel(ctx, question); err != nil {
		return fmt.Errorf("posting question %q to mentor channel: %w", questionID, err)
	}

	return nil
}
