package entities

import "testing"

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name           string
		user           User
		wantErrStrings []string
	}{
		{
			name: "empty id",
			user: User{
				SlackUserID: SlackUserID("U123"),
				Name:        "Alice",
				Email:       "alice@example.com",
				Role:        UserRoleMentor,
			},
			wantErrStrings: []string{"user.id is required"},
		},
		{
			name: "empty slack_user_id",
			user: User{
				ID:    UserID("user-1"),
				Name:  "Alice",
				Email: "alice@example.com",
				Role:  UserRoleMentor,
			},
			wantErrStrings: []string{"user.slack_user_id is required"},
		},
		{
			name: "empty name",
			user: User{
				ID:          UserID("user-1"),
				SlackUserID: SlackUserID("U123"),
				Email:       "alice@example.com",
				Role:        UserRoleMentor,
			},
			wantErrStrings: []string{"user.name is required"},
		},
		{
			name: "empty email",
			user: User{
				ID:          UserID("user-1"),
				SlackUserID: SlackUserID("U123"),
				Name:        "Alice",
				Role:        UserRoleMentor,
			},
			wantErrStrings: []string{"user.email is required"},
		},
		{
			name: "invalid email without at mark",
			user: User{
				ID:          UserID("user-1"),
				SlackUserID: SlackUserID("U123"),
				Name:        "Alice",
				Email:       "alice.example.com",
				Role:        UserRoleMentor,
			},
			wantErrStrings: []string{"user.email must be a valid email address"},
		},
		{
			name: "invalid role",
			user: User{
				ID:          UserID("user-1"),
				SlackUserID: SlackUserID("U123"),
				Name:        "Alice",
				Email:       "alice@example.com",
				Role:        UserRole("admin"),
			},
			wantErrStrings: []string{"user.role must be one of participant, mentor"},
		},
		{
			name: "empty role",
			user: User{
				ID:          UserID("user-1"),
				SlackUserID: SlackUserID("U123"),
				Name:        "Alice",
				Email:       "alice@example.com",
			},
			wantErrStrings: []string{"user.role must be one of participant, mentor"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			assertValidationResult(t, err, tt.wantErrStrings)
		})
	}
}
