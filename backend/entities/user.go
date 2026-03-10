package entities

import "time"

type UserRole string

const (
	UserRoleParticipant UserRole = "participant"
	UserRoleMentor      UserRole = "mentor"
)

// User represents a Slack workspace member (participant or mentor).
type User struct {
	ID          string    `json:"id" db:"id"`
	SlackUserID string    `json:"slack_user_id" db:"slack_user_id"`
	Name        string    `json:"name" db:"name"`
	Email       string    `json:"email" db:"email"`
	Role        UserRole  `json:"role" db:"role"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
