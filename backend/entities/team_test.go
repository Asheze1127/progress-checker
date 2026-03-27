package entities

import (
	"testing"
)

func TestTeamValidate(t *testing.T) {
	tests := []struct {
		name           string
		team           Team
		wantErrStrings []string
	}{
		{
			name: "valid team",
			team: Team{
				ID:   TeamID("team-1"),
				Name: "Alpha",
				SlackChannels: []SlackChannel{
					{
						ID:                   SlackChannelID("channel-1"),
						SlackChannelPurposes: []SlackChannelPurpose{SlackChannelPurposeProgress, SlackChannelPurposeQuestion},
					},
				},
			},
		},
		{
			name: "valid team without channels",
			team: Team{
				ID:   TeamID("team-1"),
				Name: "Alpha",
			},
		},
		{
			name: "empty id",
			team: Team{
				Name: "Alpha",
			},
			wantErrStrings: []string{"team.id is required"},
		},
		{
			name: "empty name",
			team: Team{
				ID: TeamID("team-1"),
			},
			wantErrStrings: []string{"team.name is required"},
		},
		{
			name: "empty channel id",
			team: Team{
				ID:   TeamID("team-1"),
				Name: "Alpha",
				SlackChannels: []SlackChannel{
					{
						SlackChannelPurposes: []SlackChannelPurpose{SlackChannelPurposeProgress},
					},
				},
			},
			wantErrStrings: []string{"team.slack_channels[0].id is required"},
		},
		{
			name: "duplicate slack channel ids",
			team: Team{
				ID:   TeamID("team-1"),
				Name: "Alpha",
				SlackChannels: []SlackChannel{
					{
						ID:                   SlackChannelID("channel-1"),
						SlackChannelPurposes: []SlackChannelPurpose{SlackChannelPurposeProgress},
					},
					{
						ID:                   SlackChannelID("channel-1"),
						SlackChannelPurposes: []SlackChannelPurpose{SlackChannelPurposeQuestion},
					},
				},
			},
			wantErrStrings: []string{`team.slack_channels.id contains duplicate value "channel-1"`},
		},
		{
			name: "empty purposes",
			team: Team{
				ID:   TeamID("team-1"),
				Name: "Alpha",
				SlackChannels: []SlackChannel{
					{
						ID: SlackChannelID("channel-1"),
					},
				},
			},
			wantErrStrings: []string{"team.slack_channels[0].purposes must not be empty"},
		},
		{
			name: "invalid purpose",
			team: Team{
				ID:   TeamID("team-1"),
				Name: "Alpha",
				SlackChannels: []SlackChannel{
					{
						ID:                   SlackChannelID("channel-1"),
						SlackChannelPurposes: []SlackChannelPurpose{"invalid"},
					},
				},
			},
			wantErrStrings: []string{"team.slack_channels[0].purposes[0] must be one of progress, question, notice"},
		},
		{
			name: "duplicate purposes",
			team: Team{
				ID:   TeamID("team-1"),
				Name: "Alpha",
				SlackChannels: []SlackChannel{
					{
						ID:                   SlackChannelID("channel-1"),
						SlackChannelPurposes: []SlackChannelPurpose{SlackChannelPurposeProgress, SlackChannelPurposeProgress},
					},
				},
			},
			wantErrStrings: []string{`team.slack_channels[0].purposes contains duplicate value "progress"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.team.Validate()
			assertValidationResult(t, err, tt.wantErrStrings)
		})
	}
}
