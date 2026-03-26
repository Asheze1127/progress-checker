package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
	slackpkg "github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

// InteractionHandler handles Slack interactive component callbacks.
type InteractionHandler struct {
	resolveQuestion  *usecase.ResolveQuestionUsecase
	continueQuestion *usecase.ContinueQuestionUsecase
	escalateQuestion *usecase.EscalateQuestionUsecase
}

// NewInteractionHandler creates a new InteractionHandler.
func NewInteractionHandler(
	resolve *usecase.ResolveQuestionUsecase,
	cont *usecase.ContinueQuestionUsecase,
	escalate *usecase.EscalateQuestionUsecase,
) *InteractionHandler {
	return &InteractionHandler{
		resolveQuestion:  resolve,
		continueQuestion: cont,
		escalateQuestion: escalate,
	}
}

// slackInteractionPayload represents the top-level Slack interaction callback.
type slackInteractionPayload struct {
	Type    string            `json:"type"`
	Actions []slackAction     `json:"actions"`
	User    slackUser         `json:"user"`
}

// slackAction represents a single action within a Slack interaction payload.
type slackAction struct {
	ActionID string `json:"action_id"`
	Value    string `json:"value"`
}

// slackUser represents the user who triggered the action.
type slackUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// HandleInteraction is the HTTP handler for POST /webhook/slack/interactions.
// Slack sends interactive payloads as a URL-encoded form with the JSON in
// a field named "payload".
func (h *InteractionHandler) HandleInteraction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		slog.Error("error parsing form", slog.String("error", err.Error()))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	rawPayload := r.FormValue("payload")
	if rawPayload == "" {
		http.Error(w, "missing payload", http.StatusBadRequest)
		return
	}

	var payload slackInteractionPayload
	if err := json.Unmarshal([]byte(rawPayload), &payload); err != nil {
		slog.Error("error unmarshaling interaction payload", slog.String("error", err.Error()))
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if len(payload.Actions) == 0 {
		// Acknowledge with 200 even when there are no actions to process.
		w.WriteHeader(http.StatusOK)
		return
	}

	// Process only the first action (Slack sends one action per interaction).
	action := payload.Actions[0]
	questionID := entities.QuestionID(action.Value)

	if questionID == "" {
		slog.Warn("interaction has empty question ID", slog.String("user_id", payload.User.ID))
		http.Error(w, "missing question ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	var err error
	switch action.ActionID {
	case slackpkg.ActionIDQuestionResolved:
		err = h.resolveQuestion.Execute(ctx, questionID)
	case slackpkg.ActionIDQuestionContinue:
		err = h.continueQuestion.Execute(ctx, questionID)
	case slackpkg.ActionIDQuestionEscalate:
		err = h.escalateQuestion.Execute(ctx, questionID)
	default:
		slog.Warn("unknown action_id", slog.String("action_id", action.ActionID), slog.String("user_id", payload.User.ID))
		w.WriteHeader(http.StatusOK)
		return
	}

	if err != nil {
		slog.Error("error handling action", slog.String("action_id", action.ActionID), slog.String("question_id", string(questionID)), slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Slack expects a 200 OK to acknowledge the interaction.
	w.WriteHeader(http.StatusOK)
}
