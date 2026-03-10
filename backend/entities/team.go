package entities

import "time"

type ChannelPurpose string

const (
	ChannelPurposeProgress ChannelPurpose = "progress"
	ChannelPurposeQuestion ChannelPurpose = "question"
	ChannelPurposeNotice   ChannelPurpose = "notice"
)

// Team represents a hackathon team.
type Team struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
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

// TeamChannel links a team to an external channel with a specific purpose.
type TeamChannel struct {
	ID        string         `json:"id" db:"id"`
	TeamID    string         `json:"team_id" db:"team_id"`
	SlackChID string         `json:"slack_channel_id" db:"slack_channel_id"`
	Purpose   ChannelPurpose `json:"purpose" db:"purpose"`
	IsPrimary bool           `json:"is_primary" db:"is_primary"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}
