package service

import (
	"context"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// SlackNotifier defines operations for sending Slack messages.
type SlackNotifier interface {
	// PostToMentorChannel posts a formatted question summary to the mentor channel.
	PostToMentorChannel(ctx context.Context, question *entities.Question) error
}
