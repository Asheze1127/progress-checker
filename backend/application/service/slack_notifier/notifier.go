package slacknotifier

import (
	"context"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// SlackNotifier sends notifications to Slack channels.
type SlackNotifier interface {
	PostToMentorChannel(ctx context.Context, question *entities.Question) error
}
