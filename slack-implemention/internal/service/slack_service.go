package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/slack-go/slack"
)

type SlackService struct {
	client *slack.Client
}

func NewSlackService(botToken string) *SlackService {
	return &SlackService{client: slack.New(botToken)}
}

func (s *SlackService) StartProgressFlow(userID, channelID, channelName string) (SlashResponse, error) {
	rootText := ":bar_chart: 進捗報告スレッド"
	_, threadTS, err := s.client.PostMessage(channelID, slack.MsgOptionText(rootText, false))
	if err != nil {
		return SlashResponse{}, fmt.Errorf("failed to create progress thread root message: %w", err)
	}

	metadata := ProgressModalMetadata{ChannelID: channelID, ThreadTS: threadTS}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return SlashResponse{}, fmt.Errorf("failed to encode progress metadata: %w", err)
	}

	if err := s.postProgressTemplateMessage(channelID, threadTS, string(metadataJSON)); err != nil {
		return SlashResponse{}, err
	}

	log.Printf("progress thread created channel_id=%s channel_name=%s thread_ts=%s user_id=%s", channelID, channelName, threadTS, userID)

	// Intentionally return an empty slash response to avoid extra channel output.
	return SlashResponse{}, nil
}

func (s *SlackService) StartQuestionFlow(userID, channelID, channelName, questionText string) (SlashResponse, error) {
	question := strings.TrimSpace(questionText)
	if question == "" {
		return SlashResponse{
			ResponseType: "ephemeral",
			Text:         "`/question 質問内容` の形式で入力してください。",
		}, nil
	}

	msg := fmt.Sprintf(":raising_hand: 質問受付\n<@%s>\n質問内容: %s", userID, question)
	_, threadTS, err := s.client.PostMessage(channelID, slack.MsgOptionText(msg, false))
	if err != nil {
		return SlashResponse{}, fmt.Errorf("failed to post question confirmation: %w", err)
	}

	if _, _, err := s.client.PostMessage(
		channelID,
		slack.MsgOptionText("質問内容を確認します。", false),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{ThreadTimestamp: threadTS}),
	); err != nil {
		return SlashResponse{}, fmt.Errorf("failed to post question thread confirmation: %w", err)
	}

	log.Printf("question received channel_id=%s channel_name=%s thread_ts=%s user_id=%s question=%q", channelID, channelName, threadTS, userID, question)
	log.Printf("thread reply command: go run ./cmd/thread-reply --channel %s --thread %s --text \"<返信内容>\"", channelID, threadTS)

	// Intentionally return an empty slash response to avoid extra channel output.
	return SlashResponse{}, nil
}

func (s *SlackService) OpenProgressModal(triggerID string, metadata ProgressModalMetadata) error {
	if strings.TrimSpace(triggerID) == "" {
		return fmt.Errorf("trigger id is required")
	}
	if strings.TrimSpace(metadata.ChannelID) == "" || strings.TrimSpace(metadata.ThreadTS) == "" {
		return fmt.Errorf("channel_id and thread_ts are required")
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to encode modal metadata: %w", err)
	}

	phaseOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject("企画", slack.NewTextBlockObject(slack.PlainTextType, "企画", false, false), nil),
		slack.NewOptionBlockObject("要件定義", slack.NewTextBlockObject(slack.PlainTextType, "要件定義", false, false), nil),
		slack.NewOptionBlockObject("実装", slack.NewTextBlockObject(slack.PlainTextType, "実装", false, false), nil),
		slack.NewOptionBlockObject("発表準備", slack.NewTextBlockObject(slack.PlainTextType, "発表準備", false, false), nil),
	}
	phaseSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, slack.NewTextBlockObject(slack.PlainTextType, "フェーズを選択", false, false), ProgressPhaseActionID, phaseOptions...)
	phaseBlock := slack.NewInputBlock(ProgressPhaseBlockID, slack.NewTextBlockObject(slack.PlainTextType, "フェーズ", false, false), nil, phaseSelect)

	statusOptions := []*slack.OptionBlockObject{
		slack.NewOptionBlockObject("詰まっている", slack.NewTextBlockObject(slack.PlainTextType, "詰まっている", false, false), nil),
		slack.NewOptionBlockObject("普通", slack.NewTextBlockObject(slack.PlainTextType, "普通", false, false), nil),
		slack.NewOptionBlockObject("いい感じ", slack.NewTextBlockObject(slack.PlainTextType, "いい感じ", false, false), nil),
	}
	statusSelect := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, slack.NewTextBlockObject(slack.PlainTextType, "現状を選択", false, false), ProgressStatusActionID, statusOptions...)
	statusBlock := slack.NewInputBlock(ProgressStatusBlockID, slack.NewTextBlockObject(slack.PlainTextType, "現状", false, false), nil, statusSelect)

	noteInput := slack.NewPlainTextInputBlockElement(slack.NewTextBlockObject(slack.PlainTextType, "好きなことを書いて", false, false), ProgressNoteActionID)
	noteInput.Multiline = true
	noteInputBlock := slack.NewInputBlock(ProgressNoteBlockID, slack.NewTextBlockObject(slack.PlainTextType, "備考", false, false), nil, noteInput)
	noteInputBlock.Optional = true

	view := slack.ModalViewRequest{
		Type:            slack.VTModal,
		CallbackID:      ProgressModalCallbackID,
		PrivateMetadata: string(metadataJSON),
		Title:           slack.NewTextBlockObject(slack.PlainTextType, "進捗報告", false, false),
		Submit:          slack.NewTextBlockObject(slack.PlainTextType, "送信", false, false),
		Close:           slack.NewTextBlockObject(slack.PlainTextType, "閉じる", false, false),
		Blocks: slack.Blocks{BlockSet: []slack.Block{
			phaseBlock,
			statusBlock,
			noteInputBlock,
		}},
	}

	if _, err := s.client.OpenView(triggerID, view); err != nil {
		return fmt.Errorf("failed to open progress modal: %w", err)
	}

	return nil
}

