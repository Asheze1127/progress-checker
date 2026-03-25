package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/application"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// --- Test doubles ---

type mockQuestionRepo struct {
	question    *entities.Question
	findErr     error
	updateErr   error
	updatedID     entities.QuestionID
	updatedStatus entities.QuestionStatus
}

func (r *mockQuestionRepo) FindByID(_ context.Context, id entities.QuestionID) (*entities.Question, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.question, nil
}

func (r *mockQuestionRepo) UpdateStatus(_ context.Context, id entities.QuestionID, status entities.QuestionStatus) error {
	r.updatedID = id
	r.updatedStatus = status
	return r.updateErr
}

type mockSlackNotifier struct {
	postErr error
	posted  bool
}

func (n *mockSlackNotifier) PostToMentorChannel(_ context.Context, q *entities.Question) error {
	n.posted = true
	return n.postErr
}

func newTestHandler(repo *mockQuestionRepo, notifier *mockSlackNotifier) *InteractionHandler {
	svc := application.NewQuestionActionService(repo, notifier)
	return NewInteractionHandler(svc)
}

func buildPayload(actionID, questionID string) string {
	return `{"type":"block_actions","actions":[{"action_id":"` + actionID + `","value":"` + questionID + `"}],"user":{"id":"U123","username":"testuser"}}`
}

func postInteraction(handler *InteractionHandler, payload string) *httptest.ResponseRecorder {
	form := url.Values{}
	form.Set("payload", payload)
	body := strings.NewReader(form.Encode())

	req := httptest.NewRequest(http.MethodPost, "/webhook/slack/interactions", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.HandleInteraction(rr, req)
	return rr
}

// --- Tests ---

func TestHandleInteractionResolved(t *testing.T) {
	repo := &mockQuestionRepo{
		question: &entities.Question{
			ID: "q-1", ParticipantID: "p-1", Title: "test",
			SlackChannelID: "C1", Status: entities.QuestionStatusOpen, SlackThreadTS: "123.456",
		},
	}
	notifier := &mockSlackNotifier{}
	handler := newTestHandler(repo, notifier)

	rr := postInteraction(handler, buildPayload("question_resolved", "q-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if repo.updatedStatus != entities.QuestionStatusResolved {
		t.Errorf("expected status %q, got %q", entities.QuestionStatusResolved, repo.updatedStatus)
	}
}

func TestHandleInteractionContinue(t *testing.T) {
	repo := &mockQuestionRepo{
		question: &entities.Question{
			ID: "q-1", ParticipantID: "p-1", Title: "test",
			SlackChannelID: "C1", Status: entities.QuestionStatusOpen, SlackThreadTS: "123.456",
		},
	}
	notifier := &mockSlackNotifier{}
	handler := newTestHandler(repo, notifier)

	rr := postInteraction(handler, buildPayload("question_continue", "q-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if repo.updatedStatus != entities.QuestionStatusInProgress {
		t.Errorf("expected status %q, got %q", entities.QuestionStatusInProgress, repo.updatedStatus)
	}
}

func TestHandleInteractionEscalate(t *testing.T) {
	repo := &mockQuestionRepo{
		question: &entities.Question{
			ID: "q-1", ParticipantID: "p-1", Title: "test",
			SlackChannelID: "C1", Status: entities.QuestionStatusOpen, SlackThreadTS: "123.456",
		},
	}
	notifier := &mockSlackNotifier{}
	handler := newTestHandler(repo, notifier)

	rr := postInteraction(handler, buildPayload("question_escalate", "q-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if repo.updatedStatus != entities.QuestionStatusAssignedMentor {
		t.Errorf("expected status %q, got %q", entities.QuestionStatusAssignedMentor, repo.updatedStatus)
	}
	if !notifier.posted {
		t.Error("expected PostToMentorChannel to be called")
	}
}

func TestHandleInteractionMethodNotAllowed(t *testing.T) {
	repo := &mockQuestionRepo{}
	notifier := &mockSlackNotifier{}
	handler := newTestHandler(repo, notifier)

	req := httptest.NewRequest(http.MethodGet, "/webhook/slack/interactions", nil)
	rr := httptest.NewRecorder()
	handler.HandleInteraction(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestHandleInteractionMissingPayload(t *testing.T) {
	repo := &mockQuestionRepo{}
	notifier := &mockSlackNotifier{}
	handler := newTestHandler(repo, notifier)

	req := httptest.NewRequest(http.MethodPost, "/webhook/slack/interactions", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.HandleInteraction(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleInteractionInvalidJSON(t *testing.T) {
	repo := &mockQuestionRepo{}
	notifier := &mockSlackNotifier{}
	handler := newTestHandler(repo, notifier)

	form := url.Values{}
	form.Set("payload", "not-json")
	body := strings.NewReader(form.Encode())

	req := httptest.NewRequest(http.MethodPost, "/webhook/slack/interactions", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.HandleInteraction(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleInteractionNoActions(t *testing.T) {
	repo := &mockQuestionRepo{}
	notifier := &mockSlackNotifier{}
	handler := newTestHandler(repo, notifier)

	rr := postInteraction(handler, `{"type":"block_actions","actions":[],"user":{"id":"U123"}}`)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestHandleInteractionUnknownAction(t *testing.T) {
	repo := &mockQuestionRepo{}
	notifier := &mockSlackNotifier{}
	handler := newTestHandler(repo, notifier)

	rr := postInteraction(handler, buildPayload("unknown_action", "q-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 for unknown action, got %d", rr.Code)
	}
}

func TestHandleInteractionMissingQuestionID(t *testing.T) {
	repo := &mockQuestionRepo{}
	notifier := &mockSlackNotifier{}
	handler := newTestHandler(repo, notifier)

	rr := postInteraction(handler, buildPayload("question_resolved", ""))

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestHandleInteractionServiceError(t *testing.T) {
	repo := &mockQuestionRepo{findErr: errors.New("db down")}
	notifier := &mockSlackNotifier{}
	handler := newTestHandler(repo, notifier)

	rr := postInteraction(handler, buildPayload("question_resolved", "q-1"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}
