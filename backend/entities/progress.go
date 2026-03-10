package entities

import "time"

type ProgressPhase string

const (
	ProgressPhaseIdea    ProgressPhase = "idea"
	ProgressPhaseDesign  ProgressPhase = "design"
	ProgressPhaseCoding  ProgressPhase = "coding"
	ProgressPhaseTesting ProgressPhase = "testing"
	ProgressPhaseDemo    ProgressPhase = "demo"
)

// ProgressLog records a single progress update submitted via /progress.
type ProgressLog struct {
	ID             string        `json:"id" db:"id"`
	TeamID         string        `json:"team_id" db:"team_id"`
	PostedByUserID string        `json:"posted_by_user_id" db:"posted_by_user_id"`
	Phase          ProgressPhase `json:"phase" db:"phase"`
	SOS            bool          `json:"sos" db:"sos"`
	Comment        string        `json:"comment" db:"comment"`
	SlackMsgTS     string        `json:"slack_msg_ts" db:"slack_msg_ts"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
}
