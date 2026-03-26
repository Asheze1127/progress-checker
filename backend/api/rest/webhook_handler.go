package rest

import (
	"log/slog"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

const progressCommand = "/progress"

// WebhookHandler handles incoming Slack webhook requests.
type WebhookHandler struct {
	progressUseCase *usecase.HandleProgressUseCase
}

// NewWebhookHandler creates a new WebhookHandler with the given dependencies.
func NewWebhookHandler(progressUseCase *usecase.HandleProgressUseCase) *WebhookHandler {
	return &WebhookHandler{
		progressUseCase: progressUseCase,
	}
}

// HandleWebhook processes incoming POST /webhook/slack requests.
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	command := r.FormValue("command")
	userID := r.FormValue("user_id")
	teamID := r.FormValue("team_id")
	channelID := r.FormValue("channel_id")

	switch command {
	case progressCommand:
		h.handleProgress(w, r, userID, teamID, channelID)
	default:
		http.Error(w, "unsupported command", http.StatusBadRequest)
	}
}

// handleProgress extracts structured fields from the Slack modal submission
// and passes them to the use case. Slack sends phase, sos, and comment as
// separate form values from the modal UI.
func (h *WebhookHandler) handleProgress(w http.ResponseWriter, r *http.Request, userID, teamID, channelID string) {
	phase := r.FormValue("phase")
	sos := r.FormValue("sos") == "true"
	comment := r.FormValue("comment")

	input := usecase.HandleProgressInput{
		SlackUserID: userID,
		TeamID:      teamID,
		ChannelID:   channelID,
		Phase:       entities.ProgressPhase(phase),
		SOS:         sos,
		Comment:     comment,
	}

	if err := h.progressUseCase.Execute(r.Context(), input); err != nil {
		slog.Error("failed to handle progress command", slog.String("error", err.Error()))
		http.Error(w, "failed to process progress command", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"response_type":"in_channel","text":"進捗報告を受け付けました"}`))
}
