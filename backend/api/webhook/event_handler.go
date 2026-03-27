package webhook

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack/slackevents"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	slackpkg "github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

type IssueTrigger interface {
	Execute(ctx context.Context, input usecase.TriggerIssueCreationInput) error
}

type EventHandler struct {
	triggerEmoji string
	issueTrigger IssueTrigger
}

func NewEventHandler(issueTrigger IssueTrigger, triggerEmoji string) *EventHandler {
	return &EventHandler{triggerEmoji: triggerEmoji, issueTrigger: issueTrigger}
}

func (h *EventHandler) HandleSlackEvents(c *gin.Context) {
	const maxBodySize = 1 << 20 // 1 MB
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var typeCheck struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &typeCheck); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if typeCheck.Type == slackpkg.EventTypeMessageAction {
		h.handleMessageAction(c, body)
		return
	}

	event, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	switch event.Type {
	case slackevents.URLVerification:
		h.handleURLVerification(c, body)
	case slackevents.CallbackEvent:
		h.handleEventCallback(c, event)
	default:
		c.Status(http.StatusOK)
	}
}

func (h *EventHandler) handleURLVerification(c *gin.Context, body []byte) {
	var challenge slackevents.ChallengeResponse
	if err := json.Unmarshal(body, &challenge); err != nil {
		slog.Error("failed to parse url_verification challenge", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid challenge"})
		return
	}
	c.JSON(http.StatusOK, challenge)
}

func (h *EventHandler) handleEventCallback(c *gin.Context, event slackevents.EventsAPIEvent) {
	innerEvent := event.InnerEvent
	switch innerEvent.Type {
	case string(slackevents.ReactionAdded):
		h.handleReactionAdded(c, innerEvent)
	default:
		c.Status(http.StatusOK)
	}
}

func (h *EventHandler) handleReactionAdded(c *gin.Context, innerEvent slackevents.EventsAPIInnerEvent) {
	ev, ok := innerEvent.Data.(*slackevents.ReactionAddedEvent)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unexpected event data"})
		return
	}

	if ev.Reaction != h.triggerEmoji {
		c.Status(http.StatusOK)
		return
	}

	input := usecase.TriggerIssueCreationInput{
		ChannelID:     ev.Item.Channel,
		ThreadTS:      ev.Item.Timestamp,
		TriggerUserID: ev.User,
		TriggerType:   "reaction",
	}

	if err := h.issueTrigger.Execute(c.Request.Context(), input); err != nil {
		slog.Error("failed to trigger issue creation from reaction", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *EventHandler) handleMessageAction(c *gin.Context, body []byte) {
	var payload slackpkg.MessageActionPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message action payload"})
		return
	}

	threadTS := payload.Message.ThreadTS
	if threadTS == "" {
		threadTS = payload.Message.TS
	}

	input := usecase.TriggerIssueCreationInput{
		ChannelID:     payload.Channel.ID,
		ThreadTS:      threadTS,
		TriggerUserID: payload.User.ID,
		TriggerType:   "message_action",
	}

	if err := h.issueTrigger.Execute(c.Request.Context(), input); err != nil {
		slog.Error("failed to trigger issue creation from message action", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusOK)
}
