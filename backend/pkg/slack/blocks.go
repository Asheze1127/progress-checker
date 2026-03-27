package slack

import "github.com/slack-go/slack"

// Action IDs for interactive buttons shown after AI responses.
const (
  ActionIDQuestionResolved = "question_resolved"
  ActionIDQuestionContinue = "question_continue"
  ActionIDQuestionEscalate = "question_escalate"
)

// BuildQuestionResponseBlocks returns Slack Block Kit blocks containing the AI
// response text and three action buttons: resolve, continue, and escalate.
// The questionID is embedded in each button's value so callbacks can identify
// the target question.
func BuildQuestionResponseBlocks(aiResponse string, questionID string) []slack.Block {
  textBlock := slack.NewSectionBlock(
    slack.NewTextBlockObject("mrkdwn", aiResponse, false, false),
    nil, nil,
  )

  resolveBtn := slack.NewButtonBlockElement(ActionIDQuestionResolved, questionID,
    slack.NewTextBlockObject("plain_text", "✅ 解決した", false, false),
  )
  resolveBtn.Style = slack.StylePrimary

  continueBtn := slack.NewButtonBlockElement(ActionIDQuestionContinue, questionID,
    slack.NewTextBlockObject("plain_text", "💬 続けて質問", false, false),
  )

  escalateBtn := slack.NewButtonBlockElement(ActionIDQuestionEscalate, questionID,
    slack.NewTextBlockObject("plain_text", "🙋 メンターに相談", false, false),
  )
  escalateBtn.Style = slack.StyleDanger

  actionBlock := slack.NewActionBlock("", resolveBtn, continueBtn, escalateBtn)

  return []slack.Block{textBlock, actionBlock}
}
