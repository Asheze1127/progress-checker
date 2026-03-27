package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do/v2"

	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// HandleNewQuestionInput contains the input data for creating a new question.
type HandleNewQuestionInput struct {
	ParticipantID  string
	Title          string
	Text           string
	SlackChannelID string
}

// HandleNewQuestionUseCase orchestrates creating a new question,
// persisting it, and enqueuing it for processing.
type HandleNewQuestionUseCase struct {
	repo   entities.QuestionRepository
	sender *service.QuestionSender
}

// NewHandleNewQuestionUseCase creates a new HandleNewQuestionUseCase via DI container.
func NewHandleNewQuestionUseCase(i do.Injector) (*HandleNewQuestionUseCase, error) {
	repo := do.MustInvoke[entities.QuestionRepository](i)
	sender := do.MustInvoke[*service.QuestionSender](i)
	return &HandleNewQuestionUseCase{
		repo:   repo,
		sender: sender,
	}, nil
}

// Execute creates a new question, saves it to the database, and enqueues it for processing.
func (uc *HandleNewQuestionUseCase) Execute(ctx context.Context, input HandleNewQuestionInput) error {
	questionID := uuid.New().String()
	threadTS := uuid.New().String()

	question := &entities.Question{
		ID:             entities.QuestionID(questionID),
		ParticipantID:  entities.ParticipantID(input.ParticipantID),
		Title:          input.Title,
		SlackChannelID: entities.SlackChannelID(input.SlackChannelID),
		Status:         entities.QuestionStatusOpen,
		SlackThreadTS:  threadTS,
	}

	if err := uc.repo.Save(ctx, question); err != nil {
		return fmt.Errorf("failed to save question: %w", err)
	}

	msg := service.QuestionNewMessage{
		QuestionID:     questionID,
		ParticipantID:  input.ParticipantID,
		Title:          input.Title,
		Text:           input.Text,
		SlackChannelID: input.SlackChannelID,
		SlackThreadTS:  threadTS,
	}

	if err := uc.sender.SendNewQuestion(ctx, msg); err != nil {
		return fmt.Errorf("failed to send question to queue: %w", err)
	}

	return nil
}
