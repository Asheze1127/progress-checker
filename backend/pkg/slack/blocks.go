package slack

import "encoding/json"

// Action IDs for interactive buttons shown after AI responses.
const (
	ActionIDQuestionResolved = "question_resolved"
	ActionIDQuestionContinue = "question_continue"
	ActionIDQuestionEscalate = "question_escalate"
)

// Button styles defined by Slack Block Kit.
const (
	ButtonStylePrimary = "primary"
	ButtonStyleDanger  = "danger"
)

// Block represents a single Slack Block Kit block.
type Block struct {
	Type     string    `json:"type"`
	Text     *TextObj  `json:"text,omitempty"`
	Elements []Element `json:"elements,omitempty"`
}

// TextObj represents a Slack Block Kit text object.
type TextObj struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Element represents a Slack Block Kit element such as a button.
type Element struct {
	Type    string  `json:"type"`
	Text    TextObj `json:"text"`
	ActionID string `json:"action_id"`
	Style   string  `json:"style,omitempty"`
	Value   string  `json:"value,omitempty"`
}

// BuildQuestionResponseBlocks returns Slack Block Kit blocks containing the AI
// response text and three action buttons: resolve, continue, and escalate.
// The questionID is embedded in each button's value so callbacks can identify
// the target question.
func BuildQuestionResponseBlocks(aiResponse string, questionID string) []Block {
	return []Block{
		{
			Type: "section",
			Text: &TextObj{
				Type: "mrkdwn",
				Text: aiResponse,
			},
		},
		{
			Type: "actions",
			Elements: []Element{
				{
					Type:     "button",
					Text:     TextObj{Type: "plain_text", Text: "\u2705 \u89e3\u6c7a\u3057\u305f"},
					ActionID: ActionIDQuestionResolved,
					Style:    ButtonStylePrimary,
					Value:    questionID,
				},
				{
					Type:     "button",
					Text:     TextObj{Type: "plain_text", Text: "\U0001F4AC \u7d9a\u3051\u3066\u8cea\u554f"},
					ActionID: ActionIDQuestionContinue,
					Value:    questionID,
				},
				{
					Type:     "button",
					Text:     TextObj{Type: "plain_text", Text: "\U0001F64B \u30e1\u30f3\u30bf\u30fc\u306b\u76f8\u8ac7"},
					ActionID: ActionIDQuestionEscalate,
					Style:    ButtonStyleDanger,
					Value:    questionID,
				},
			},
		},
	}
}

// MarshalBlocks serializes blocks to JSON bytes.
func MarshalBlocks(blocks []Block) ([]byte, error) {
	return json.Marshal(blocks)
}
