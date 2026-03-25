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
	ID                string
	Name              string
	TeamMembers       []TeamMember
	MentorAssignments []MentorAssignment
	TeamChannels      []TeamChannel
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// TeamMember represents the membership relation between a user and a team.
type TeamMember struct {
	ID       string
	TeamID   string
	UserID   string
	JoinedAt time.Time
}

// MentorAssignment represents which mentor is responsible for which team.
type MentorAssignment struct {
	ID           string
	TeamID       string
	MentorUserID string
	CreatedAt    time.Time
}

// TeamChannel links a team to an external channel with a specific purpose.
// MVP keeps a single channel for each team-purpose pair.
type TeamChannel struct {
	ID             string
	TeamID         string
	SlackChannelID string
	Purpose        ChannelPurpose
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
