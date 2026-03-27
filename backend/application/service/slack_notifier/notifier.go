package slacknotifier

import (
	"context"
	"log/slog"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

type SlackNotifier interface {
	PostToMentorChannel(ctx context.Context, question *entities.Question) error
}

type NoopSlackNotifier struct{}

func (n *NoopSlackNotifier) PostToMentorChannel(_ context.Context, question *entities.Question) error {
	slog.Warn("slack notifier not configured, skipping mentor notification", slog.String("question_id", string(question.ID)))
	return nil
}
