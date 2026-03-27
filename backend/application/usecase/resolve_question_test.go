package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

func TestResolveQuestion(t *testing.T) {
	repo := &stubQuestionRepo{question: newTestQuestion(entities.QuestionStatusOpen)}
	uc := NewResolveQuestionUseCase(repo)

	err := uc.Execute(context.Background(), "q-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updatedStatus != entities.QuestionStatusResolved {
		t.Errorf("expected status %q, got %q", entities.QuestionStatusResolved, repo.updatedStatus)
	}
}

func TestResolveQuestionIdempotent(t *testing.T) {
	repo := &stubQuestionRepo{question: newTestQuestion(entities.QuestionStatusResolved)}
	uc := NewResolveQuestionUseCase(repo)

	err := uc.Execute(context.Background(), "q-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updatedStatus != "" {
		t.Errorf("expected no update, but status was set to %q", repo.updatedStatus)
	}
}

func TestResolveQuestionFindError(t *testing.T) {
	repo := &stubQuestionRepo{findErr: errors.New("not found")}
	uc := NewResolveQuestionUseCase(repo)

	err := uc.Execute(context.Background(), "q-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolveQuestionUpdateError(t *testing.T) {
	repo := &stubQuestionRepo{
		question:  newTestQuestion(entities.QuestionStatusOpen),
		updateErr: errors.New("db error"),
	}
	uc := NewResolveQuestionUseCase(repo)

	err := uc.Execute(context.Background(), "q-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
