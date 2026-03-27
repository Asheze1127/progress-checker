package webhook

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
)

var slackMentionRegex = regexp.MustCompile(`<@([A-Z0-9]+)(?:\|[^>]*)?>`)

// CommandHandler handles Slack slash commands.
type CommandHandler struct {
	createMentorUC *usecase.CreateMentorUseCase
}

// NewCommandHandler creates a new CommandHandler.
func NewCommandHandler(createMentorUC *usecase.CreateMentorUseCase) *CommandHandler {
	return &CommandHandler{createMentorUC: createMentorUC}
}

// HandleCommand processes incoming Slack slash commands.
// Slash commands are sent as application/x-www-form-urlencoded.
func (h *CommandHandler) HandleCommand(c *gin.Context) {
	command := c.PostForm("command")
	text := c.PostForm("text")
	callerID := c.PostForm("user_id")

	switch command {
	case "/create-mentor":
		h.handleCreateMentor(c, callerID, text)
	default:
		c.JSON(http.StatusOK, gin.H{"text": fmt.Sprintf("Unknown command: %s", command)})
	}
}

func (h *CommandHandler) handleCreateMentor(c *gin.Context, callerSlackID, text string) {
	mentorSlackID, teamName, err := parseCreateMentorArgs(text)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"response_type": "ephemeral",
			"text":          fmt.Sprintf("Usage: /create-mentor @user TeamName\nError: %s", err.Error()),
		})
		return
	}

	result, err := h.createMentorUC.Execute(c.Request.Context(), callerSlackID, mentorSlackID, teamName)
	if err != nil {
		slog.Error("failed to create mentor", slog.String("error", err.Error()), slog.String("caller", callerSlackID))
		c.JSON(http.StatusOK, gin.H{
			"response_type": "ephemeral",
			"text":          fmt.Sprintf("Failed to create mentor: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response_type": "ephemeral",
		"text": fmt.Sprintf(
			"Mentor created successfully!\nName: %s\nTeam: %s\nSetup URL: %s\n\nPlease share this URL with the mentor to set their password.",
			result.User.Name, teamName, result.SetupURL,
		),
	})
}

func parseCreateMentorArgs(text string) (slackUserID, teamName string, err error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", "", fmt.Errorf("arguments required: @user TeamName")
	}

	matches := slackMentionRegex.FindStringSubmatch(text)
	if len(matches) < 2 {
		return "", "", fmt.Errorf("please mention a user with @")
	}

	slackUserID = matches[1]

	remaining := strings.TrimSpace(slackMentionRegex.ReplaceAllString(text, ""))
	if remaining == "" {
		return "", "", fmt.Errorf("team name is required after the user mention")
	}

	return slackUserID, remaining, nil
}
