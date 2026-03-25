package postgres

import (
	"testing"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

func TestQuestionRepositoryImplementsInterface(t *testing.T) {
	var _ entities.QuestionRepository = (*QuestionRepository)(nil)
}
