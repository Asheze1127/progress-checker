package entities

import "testing"

func TestParticipantValidate(t *testing.T) {
	tests := []struct {
		name           string
		participant    Participant
		wantErrStrings []string
	}{
		{
			name: "valid participant",
			participant: Participant{
				id:     UserID("user-1"),
				TeamID: TeamID("team-1"),
			},
		},
		{
			name: "empty id",
			participant: Participant{
				TeamID: TeamID("team-1"),
			},
			wantErrStrings: []string{"participant.id is required"},
		},
		{
			name: "empty team_id",
			participant: Participant{
				id: UserID("user-1"),
			},
			wantErrStrings: []string{"participant.team_id is required"},
		},
		{
			name:           "all fields missing",
			participant:    Participant{},
			wantErrStrings: []string{"participant.id is required", "participant.team_id is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.participant.Validate()
			assertValidationResult(t, err, tt.wantErrStrings)
		})
	}
}
