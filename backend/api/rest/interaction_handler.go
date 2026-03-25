package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/application"
	"github.com/Asheze1127/progress-checker/backend/entities"
	slackpkg "github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

// InteractionHandler handles Slack interactive component callbacks.
type InteractionHandler struct {
	questionActionService *application.QuestionActionService
}

// NewInteractionHandler creates a new InteractionHandler.
func NewInteractionHandler(svc *application.QuestionActionService) *InteractionHandler {
	return &InteractionHandler{questionActionService: svc}
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
		log.Printf("error parsing form: %v", err)
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
		log.Printf("error unmarshaling interaction payload: %v", err)
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
		log.Printf("interaction from user %s has empty question ID", payload.User.ID)
		http.Error(w, "missing question ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	var err error
	switch action.ActionID {
	case slackpkg.ActionIDQuestionResolved:
		err = h.questionActionService.ResolveQuestion(ctx, questionID)
	case slackpkg.ActionIDQuestionContinue:
		err = h.questionActionService.ContinueQuestion(ctx, questionID)
	case slackpkg.ActionIDQuestionEscalate:
		err = h.questionActionService.EscalateToMentor(ctx, questionID)
	default:
		log.Printf("unknown action_id %q from user %s", action.ActionID, payload.User.ID)
		w.WriteHeader(http.StatusOK)
		return
	}

	if err != nil {
		log.Printf("error handling action %q for question %q: %v", action.ActionID, questionID, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Slack expects a 200 OK to acknowledge the interaction.
	w.WriteHeader(http.StatusOK)
}
