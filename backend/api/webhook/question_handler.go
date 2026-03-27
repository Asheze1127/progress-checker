package webhook

import (
	"log/slog"
	"net/http"

	"github.com/Asheze1127/progress-checker/backend/api/httputil"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
)

type QuestionHandler struct {
	handleNewQuestionUC *usecase.HandleNewQuestionUseCase
}

func NewQuestionHandler(handleNewQuestionUC *usecase.HandleNewQuestionUseCase) *QuestionHandler {
	return &QuestionHandler{handleNewQuestionUC: handleNewQuestionUC}
}

const (
	slackCommandQuestion = "/question"
	contentTypeJSON      = "application/json"
)

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
	if threadTS != "" {
		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(http.StatusOK)
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
	input := usecase.HandleNewQuestionInput{
		ParticipantID: userID, Title: truncateTitle(text), Text: text, SlackChannelID: channelID,
	}
	if err := h.handleNewQuestionUC.Execute(r.Context(), input); err != nil {
		slog.Error("failed to handle new question", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"text": "Your question has been received and is being processed."})
}

const maxTitleLength = 100

func truncateTitle(text string) string {
	if len(text) <= maxTitleLength {
		return text
	}
	return text[:maxTitleLength] + "..."
}
