package rest

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/application"
)

const progressCommand = "/progress"

// ProgressCommandHandler defines the interface for handling progress commands.
type ProgressCommandHandler interface {
	HandleProgressCommand(ctx context.Context, input application.ProgressCommandInput) error
}

// WebhookHandler handles incoming Slack webhook requests.
type WebhookHandler struct {
	progressHandler ProgressCommandHandler
}

// NewWebhookHandler creates a new WebhookHandler with the given dependencies.
func NewWebhookHandler(progressHandler ProgressCommandHandler) *WebhookHandler {
	return &WebhookHandler{
		progressHandler: progressHandler,
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
		h.handleProgress(r.Context(), w, userID, teamID, channelID, text)
	default:
		http.Error(w, "unsupported command: "+command, http.StatusBadRequest)
	}
}

// handleProgress parses the progress text and delegates to the application service.
// Expected text format: "phase:<phase> [sos:true] [comment:<comment>]"
func (h *WebhookHandler) handleProgress(ctx context.Context, w http.ResponseWriter, userID, teamID, channelID, text string) {
	phase, sos, comment := parseProgressText(text)

	input := application.ProgressCommandInput{
		SlackUserID: userID,
		TeamID:      teamID,
		ChannelID:   channelID,
		Phase:       phase,
		SOS:         sos,
		Comment:     comment,
	}

	if err := h.progressHandler.HandleProgressCommand(ctx, input); err != nil {
		log.Printf("ERROR: failed to handle progress command: %v", err)
		http.Error(w, "failed to process progress command", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"response_type":"in_channel","text":"進捗報告を受け付けました"}`))
}

// parseProgressText parses the text field from the /progress slash command.
// Format: "phase:<phase> [sos:true] [comment:<comment>]"
// Example: "phase:coding sos:true comment:Working on feature X"
func parseProgressText(text string) (phase string, sos bool, comment string) {
	parts := strings.Fields(text)

	var commentParts []string
	collectingComment := false

	for _, part := range parts {
		if collectingComment {
			commentParts = append(commentParts, part)
			continue
		}

		if strings.HasPrefix(part, "phase:") {
			phase = strings.TrimPrefix(part, "phase:")
		} else if strings.HasPrefix(part, "sos:") {
			sos = strings.TrimPrefix(part, "sos:") == "true"
		} else if strings.HasPrefix(part, "comment:") {
			collectingComment = true
			firstWord := strings.TrimPrefix(part, "comment:")
			if firstWord != "" {
				commentParts = append(commentParts, firstWord)
			}
		}
	}

	comment = strings.Join(commentParts, " ")
	return phase, sos, comment
}
