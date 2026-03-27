package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

func TestContinueQuestion(t *testing.T) {
	repo := &stubQuestionRepo{question: newTestQuestion(entities.QuestionStatusOpen)}
	uc := NewContinueQuestionUseCase(repo)

	err := uc.Execute(context.Background(), "q-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updatedStatus != entities.QuestionStatusInProgress {
		t.Errorf("expected status %q, got %q", entities.QuestionStatusInProgress, repo.updatedStatus)
	}
}

func TestContinueQuestionIdempotent(t *testing.T) {
	repo := &stubQuestionRepo{question: newTestQuestion(entities.QuestionStatusInProgress)}
	uc := NewContinueQuestionUseCase(repo)

	err := uc.Execute(context.Background(), "q-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updatedStatus != "" {
		t.Errorf("expected no update, but status was set to %q", repo.updatedStatus)
	}
}

func TestContinueQuestionFindError(t *testing.T) {
	repo := &stubQuestionRepo{findErr: errors.New("not found")}
	uc := NewContinueQuestionUseCase(repo)

	err := uc.Execute(context.Background(), "q-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
