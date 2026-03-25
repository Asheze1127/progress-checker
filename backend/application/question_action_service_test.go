package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// --- Test doubles ---

type stubQuestionRepo struct {
	question    *entities.Question
	findErr     error
	updateErr   error
	updatedID     entities.QuestionID
	updatedStatus entities.QuestionStatus
}

func (r *stubQuestionRepo) FindByID(_ context.Context, id entities.QuestionID) (*entities.Question, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.question, nil
}

func (r *stubQuestionRepo) UpdateStatus(_ context.Context, id entities.QuestionID, status entities.QuestionStatus) error {
	r.updatedID = id
	r.updatedStatus = status
	return r.updateErr
}

type stubSlackNotifier struct {
	postErr  error
	posted   bool
	question *entities.Question
}

func (n *stubSlackNotifier) PostToMentorChannel(_ context.Context, q *entities.Question) error {
	n.posted = true
	n.question = q
	return n.postErr
}

func newTestQuestion(status entities.QuestionStatus) *entities.Question {
	return &entities.Question{
		ID:             "q-1",
		ParticipantID:  "p-1",
		Title:          "Test question",
		SlackChannelID: "C123",
		Status:         status,
		SlackThreadTS:  "1700000000.000100",
	}
}

// --- ResolveQuestion tests ---

func TestResolveQuestion(t *testing.T) {
	repo := &stubQuestionRepo{question: newTestQuestion(entities.QuestionStatusOpen)}
	notifier := &stubSlackNotifier{}
	svc := NewQuestionActionService(repo, notifier)

	err := svc.ResolveQuestion(context.Background(), "q-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updatedStatus != entities.QuestionStatusResolved {
		t.Errorf("expected status %q, got %q", entities.QuestionStatusResolved, repo.updatedStatus)
	}
}

func TestResolveQuestionIdempotent(t *testing.T) {
	repo := &stubQuestionRepo{question: newTestQuestion(entities.QuestionStatusResolved)}
	notifier := &stubSlackNotifier{}
	svc := NewQuestionActionService(repo, notifier)

	err := svc.ResolveQuestion(context.Background(), "q-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// UpdateStatus should not have been called.
	if repo.updatedStatus != "" {
		t.Errorf("expected no update, but status was set to %q", repo.updatedStatus)
	}
}

func TestResolveQuestionFindError(t *testing.T) {
	repo := &stubQuestionRepo{findErr: errors.New("not found")}
	notifier := &stubSlackNotifier{}
	svc := NewQuestionActionService(repo, notifier)

	err := svc.ResolveQuestion(context.Background(), "q-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolveQuestionUpdateError(t *testing.T) {
	repo := &stubQuestionRepo{
		question:  newTestQuestion(entities.QuestionStatusOpen),
		updateErr: errors.New("db error"),
	}
	notifier := &stubSlackNotifier{}
	svc := NewQuestionActionService(repo, notifier)

	err := svc.ResolveQuestion(context.Background(), "q-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- ContinueQuestion tests ---

func TestContinueQuestion(t *testing.T) {
	repo := &stubQuestionRepo{question: newTestQuestion(entities.QuestionStatusOpen)}
	notifier := &stubSlackNotifier{}
	svc := NewQuestionActionService(repo, notifier)

	err := svc.ContinueQuestion(context.Background(), "q-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updatedStatus != entities.QuestionStatusInProgress {
		t.Errorf("expected status %q, got %q", entities.QuestionStatusInProgress, repo.updatedStatus)
	}
}

func TestContinueQuestionIdempotent(t *testing.T) {
	repo := &stubQuestionRepo{question: newTestQuestion(entities.QuestionStatusInProgress)}
	notifier := &stubSlackNotifier{}
	svc := NewQuestionActionService(repo, notifier)

	err := svc.ContinueQuestion(context.Background(), "q-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.updatedStatus != "" {
		t.Errorf("expected no update, but status was set to %q", repo.updatedStatus)
	}
}

// --- EscalateToMentor tests ---

func TestEscalateToMentor(t *testing.T) {
	repo := &stubQuestionRepo{question: newTestQuestion(entities.QuestionStatusOpen)}
	notifier := &stubSlackNotifier{}
	svc := NewQuestionActionService(repo, notifier)

	err := svc.EscalateToMentor(context.Background(), "q-1")
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
	svc := NewQuestionActionService(repo, notifier)

	err := svc.EscalateToMentor(context.Background(), "q-1")
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
	svc := NewQuestionActionService(repo, notifier)

	err := svc.EscalateToMentor(context.Background(), "q-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
