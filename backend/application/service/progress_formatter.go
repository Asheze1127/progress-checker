package service

import (
	"fmt"

	"github.com/samber/do/v2"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// ProgressFormatter formats progress data into Slack message text.
type ProgressFormatter struct{}

// NewProgressFormatter creates a new ProgressFormatter via DI container.
func NewProgressFormatter(_ do.Injector) (*ProgressFormatter, error) {
	return &ProgressFormatter{}, nil
}

// FormatSlackMessage builds a Slack message string from a progress log and team ID.
func (f *ProgressFormatter) FormatSlackMessage(teamID string, log *entities.ProgressLog) string {
	if len(log.ProgressBodies) == 0 {
		return fmt.Sprintf(":bar_chart: 進捗報告\nチーム: %s", teamID)
	}

	body := log.ProgressBodies[0]

	sosEmoji := ""
	if body.SOS {
		sosEmoji = " :sos:"
	}

	return fmt.Sprintf(
		":bar_chart: 進捗報告\nチーム: %s\nフェーズ: %s%s\nコメント: %s",
		teamID,
		string(body.Phase),
		sosEmoji,
		body.Comment,
	)
}
