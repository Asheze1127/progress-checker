package webhook

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
)

type QuestionHandler struct {
	handleNewQuestionUC *usecase.HandleNewQuestionUseCase
}

func NewQuestionHandler(handleNewQuestionUC *usecase.HandleNewQuestionUseCase) *QuestionHandler {
	return &QuestionHandler{handleNewQuestionUC: handleNewQuestionUC}
}

const slackCommandQuestion = "/question"

func (h *QuestionHandler) HandleWebhook(c *gin.Context) {
	command := c.PostForm("command")
	if command != slackCommandQuestion {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported command"})
		return
	}

	threadTS := c.PostForm("thread_ts")
	if threadTS != "" {
		c.JSON(http.StatusOK, gin.H{"text": "Follow-up questions are not yet supported."})
		return
	}

	text := c.PostForm("text")
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question text is required"})
		return
	}

	userID := c.PostForm("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	channelID := c.PostForm("channel_id")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel_id is required"})
		return
	}

	input := usecase.HandleNewQuestionInput{
		ParticipantID: userID, Title: truncateTitle(text), Text: text, SlackChannelID: channelID,
	}

	if err := h.handleNewQuestionUC.Execute(c.Request.Context(), input); err != nil {
		slog.Error("failed to handle new question", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"text": "Your question has been received and is being processed."})
}

const maxTitleLength = 100

func truncateTitle(text string) string {
	runes := []rune(text)
	if len(runes) <= maxTitleLength {
		return text
	}
	return string(runes[:maxTitleLength]) + "..."
}
