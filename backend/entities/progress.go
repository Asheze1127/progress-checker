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
	ID             string
	TeamID         string
	PostedByUserID string
	SlackMsgTS     string
	CreatedAt      time.Time
	Bodies         []ProgressBody
}

// ProgressBody represents a structured section within a progress submission.
type ProgressBody struct {
	ID            string
	ProgressLogID string
	Order         int
	Phase         ProgressPhase
	SOS           bool
	Comment       string
}
