package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

func TestEscalateToMentor(t *testing.T) {
	repo := &stubQuestionRepo{question: newTestQuestion(entities.QuestionStatusOpen)}
	notifier := &stubSlackNotifier{}
	uc := &EscalateQuestionUseCase{questionRepo: repo, slackNotifier: notifier}

	err := uc.Execute(context.Background(), "q-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updatedStatus != entities.QuestionStatusAssignedMentor {
		t.Errorf("expected status %q, got %q", entities.QuestionStatusAssignedMentor, repo.updatedStatus)
	}
	if !notifier.posted {
		t.Error("expected PostToMentorChannel to be called")
	}
}

func TestEscalateToMentorIdempotent(t *testing.T) {
	repo := &stubQuestionRepo{question: newTestQuestion(entities.QuestionStatusAssignedMentor)}
	notifier := &stubSlackNotifier{}
	uc := &EscalateQuestionUseCase{questionRepo: repo, slackNotifier: notifier}

	err := uc.Execute(context.Background(), "q-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updatedStatus != "" {
		t.Errorf("expected no update, but status was set to %q", repo.updatedStatus)
	}
	if notifier.posted {
		t.Error("expected PostToMentorChannel not to be called")
	}
}

func TestEscalateToMentorNotifyError(t *testing.T) {
	repo := &stubQuestionRepo{question: newTestQuestion(entities.QuestionStatusOpen)}
	notifier := &stubSlackNotifier{postErr: errors.New("slack error")}
	uc := &EscalateQuestionUseCase{questionRepo: repo, slackNotifier: notifier}

	err := uc.Execute(context.Background(), "q-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
