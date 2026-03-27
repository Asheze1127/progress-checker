package entities

import (
	"testing"
	"time"
)

func TestProgressLogValidate(t *testing.T) {
	validTime := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		progressLog    ProgressLog
		wantErrStrings []string
	}{
		{
			name: "valid progress log",
			progressLog: ProgressLog{
				ID:            ProgressLogID("progress-1"),
				ParticipantID: ParticipantID("participant-1"),
				ProgressBodies: []ProgressBody{
					{Phase: ProgressPhaseCoding, SubmittedAt: validTime},
				},
			},
		},
		{
			name: "valid progress log with multiple bodies",
			progressLog: ProgressLog{
				ID:            ProgressLogID("progress-1"),
				ParticipantID: ParticipantID("participant-1"),
				ProgressBodies: []ProgressBody{
					{Phase: ProgressPhaseIdea, SubmittedAt: validTime},
					{Phase: ProgressPhaseDesign, SubmittedAt: validTime},
					{Phase: ProgressPhaseCoding, SubmittedAt: validTime},
					{Phase: ProgressPhaseTesting, SubmittedAt: validTime},
					{Phase: ProgressPhaseDemo, SubmittedAt: validTime},
				},
			},
		},
		{
			name: "empty id",
			progressLog: ProgressLog{
				ParticipantID: ParticipantID("participant-1"),
				ProgressBodies: []ProgressBody{
					{Phase: ProgressPhaseCoding, SubmittedAt: validTime},
				},
			},
			wantErrStrings: []string{"progress_log.id is required"},
		},
		{
			name: "empty participant_id",
			progressLog: ProgressLog{
				ID: ProgressLogID("progress-1"),
				ProgressBodies: []ProgressBody{
					{Phase: ProgressPhaseCoding, SubmittedAt: validTime},
				},
			},
			wantErrStrings: []string{"progress_log.participant_id is required"},
		},
		{
			name: "empty progress_bodies",
			progressLog: ProgressLog{
				ID:            ProgressLogID("progress-1"),
				ParticipantID: ParticipantID("participant-1"),
			},
			wantErrStrings: []string{"progress_log.progress_bodies must not be empty"},
		},
		{
			name: "invalid phase in body",
			progressLog: ProgressLog{
				ID:            ProgressLogID("progress-1"),
				ParticipantID: ParticipantID("participant-1"),
				ProgressBodies: []ProgressBody{
					{Phase: ProgressPhase("unknown"), SubmittedAt: validTime},
				},
			},
			wantErrStrings: []string{"progress_log.progress_bodies[0].phase must be one of idea, design, coding, testing, demo"},
		},
		{
			name: "zero submitted_at in body",
			progressLog: ProgressLog{
				ID:            ProgressLogID("progress-1"),
				ParticipantID: ParticipantID("participant-1"),
				ProgressBodies: []ProgressBody{
					{Phase: ProgressPhaseCoding},
				},
			},
			wantErrStrings: []string{"progress_log.progress_bodies[0].submitted_at must be set"},
		},
		{
			name: "invalid body with both errors",
			progressLog: ProgressLog{
				ID:            ProgressLogID("progress-1"),
				ParticipantID: ParticipantID("participant-1"),
				ProgressBodies: []ProgressBody{
					{},
				},
			},
			wantErrStrings: []string{
				"progress_log.progress_bodies[0].phase must be one of",
				"progress_log.progress_bodies[0].submitted_at must be set",
			},
		},
		{
			name: "multiple bodies with errors at different indices",
			progressLog: ProgressLog{
				ID:            ProgressLogID("progress-1"),
				ParticipantID: ParticipantID("participant-1"),
				ProgressBodies: []ProgressBody{
					{Phase: ProgressPhaseCoding, SubmittedAt: validTime},
					{Phase: ProgressPhase("bad"), SubmittedAt: validTime},
				},
			},
			wantErrStrings: []string{"progress_log.progress_bodies[1].phase must be one of"},
		},
		{
			name:        "all fields missing",
			progressLog: ProgressLog{},
			wantErrStrings: []string{
				"progress_log.id is required",
				"progress_log.participant_id is required",
				"progress_log.progress_bodies must not be empty",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.progressLog.Validate()
			assertValidationResult(t, err, tt.wantErrStrings)
		})
	}
}

func TestProgressBodyValidate(t *testing.T) {
	validTime := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		progressBody   ProgressBody
		wantErrStrings []string
	}{
		{
			name:         "valid body",
			progressBody: ProgressBody{Phase: ProgressPhaseCoding, SubmittedAt: validTime},
		},
		{
			name:           "invalid phase",
			progressBody:   ProgressBody{Phase: ProgressPhase("bad"), SubmittedAt: validTime},
			wantErrStrings: []string{"progress_body.phase must be one of idea, design, coding, testing, demo"},
		},
		{
			name:           "empty phase",
			progressBody:   ProgressBody{SubmittedAt: validTime},
			wantErrStrings: []string{"progress_body.phase must be one of idea, design, coding, testing, demo"},
		},
		{
			name:           "zero submitted_at",
			progressBody:   ProgressBody{Phase: ProgressPhaseCoding},
			wantErrStrings: []string{"progress_body.submitted_at must be set"},
		},
		{
			name:         "all fields missing",
			progressBody: ProgressBody{},
			wantErrStrings: []string{
				"progress_body.phase must be one of idea, design, coding, testing, demo",
				"progress_body.submitted_at must be set",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.progressBody.Validate()
			assertValidationResult(t, err, tt.wantErrStrings)
		})
	}
}
