package slackapp

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/Asheze1127/progress-checker/slack-implemention/internal/service"
	"github.com/slack-go/slack"
)

type CommandHandler struct {
	signingSecret string
	service       *service.SlackService
}

func NewCommandHandler(signingSecret string, svc *service.SlackService) *CommandHandler {
	return &CommandHandler{signingSecret: signingSecret, service: svc}
}

func (h *CommandHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, service.SlashResponse{Text: "method not allowed"})
		return
	}

	body, err := readVerifiedBody(r, h.signingSecret)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, service.SlashResponse{Text: "invalid request signature"})
		return
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))
	cmd, err := slack.SlashCommandParse(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, service.SlashResponse{Text: "failed to parse slash command"})
		return
	}

	var resp service.SlashResponse
	switch cmd.Command {
	case "/progress":
		resp, err = h.service.StartProgressFlow(cmd.UserID, cmd.ChannelID, cmd.ChannelName)
	case "/question":
		resp, err = h.service.StartQuestionFlow(cmd.UserID, cmd.ChannelID, cmd.ChannelName, cmd.Text)
	default:
		resp = service.SlashResponse{
			ResponseType: "ephemeral",
			Text:         "対応していないコマンドです。利用可能: /progress, /question",
		}
	}

	if err != nil {
		log.Printf("slash command handling error: %v", err)
		writeJSON(w, http.StatusInternalServerError, service.SlashResponse{
			ResponseType: "ephemeral",
			Text:         "コマンド処理に失敗しました。時間をおいて再試行してください。",
		})
		return
	}

	if resp.ResponseType == "" && resp.Text == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
