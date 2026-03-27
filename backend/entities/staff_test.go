package entities

import (
  "testing"
)

func TestStaffValidate(t *testing.T) {
  slackUserID := SlackUserID("U123")
  blankSlackUserID := SlackUserID("")
  tests := []struct {
    name           string
    staff          Staff
    wantErrStrings []string
  }{
    {
      name: "valid staff without slack user",
      staff: Staff{
        ID:    StaffID("staff-1"),
        Name:  "Operations",
        Email: "ops@example.com",
      },
    },
    {
      name: "valid staff with slack user",
      staff: Staff{
        ID:          StaffID("staff-1"),
        SlackUserID: &slackUserID,
        Name:        "Operations",
        Email:       "ops@example.com",
      },
    },
    {
      name: "empty id",
      staff: Staff{
        Name:  "Operations",
        Email: "ops@example.com",
      },
      wantErrStrings: []string{"staff.id is required"},
    },
    {
      name: "empty slack user id",
      staff: Staff{
        ID:          StaffID("staff-1"),
        SlackUserID: &blankSlackUserID,
        Name:        "Operations",
        Email:       "ops@example.com",
      },
      wantErrStrings: []string{"staff.slack_user_id is required"},
    },
    {
      name: "empty name",
      staff: Staff{
        ID:    StaffID("staff-1"),
        Email: "ops@example.com",
      },
      wantErrStrings: []string{"staff.name is required"},
    },
    {
      name: "empty email",
      staff: Staff{
        ID:   StaffID("staff-1"),
        Name: "Operations",
      },
      wantErrStrings: []string{"staff.email is required"},
    },
    {
      name: "invalid email without at mark",
      staff: Staff{
        ID:    StaffID("staff-1"),
        Name:  "Operations",
        Email: "ops.example.com",
      },
      wantErrStrings: []string{"staff.email must be a valid email address"},
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      err := tt.staff.Validate()
      assertValidationResult(t, err, tt.wantErrStrings)
    })
  }
}
