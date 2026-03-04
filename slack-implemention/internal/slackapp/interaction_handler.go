package slackapp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/Asheze1127/progress-checker/slack-implemention/internal/service"
)

type InteractionHandler struct {
	signingSecret string
	service       *service.SlackService
}

func NewInteractionHandler(signingSecret string, svc *service.SlackService) *InteractionHandler {
	return &InteractionHandler{signingSecret: signingSecret, service: svc}
}

func (h *InteractionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := readVerifiedBody(r, h.signingSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	payload, err := parseInteractionPayload(body)
	if err != nil {
		log.Printf("failed to parse interaction payload: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch payload.Type {
	case "block_actions":
		h.handleBlockActions(w, payload)
	case "view_submission":
		h.handleViewSubmission(w, payload)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

func (h *InteractionHandler) handleBlockActions(w http.ResponseWriter, payload interactionPayload) {
	action, ok := findAction(payload.Actions, service.ProgressOpenModalActionID)
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}

	metadata := service.ProgressModalMetadata{}
	if strings.TrimSpace(action.Value) != "" {
		if err := json.Unmarshal([]byte(action.Value), &metadata); err != nil {
			log.Printf("failed to decode action metadata: %v", err)
		}
	}

	if metadata.ChannelID == "" {
		metadata.ChannelID = payload.Channel.ID
	}
	if metadata.ThreadTS == "" {
		metadata.ThreadTS = firstNonEmpty(payload.Message.ThreadTS, payload.Message.TS)
	}

	if err := h.service.OpenProgressModal(payload.TriggerID, metadata); err != nil {
		log.Printf("failed to open modal: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *InteractionHandler) handleViewSubmission(w http.ResponseWriter, payload interactionPayload) {
	if payload.View.CallbackID != service.ProgressModalCallbackID {
		writeJSON(w, http.StatusOK, map[string]string{"response_action": "clear"})
		return
	}

	metadata := service.ProgressModalMetadata{}
	if strings.TrimSpace(payload.View.PrivateMetadata) != "" {
		if err := json.Unmarshal([]byte(payload.View.PrivateMetadata), &metadata); err != nil {
			log.Printf("failed to decode private metadata: %v", err)
		}
	}

	phase := selectedOptionValue(payload.View.State.Values, service.ProgressPhaseBlockID, service.ProgressPhaseActionID)
	status := selectedOptionValue(payload.View.State.Values, service.ProgressStatusBlockID, service.ProgressStatusActionID)
	note := textInputValue(payload.View.State.Values, service.ProgressNoteBlockID, service.ProgressNoteActionID)

	validationErrors := map[string]string{}
	if phase == "" {
		validationErrors[service.ProgressPhaseBlockID] = "フェーズを選択してください。"
	}
	if status == "" {
		validationErrors[service.ProgressStatusBlockID] = "現状を選択してください。"
	}
	if metadata.ChannelID == "" || metadata.ThreadTS == "" {
		validationErrors[service.ProgressNoteBlockID] = "スレッド情報の取得に失敗しました。再度ボタンから入力してください。"
	}

	if len(validationErrors) > 0 {
		writeJSON(w, http.StatusOK, map[string]any{
			"response_action": "errors",
			"errors":          validationErrors,
		})
		return
	}

	if err := h.service.SubmitProgressReport(payload.User.ID, metadata, phase, status, note); err != nil {
		log.Printf("failed to submit progress report: %v", err)
		writeJSON(w, http.StatusOK, map[string]any{
			"response_action": "errors",
			"errors": map[string]string{
				service.ProgressNoteBlockID: fmt.Sprintf("送信に失敗しました: %v", err),
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"response_action": "clear"})
}

func parseInteractionPayload(body []byte) (interactionPayload, error) {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return interactionPayload{}, fmt.Errorf("parse form body: %w", err)
	}

	payloadRaw := values.Get("payload")
	if payloadRaw == "" {
		return interactionPayload{}, fmt.Errorf("payload is empty")
	}

	var payload interactionPayload
	if err := json.Unmarshal([]byte(payloadRaw), &payload); err != nil {
		return interactionPayload{}, fmt.Errorf("parse payload json: %w", err)
	}
	return payload, nil
}

func findAction(actions []interactionAction, actionID string) (interactionAction, bool) {
	for _, action := range actions {
		if action.ActionID == actionID {
			return action, true
		}
	}
	return interactionAction{}, false
}

func selectedOptionValue(values map[string]map[string]interactionStateValue, blockID, actionID string) string {
	if values == nil {
		return ""
	}
	blockValues, ok := values[blockID]
	if !ok {
		return ""
	}
	actionValue, ok := blockValues[actionID]
	if !ok {
		return ""
	}
	return strings.TrimSpace(actionValue.SelectedOption.Value)
}

func textInputValue(values map[string]map[string]interactionStateValue, blockID, actionID string) string {
	if values == nil {
		return ""
	}
	blockValues, ok := values[blockID]
	if !ok {
		return ""
	}
	actionValue, ok := blockValues[actionID]
	if !ok {
		return ""
	}
	return strings.TrimSpace(actionValue.Value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

type interactionPayload struct {
	Type      string              `json:"type"`
	TriggerID string              `json:"trigger_id"`
	User      interactionUser     `json:"user"`
	Channel   interactionChannel  `json:"channel"`
	Message   interactionMessage  `json:"message"`
	Actions   []interactionAction `json:"actions"`
	View      interactionView     `json:"view"`
}

type interactionUser struct {
	ID string `json:"id"`
}

type interactionChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type interactionMessage struct {
	TS       string `json:"ts"`
	ThreadTS string `json:"thread_ts"`
}

type interactionAction struct {
	ActionID string `json:"action_id"`
	Value    string `json:"value"`
}

type interactionView struct {
	CallbackID      string           `json:"callback_id"`
	PrivateMetadata string           `json:"private_metadata"`
	State           interactionState `json:"state"`
}

type interactionState struct {
	Values map[string]map[string]interactionStateValue `json:"values"`
}

type interactionStateValue struct {
	Value          string `json:"value"`
	SelectedOption struct {
		Value string `json:"value"`
	} `json:"selected_option"`
}
