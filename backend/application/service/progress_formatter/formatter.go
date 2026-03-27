package progressformatter

import (
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

type ProgressFormatter struct{}

func NewProgressFormatter() *ProgressFormatter { return &ProgressFormatter{} }

func (f *ProgressFormatter) FormatSlackMessage(teamID string, log *entities.ProgressLog) string {
	if len(log.ProgressBodies) == 0 {
		return fmt.Sprintf(":bar_chart: 進捗報告\nチーム: %s", teamID)
	}
	body := log.ProgressBodies[0]
	sosEmoji := ""
	if body.SOS {
		sosEmoji = " :sos:"
	}
	return fmt.Sprintf(":bar_chart: 進捗報告\nチーム: %s\nフェーズ: %s%s\nコメント: %s", teamID, string(body.Phase), sosEmoji, body.Comment)
}
