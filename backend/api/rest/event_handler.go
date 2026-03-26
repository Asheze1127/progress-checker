package rest

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	slackpkg "github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

// IssueTrigger defines the interface for triggering issue creation.
type IssueTrigger interface {
	Execute(ctx context.Context, input usecase.TriggerIssueCreationInput) error
}

// EventHandler handles incoming Slack Events API webhooks.
type EventHandler struct {
	triggerEmoji string
	issueTrigger IssueTrigger
}

// NewEventHandler creates a new EventHandler with the configured trigger emoji.
func NewEventHandler(issueTrigger IssueTrigger, triggerEmoji string) *EventHandler {
	return &EventHandler{
		triggerEmoji: triggerEmoji,
		issueTrigger: issueTrigger,
	}
}

// HandleSlackEvents is the HTTP handler for POST /webhook/slack/events.
func (h *EventHandler) HandleSlackEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event slackpkg.SlackEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	switch event.Type {
	case slackpkg.EventTypeURLVerification:
		h.handleURLVerification(w, event)
	case slackpkg.EventTypeEventCallback:
		h.handleEventCallback(w, r, event)
	case slackpkg.EventTypeMessageAction:
		h.handleMessageAction(w, r, event)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

func (h *EventHandler) handleURLVerification(w http.ResponseWriter, event slackpkg.SlackEvent) {
	w.Header().Set("Content-Type", "application/json")
	resp := slackpkg.URLVerificationResponse{Challenge: event.Challenge}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode url_verification response", slog.String("error", err.Error()))
	}
}

func (h *EventHandler) handleEventCallback(w http.ResponseWriter, r *http.Request, event slackpkg.SlackEvent) {
	switch event.Event.Type {
	case slackpkg.EventTypeReactionAdded:
		h.handleReactionAdded(w, r, event)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

func (h *EventHandler) handleReactionAdded(w http.ResponseWriter, r *http.Request, event slackpkg.SlackEvent) {
	if event.Event.Reaction != h.triggerEmoji {
		w.WriteHeader(http.StatusOK)
		return
	}

	input := usecase.TriggerIssueCreationInput{
		ChannelID:     event.Event.Item.Channel,
		ThreadTS:      event.Event.Item.TS,
		TriggerUserID: event.Event.User,
		TriggerType:   "reaction",
	}

	if err := h.issueTrigger.Execute(r.Context(), input); err != nil {
		slog.Error("failed to trigger issue creation from reaction", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *EventHandler) handleMessageAction(w http.ResponseWriter, r *http.Request, event slackpkg.SlackEvent) {
	threadTS := event.Message.ThreadTS
	if threadTS == "" {
		threadTS = event.Message.TS
	}

	input := usecase.TriggerIssueCreationInput{
		ChannelID:     event.Channel.ID,
		ThreadTS:      threadTS,
		TriggerUserID: event.User.ID,
		TriggerType:   "message_action",
	}

	if err := h.issueTrigger.Execute(r.Context(), input); err != nil {
		slog.Error("failed to trigger issue creation from message action", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
