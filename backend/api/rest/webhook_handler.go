package rest

import (
	"log"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
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
	text := r.FormValue("text")

	switch command {
	case progressCommand:
		h.handleProgress(w, r, userID, teamID, channelID, text)
	default:
		http.Error(w, "unsupported command: "+command, http.StatusBadRequest)
	}
}

// handleProgress passes the raw Slack text directly to the use case.
// Slack sends the user's input as-is in the text field, so no custom parsing is needed.
func (h *WebhookHandler) handleProgress(w http.ResponseWriter, r *http.Request, userID, teamID, channelID, text string) {
	input := usecase.HandleProgressInput{
		SlackUserID: userID,
		TeamID:      teamID,
		ChannelID:   channelID,
		Text:        text,
	}

	if err := h.progressUseCase.Execute(r.Context(), input); err != nil {
		log.Printf("ERROR: failed to handle progress command: %v", err)
		http.Error(w, "failed to process progress command", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"response_type":"in_channel","text":"進捗報告を受け付けました"}`))
}
