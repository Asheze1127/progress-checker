package webhook

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
	slackpkg "github.com/Asheze1127/progress-checker/backend/pkg/slack"
)

type InteractionHandler struct {
	resolveQuestion  *usecase.ResolveQuestionUseCase
	continueQuestion *usecase.ContinueQuestionUseCase
	escalateQuestion *usecase.EscalateQuestionUseCase
}

func NewInteractionHandler(resolve *usecase.ResolveQuestionUseCase, cont *usecase.ContinueQuestionUseCase, escalate *usecase.EscalateQuestionUseCase) *InteractionHandler {
	return &InteractionHandler{resolveQuestion: resolve, continueQuestion: cont, escalateQuestion: escalate}
}

// HandleInteraction is the Gin handler for POST /webhook/slack/interactions.
// Slack sends interactive payloads as a URL-encoded form with the JSON in
// a field named "payload".
func (h *InteractionHandler) HandleInteraction(c *gin.Context) {
	rawPayload := c.PostForm("payload")
	if rawPayload == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing payload"})
		return
	}

	var callback slack.InteractionCallback
	if err := json.Unmarshal([]byte(rawPayload), &callback); err != nil {
		slog.Error("error unmarshaling interaction payload", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if len(callback.ActionCallback.BlockActions) == 0 {
		c.Status(http.StatusOK)
		return
	}

	action := callback.ActionCallback.BlockActions[0]
	questionID := entities.QuestionID(action.Value)
	if questionID == "" {
		slog.Warn("interaction has empty question ID", slog.String("user_id", callback.User.ID))
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing question ID"})
		return
	}

	ctx := c.Request.Context()
	var err error
	switch action.ActionID {
	case slackpkg.ActionIDQuestionResolved:
		err = h.resolveQuestion.Execute(ctx, questionID)
	case slackpkg.ActionIDQuestionContinue:
		err = h.continueQuestion.Execute(ctx, questionID)
	case slackpkg.ActionIDQuestionEscalate:
		err = h.escalateQuestion.Execute(ctx, questionID)
	default:
		slog.Warn("unknown action_id", slog.String("action_id", action.ActionID), slog.String("user_id", callback.User.ID))
		c.Status(http.StatusOK)
		return
	}

	if err != nil {
		slog.Error("error handling action", slog.String("action_id", action.ActionID), slog.String("question_id", string(questionID)), slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusOK)
}