func (s *SlackService) SubmitProgressReport(userID string, metadata ProgressModalMetadata, phase, status, note string) error {
	if strings.TrimSpace(metadata.ChannelID) == "" || strings.TrimSpace(metadata.ThreadTS) == "" {
		return fmt.Errorf("channel_id and thread_ts are required")
	}
	if !isAllowedProgressPhase(phase) {
		return fmt.Errorf("invalid phase: %s", phase)
	}
	if !isAllowedProgressStatus(status) {
		return fmt.Errorf("invalid status: %s", status)
	}

	message := strings.Join([]string{
		":white_check_mark: *進捗報告*",
		fmt.Sprintf("*報告者:* <@%s>", userID),
		fmt.Sprintf("*フェーズ:* %s", phase),
		fmt.Sprintf("*現状:* %s", status),
		fmt.Sprintf("*備考:* %s", emptyAsDash(strings.TrimSpace(note))),
	}, "\n")

	if _, _, err := s.client.PostMessage(
		metadata.ChannelID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{ThreadTimestamp: metadata.ThreadTS}),
	); err != nil {
		return fmt.Errorf("failed to post progress report: %w", err)
	}

	log.Printf("progress report posted channel_id=%s thread_ts=%s user_id=%s phase=%s status=%s", metadata.ChannelID, metadata.ThreadTS, userID, phase, status)
	return nil
}

func (s *SlackService) PostThreadReply(channelID, threadTS, text string) error {
	if strings.TrimSpace(channelID) == "" {
		return fmt.Errorf("channel is required")
	}
	if strings.TrimSpace(threadTS) == "" {
		return fmt.Errorf("thread is required")
	}
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("text is required")
	}

	if _, _, err := s.client.PostMessage(
		channelID,
		slack.MsgOptionText(text, false),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{ThreadTimestamp: threadTS}),
	); err != nil {
		return fmt.Errorf("failed to post thread reply: %w", err)
	}

	log.Printf("thread reply posted channel_id=%s thread_ts=%s", channelID, threadTS)
	return nil
}

func (s *SlackService) postProgressTemplateMessage(channelID, threadTS, metadataJSON string) error {
	formatText := strings.Join([]string{
		"*進捗フォーマット*",
		"- フェーズ: 企画 / 要件定義 / 実装 / 発表準備",
		"- 現状: 詰まっている / 普通 / いい感じ",
		"- 備考: 好きなことを書いてください",
	}, "\n")

	button := slack.NewButtonBlockElement(
		ProgressOpenModalActionID,
		metadataJSON,
		slack.NewTextBlockObject(slack.PlainTextType, "進捗を入力する", false, false),
	)
	button.Style = slack.StylePrimary

	sectionBlock := slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, formatText, false, false), nil, nil)
	actionBlock := slack.NewActionBlock("progress_form_action_block", button)

	if _, _, err := s.client.PostMessage(
		channelID,
		slack.MsgOptionText("進捗フォーマット", false),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{ThreadTimestamp: threadTS}),
		slack.MsgOptionBlocks(sectionBlock, actionBlock),
	); err != nil {
		return fmt.Errorf("failed to post progress template in thread: %w", err)
	}

	return nil
}

func isAllowedProgressPhase(phase string) bool {
	switch phase {
	case "企画", "要件定義", "実装", "発表準備":
		return true
	default:
		return false
	}
}

func isAllowedProgressStatus(status string) bool {
	switch status {
	case "詰まっている", "普通", "いい感じ":
		return true
	default:
		return false
	}
}

func emptyAsDash(input string) string {
	if input == "" {
		return "-"
	}
	return input
}
