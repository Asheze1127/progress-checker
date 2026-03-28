package webhook

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

const progressCommand = "/progress"

type WebhookHandler struct {
	progressUseCase *usecase.HandleProgressUseCase
}

func NewWebhookHandler(progressUseCase *usecase.HandleProgressUseCase) *WebhookHandler {
	return &WebhookHandler{progressUseCase: progressUseCase}
}

func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	command := c.PostForm("command")
	userID := c.PostForm("user_id")
	teamID := c.PostForm("team_id")
	channelID := c.PostForm("channel_id")

	switch command {
	case progressCommand:
		h.handleProgress(c, userID, teamID, channelID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported command"})
	}
}

func (h *WebhookHandler) handleProgress(c *gin.Context, userID, teamID, channelID string) {
	phase := entities.ProgressPhase(c.PostForm("phase"))
	if !phase.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid phase value"})
		return
	}
	sos := c.PostForm("sos") == "true"
	comment := c.PostForm("comment")

	input := usecase.HandleProgressInput{
		SlackUserID: userID, TeamID: teamID, ChannelID: channelID,
		Phase: phase, SOS: sos, Comment: comment,
	}

	if err := h.progressUseCase.Execute(c.Request.Context(), input); err != nil {
		slog.Error("failed to handle progress command", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process progress command"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response_type": "in_channel", "text": "進捗報告を受け付けました"})
}
