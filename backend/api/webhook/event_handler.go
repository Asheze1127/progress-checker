package webhook

import (
  "context"
  "encoding/json"
  "io"
  "log/slog"
  "net/http"

  "github.com/Asheze1127/progress-checker/backend/application/usecase"
  slackpkg "github.com/Asheze1127/progress-checker/backend/pkg/slack"
  "github.com/slack-go/slack/slackevents"
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

func (h *EventHandler) HandleSlackEvents(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodPost {
    http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    return
  }

  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "failed to read body", http.StatusBadRequest)
    return
  }

  // Check raw type field first for message_action, which is not an Events API
  // event and cannot be parsed by slackevents.ParseEvent.
  var typeCheck struct {
    Type string `json:"type"`
  }
  if err := json.Unmarshal(body, &typeCheck); err != nil {
    http.Error(w, "invalid request body", http.StatusBadRequest)
    return
  }

  if typeCheck.Type == slackpkg.EventTypeMessageAction {
    h.handleMessageAction(w, r, body)
    return
  }

  event, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
  if err != nil {
    http.Error(w, "invalid request body", http.StatusBadRequest)
    return
  }
  switch event.Type {
  case slackevents.URLVerification:
    h.handleURLVerification(w, body)
  case slackevents.CallbackEvent:
    h.handleEventCallback(w, r, event)
  default:
    w.WriteHeader(http.StatusOK)
  }
}

func (h *EventHandler) handleURLVerification(w http.ResponseWriter, body []byte) {
  var challenge slackevents.ChallengeResponse
  if err := json.Unmarshal(body, &challenge); err != nil {
    slog.Error("failed to parse url_verification challenge", slog.String("error", err.Error()))
    http.Error(w, "invalid challenge", http.StatusBadRequest)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  if err := json.NewEncoder(w).Encode(challenge); err != nil {
    slog.Error("failed to encode url_verification response", slog.String("error", err.Error()))
  }
}

func (h *EventHandler) handleEventCallback(w http.ResponseWriter, r *http.Request, event slackevents.EventsAPIEvent) {
  innerEvent := event.InnerEvent
  switch innerEvent.Type {
  case string(slackevents.ReactionAdded):
    h.handleReactionAdded(w, r, innerEvent)
  default:
    w.WriteHeader(http.StatusOK)
  }
}

func (h *EventHandler) handleReactionAdded(w http.ResponseWriter, r *http.Request, innerEvent slackevents.EventsAPIInnerEvent) {
  ev, ok := innerEvent.Data.(*slackevents.ReactionAddedEvent)
  if !ok {
    http.Error(w, "unexpected event data", http.StatusBadRequest)
    return
  }

  if ev.Reaction != h.triggerEmoji {
    w.WriteHeader(http.StatusOK)
    return
  }
  input := usecase.TriggerIssueCreationInput{
    ChannelID:     ev.Item.Channel,
    ThreadTS:      ev.Item.Timestamp,
    TriggerUserID: ev.User,
    TriggerType:   "reaction",
  }
  if err := h.issueTrigger.Execute(r.Context(), input); err != nil {
    slog.Error("failed to trigger issue creation from reaction", slog.String("error", err.Error()))
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
  }
  w.WriteHeader(http.StatusOK)
}

func (h *EventHandler) handleMessageAction(w http.ResponseWriter, r *http.Request, body []byte) {
  var payload slackpkg.MessageActionPayload
  if err := json.Unmarshal(body, &payload); err != nil {
    http.Error(w, "invalid message action payload", http.StatusBadRequest)
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
  if err := h.issueTrigger.Execute(r.Context(), input); err != nil {
    slog.Error("failed to trigger issue creation from message action", slog.String("error", err.Error()))
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
  }
  w.WriteHeader(http.StatusOK)
}
