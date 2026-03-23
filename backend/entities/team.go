package entities

type SlackChannelPurpose string

const (
	SlackChannelPurposeProgress SlackChannelPurpose = "progress"
	SlackChannelPurposeQuestion SlackChannelPurpose = "question"
	SlackChannelPurposeNotice   SlackChannelPurpose = "notice"
)

type TeamID string

type Team struct {
	ID            TeamID
	Name          string
	SlackChannels []SlackChannel
}

type SlackChannel struct {
	ID                   SlackChannelID
	SlackChannelPurposes []SlackChannelPurpose
}
