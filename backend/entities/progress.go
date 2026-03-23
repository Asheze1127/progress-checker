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

type ProgressLogID string

type ProgressLog struct {
	ID             ProgressLogID
	ParticipantID  ParticipantID
	ProgressBodies []ProgressBody
}

type ProgressBody struct {
	Phase       ProgressPhase
	SOS         bool
	Comment     string
	SubmittedAt time.Time
}
