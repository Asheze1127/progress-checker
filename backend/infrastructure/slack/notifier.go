package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/slack-go/slack"

	slacknotifier "github.com/Asheze1127/progress-checker/backend/application/service/slack_notifier"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// Compile-time check that MentorNotifier implements slacknotifier.SlackNotifier.
var _ slacknotifier.SlackNotifier = (*MentorNotifier)(nil)

// MentorNotifier sends notifications to a mentor Slack channel.
type MentorNotifier struct {
	api             *slack.Client
	mentorChannelID string
}

// NewMentorNotifier creates a new MentorNotifier.
func NewMentorNotifier(api *slack.Client, mentorChannelID string) *MentorNotifier {
	return &MentorNotifier{
		api:             api,
		mentorChannelID: mentorChannelID,
	}
}

// PostToMentorChannel posts a notification about an escalated question to the mentor channel.
func (n *MentorNotifier) PostToMentorChannel(ctx context.Context, question *entities.Question) error {
	title := sanitizeSlackText(string(question.Title))
	participantID := sanitizeSlackText(string(question.ParticipantID))
	text := fmt.Sprintf("New escalated question from participant %s: %s", participantID, title)
	_, _, err := n.api.PostMessageContext(ctx, n.mentorChannelID, slack.MsgOptionText(text, false))
	if err != nil {
		return fmt.Errorf("failed to post to mentor channel: %w", err)
	}
	return nil
}

// sanitizeSlackText escapes Slack mrkdwn special characters to prevent injection.
func sanitizeSlackText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
