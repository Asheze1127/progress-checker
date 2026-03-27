package usecase

import (
	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// NewHandleNewQuestionUseCaseForTest creates a HandleNewQuestionUseCase for testing without DI.
func NewHandleNewQuestionUseCaseForTest(repo entities.QuestionRepository, sender *service.QuestionSender) *HandleNewQuestionUseCase {
	return &HandleNewQuestionUseCase{
		repo:   repo,
		sender: sender,
	}
}
