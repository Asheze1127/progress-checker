package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type ProgressPhase string

const (
	ProgressPhaseIdea    ProgressPhase = "idea"
	ProgressPhaseDesign  ProgressPhase = "design"
	ProgressPhaseCoding  ProgressPhase = "coding"
	ProgressPhaseTesting ProgressPhase = "testing"
	ProgressPhaseDemo    ProgressPhase = "demo"
)

// IsValid returns true if the phase is a known valid value.
func (p ProgressPhase) IsValid() bool {
	switch p {
	case ProgressPhaseIdea, ProgressPhaseDesign, ProgressPhaseCoding, ProgressPhaseTesting, ProgressPhaseDemo:
		return true
	default:
		return false
	}
}

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

func (p ProgressLog) Validate() error {
	var errs []error

	if strings.TrimSpace(string(p.ID)) == "" {
		errs = append(errs, fmt.Errorf("progress_log.id is required"))
	}

	if strings.TrimSpace(string(p.ParticipantID)) == "" {
		errs = append(errs, fmt.Errorf("progress_log.participant_id is required"))
	}

	if len(p.ProgressBodies) == 0 {
		errs = append(errs, fmt.Errorf("progress_log.progress_bodies must not be empty"))
	}

	for i, progressBody := range p.ProgressBodies {
		switch progressBody.Phase {
		case ProgressPhaseIdea, ProgressPhaseDesign, ProgressPhaseCoding, ProgressPhaseTesting, ProgressPhaseDemo:
		default:
			errs = append(errs, fmt.Errorf("progress_log.progress_bodies[%d].phase must be one of idea, design, coding, testing, demo", i))
		}

		if progressBody.SubmittedAt.IsZero() {
			errs = append(errs, fmt.Errorf("progress_log.progress_bodies[%d].submitted_at must be set", i))
		}
	}

	return errors.Join(errs...)
}

func (p ProgressBody) Validate() error {
	var errs []error

	switch p.Phase {
	case ProgressPhaseIdea, ProgressPhaseDesign, ProgressPhaseCoding, ProgressPhaseTesting, ProgressPhaseDemo:
	default:
		errs = append(errs, fmt.Errorf("progress_body.phase must be one of idea, design, coding, testing, demo"))
	}

	if p.SubmittedAt.IsZero() {
		errs = append(errs, fmt.Errorf("progress_body.submitted_at must be set"))
	}

	return errors.Join(errs...)
}
