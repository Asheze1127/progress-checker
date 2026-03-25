package rest

import (
	"log"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/application"
)

// QuestionHandler handles HTTP requests for the /question slash command webhook.
type QuestionHandler struct {
	service *application.QuestionService
}

// NewQuestionHandler creates a new QuestionHandler with the given service.
func NewQuestionHandler(service *application.QuestionService) *QuestionHandler {
	return &QuestionHandler{service: service}
}

const (
	slackCommandQuestion = "/question"
	contentTypeJSON      = "application/json"
)

// HandleWebhook processes incoming Slack slash command webhooks for /question.
// It parses the Slack payload and routes new questions to the SQS queue.
// Slack requires a response within 3 seconds.
func (h *QuestionHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	command := r.FormValue("command")
	if command != slackCommandQuestion {
		http.Error(w, "unsupported command", http.StatusBadRequest)
		return
	}

	threadTS := r.FormValue("thread_ts")

	// Only handle new questions (no thread_ts). Follow-up questions are handled separately.
	if threadTS != "" {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", contentTypeJSON)
		_, _ = w.Write([]byte(`{"text":"Follow-up questions are not yet supported."}`))
		return
	}

	text := r.FormValue("text")
	if text == "" {
		http.Error(w, "question text is required", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	channelID := r.FormValue("channel_id")
	if channelID == "" {
		http.Error(w, "channel_id is required", http.StatusBadRequest)
		return
	}

	input := application.NewQuestionInput{
		ParticipantID:  userID,
		Title:          truncateTitle(text),
		Text:           text,
		SlackChannelID: channelID,
	}

	if err := h.service.HandleNewQuestion(r.Context(), input); err != nil {
		log.Printf("ERROR: failed to handle new question: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"text":"Your question has been received and is being processed."}`))
}

const maxTitleLength = 100

// truncateTitle extracts a title from the question text, truncating if needed.
func truncateTitle(text string) string {
	if len(text) <= maxTitleLength {
		return text
	}
	return text[:maxTitleLength] + "..."
}
