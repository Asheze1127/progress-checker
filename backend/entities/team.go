package entities

import "time"

// Team represents a hackathon team.
type Team struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	SlackChID string    `json:"slack_channel_id" db:"slack_channel_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TeamMember represents the membership relation between a user and a team.
type TeamMember struct {
	ID       string    `json:"id" db:"id"`
	TeamID   string    `json:"team_id" db:"team_id"`
	UserID   string    `json:"user_id" db:"user_id"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}
