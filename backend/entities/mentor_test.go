package entities

import (
	"strings"
	"testing"
)

func TestMentorValidate(t *testing.T) {
	tests := []struct {
		name           string
		mentor         Mentor
		wantErrStrings []string
	}{
		{
			name: "valid mentor with teams",
			mentor: Mentor{
				id:      UserID("user-1"),
				TeamIDs: []TeamID{"team-1", "team-2"},
			},
		},
		{
			name: "valid mentor without teams",
			mentor: Mentor{
				id: UserID("user-1"),
			},
		},
		{
			name:           "empty id",
			mentor:         Mentor{},
			wantErrStrings: []string{"mentor.id is required"},
		},
		{
			name: "empty team_id in array",
			mentor: Mentor{
				id:      UserID("user-1"),
				TeamIDs: []TeamID{"team-1", ""},
			},
			wantErrStrings: []string{"mentor.team_ids[1] is required"},
		},
		{
			name: "duplicate team ids",
			mentor: Mentor{
				id:      UserID("user-1"),
				TeamIDs: []TeamID{"team-1", "team-1"},
			},
			wantErrStrings: []string{`mentor.team_ids contains duplicate value "team-1"`},
		},
		{
			name: "multiple errors",
			mentor: Mentor{
				TeamIDs: []TeamID{"", "team-1", "team-1"},
			},
			wantErrStrings: []string{
				"mentor.id is required",
				"mentor.team_ids[0] is required",
				`mentor.team_ids contains duplicate value "team-1"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mentor.Validate()
			assertValidationResult(t, err, tt.wantErrStrings)
		})
	}
}

func assertValidationResult(t *testing.T, err error, wantErrStrings []string) {
	t.Helper()

	if len(wantErrStrings) == 0 {
		if err != nil {
			t.Fatalf("Validate() error = %v", err)
		}

		return
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	for _, wantErrString := range wantErrStrings {
		if !strings.Contains(err.Error(), wantErrString) {
			t.Fatalf("error %q does not contain expected message %q", err.Error(), wantErrString)
		}
	}
}
