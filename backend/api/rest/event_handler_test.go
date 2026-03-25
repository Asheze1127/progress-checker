package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	slackpkg "github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

// --- Test doubles ---

type spyIssueTrigger struct {
	calledWith usecase.TriggerIssueCreationInput
	called     bool
	err        error
}

func (s *spyIssueTrigger) Execute(_ context.Context, input usecase.TriggerIssueCreationInput) error {
	s.called = true
	s.calledWith = input
	return s.err
}

// --- Helpers ---

func newEventHandler(trigger IssueTrigger, emoji string) *EventHandler {
	if emoji == "" {
		emoji = defaultTriggerEmoji
	}
	return &EventHandler{
		triggerEmoji: emoji,
		issueTrigger: trigger,
	}
}

func postJSON(handler http.HandlerFunc, body interface{}) *httptest.ResponseRecorder {
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/webhook/slack/events", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// --- Tests ---

func TestURLVerification(t *testing.T) {
	t.Parallel()

	spy := &spyIssueTrigger{}
	handler := newEventHandler(spy, "ticket")

	event := slackpkg.SlackEvent{
		Type:      slackpkg.EventTypeURLVerification,
		Challenge: "test-challenge-value",
	}

	rr := postJSON(handler.HandleSlackEvents, event)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp slackpkg.URLVerificationResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Challenge != "test-challenge-value" {
		t.Errorf("expected challenge 'test-challenge-value', got %q", resp.Challenge)
	}

	if spy.called {
		t.Error("issue trigger should not be called for url_verification")
	}
}

func TestReactionAdded_TriggerEmoji(t *testing.T) {
	t.Parallel()

	spy := &spyIssueTrigger{}
	handler := newEventHandler(spy, "ticket")

	event := slackpkg.SlackEvent{
		Type: slackpkg.EventTypeEventCallback,
		Event: slackpkg.EventCallback{
			Type:     slackpkg.EventTypeReactionAdded,
			User:     "U123",
			Reaction: "ticket",
			Item: slackpkg.ReactionItem{
				Type:    "message",
				Channel: "C456",
				TS:      "1234567890.123456",
			},
		},
	}

	rr := postJSON(handler.HandleSlackEvents, event)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if !spy.called {
		t.Fatal("expected issue trigger to be called")
	}

	if spy.calledWith.ChannelID != "C456" {
		t.Errorf("expected channel_id 'C456', got %q", spy.calledWith.ChannelID)
	}
	if spy.calledWith.ThreadTS != "1234567890.123456" {
		t.Errorf("expected thread_ts '1234567890.123456', got %q", spy.calledWith.ThreadTS)
	}
	if spy.calledWith.TriggerUserID != "U123" {
		t.Errorf("expected trigger_user_id 'U123', got %q", spy.calledWith.TriggerUserID)
	}
	if spy.calledWith.TriggerType != "reaction" {
		t.Errorf("expected trigger_type 'reaction', got %q", spy.calledWith.TriggerType)
	}
}

func TestReactionAdded_NonTriggerEmoji(t *testing.T) {
	t.Parallel()

	spy := &spyIssueTrigger{}
	handler := newEventHandler(spy, "ticket")

	event := slackpkg.SlackEvent{
		Type: slackpkg.EventTypeEventCallback,
		Event: slackpkg.EventCallback{
			Type:     slackpkg.EventTypeReactionAdded,
			User:     "U123",
			Reaction: "thumbsup",
			Item: slackpkg.ReactionItem{
				Type:    "message",
				Channel: "C456",
				TS:      "1234567890.123456",
			},
		},
	}

	rr := postJSON(handler.HandleSlackEvents, event)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if spy.called {
		t.Error("issue trigger should not be called for non-trigger emoji")
	}
}

func TestMessageAction(t *testing.T) {
	t.Parallel()

	spy := &spyIssueTrigger{}
	handler := newEventHandler(spy, "ticket")

	event := slackpkg.SlackEvent{
		Type:       slackpkg.EventTypeMessageAction,
		CallbackID: "create_issue",
		TriggerID:  "trigger123",
		User:       slackpkg.IDField{ID: "U789"},
		Channel:    slackpkg.IDField{ID: "C456"},
		Message: slackpkg.Message{
			Text:     "We should track this",
			TS:       "1234567890.000001",
			ThreadTS: "1234567890.000000",
		},
	}

	rr := postJSON(handler.HandleSlackEvents, event)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if !spy.called {
		t.Fatal("expected issue trigger to be called")
	}

	if spy.calledWith.ChannelID != "C456" {
		t.Errorf("expected channel_id 'C456', got %q", spy.calledWith.ChannelID)
	}
	if spy.calledWith.ThreadTS != "1234567890.000000" {
		t.Errorf("expected thread_ts '1234567890.000000' (thread_ts), got %q", spy.calledWith.ThreadTS)
	}
	if spy.calledWith.TriggerUserID != "U789" {
		t.Errorf("expected trigger_user_id 'U789', got %q", spy.calledWith.TriggerUserID)
	}
	if spy.calledWith.TriggerType != "message_action" {
		t.Errorf("expected trigger_type 'message_action', got %q", spy.calledWith.TriggerType)
	}
}

func TestMessageAction_NoThreadTS_FallsBackToMessageTS(t *testing.T) {
	t.Parallel()

	spy := &spyIssueTrigger{}
	handler := newEventHandler(spy, "ticket")

	event := slackpkg.SlackEvent{
		Type:    slackpkg.EventTypeMessageAction,
		User:    slackpkg.IDField{ID: "U789"},
		Channel: slackpkg.IDField{ID: "C456"},
		Message: slackpkg.Message{
			Text: "A standalone message",
			TS:   "1234567890.000099",
		},
	}

	rr := postJSON(handler.HandleSlackEvents, event)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if spy.calledWith.ThreadTS != "1234567890.000099" {
		t.Errorf("expected thread_ts to fall back to message TS '1234567890.000099', got %q", spy.calledWith.ThreadTS)
	}
}

func TestReactionAdded_TriggerError(t *testing.T) {
	t.Parallel()

	spy := &spyIssueTrigger{err: fmt.Errorf("sqs failure")}
	handler := newEventHandler(spy, "ticket")

	event := slackpkg.SlackEvent{
		Type: slackpkg.EventTypeEventCallback,
		Event: slackpkg.EventCallback{
			Type:     slackpkg.EventTypeReactionAdded,
			User:     "U123",
			Reaction: "ticket",
			Item: slackpkg.ReactionItem{
				Type:    "message",
				Channel: "C456",
				TS:      "1234567890.123456",
			},
		},
	}

	rr := postJSON(handler.HandleSlackEvents, event)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	t.Parallel()

	spy := &spyIssueTrigger{}
	handler := newEventHandler(spy, "ticket")

	req := httptest.NewRequest(http.MethodGet, "/webhook/slack/events", nil)
	rr := httptest.NewRecorder()
	handler.HandleSlackEvents(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestInvalidJSON(t *testing.T) {
	t.Parallel()

	spy := &spyIssueTrigger{}
	handler := newEventHandler(spy, "ticket")

	req := httptest.NewRequest(http.MethodPost, "/webhook/slack/events", bytes.NewReader([]byte("not json")))
	rr := httptest.NewRecorder()
	handler.HandleSlackEvents(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestCustomTriggerEmoji(t *testing.T) {
	t.Parallel()

	spy := &spyIssueTrigger{}
	handler := newEventHandler(spy, "github")

	event := slackpkg.SlackEvent{
		Type: slackpkg.EventTypeEventCallback,
		Event: slackpkg.EventCallback{
			Type:     slackpkg.EventTypeReactionAdded,
			User:     "U123",
			Reaction: "github",
			Item: slackpkg.ReactionItem{
				Type:    "message",
				Channel: "C456",
				TS:      "1234567890.123456",
			},
		},
	}

	rr := postJSON(handler.HandleSlackEvents, event)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	if !spy.called {
		t.Error("expected issue trigger to be called for custom emoji")
	}
}
