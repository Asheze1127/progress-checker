package entities

import (
  "errors"
  "fmt"
  "strings"
)

type SlackChannelPurpose string

const (
  SlackChannelPurposeProgress SlackChannelPurpose = "progress"
  SlackChannelPurposeQuestion SlackChannelPurpose = "question"
  SlackChannelPurposeNotice   SlackChannelPurpose = "notice"
)

type TeamID string

type SlackChannelID string

type Team struct {
  ID            TeamID
  Name          string
  SlackChannels []SlackChannel
}

type SlackChannel struct {
  ID                   SlackChannelID
  SlackChannelPurposes []SlackChannelPurpose
}

func (t Team) Validate() error {
  var errs []error

  if strings.TrimSpace(string(t.ID)) == "" {
    errs = append(errs, fmt.Errorf("team.id is required"))
  }

  if strings.TrimSpace(t.Name) == "" {
    errs = append(errs, fmt.Errorf("team.name is required"))
  }

  seenChannelIDs := make(map[SlackChannelID]struct{}, len(t.SlackChannels))
  for i, slackChannel := range t.SlackChannels {
    if strings.TrimSpace(string(slackChannel.ID)) == "" {
      errs = append(errs, fmt.Errorf("team.slack_channels[%d].id is required", i))
    }

    if _, ok := seenChannelIDs[slackChannel.ID]; ok {
      errs = append(errs, fmt.Errorf("team.slack_channels.id contains duplicate value %q", slackChannel.ID))
    } else {
      seenChannelIDs[slackChannel.ID] = struct{}{}
    }

    if len(slackChannel.SlackChannelPurposes) == 0 {
      errs = append(errs, fmt.Errorf("team.slack_channels[%d].purposes must not be empty", i))
    }

    seenPurposes := make(map[SlackChannelPurpose]struct{}, len(slackChannel.SlackChannelPurposes))
    for j, purpose := range slackChannel.SlackChannelPurposes {
      switch purpose {
      case SlackChannelPurposeProgress, SlackChannelPurposeQuestion, SlackChannelPurposeNotice:
      default:
        errs = append(errs, fmt.Errorf("team.slack_channels[%d].purposes[%d] must be one of progress, question, notice", i, j))
      }

      if _, ok := seenPurposes[purpose]; ok {
        errs = append(errs, fmt.Errorf("team.slack_channels[%d].purposes contains duplicate value %q", i, purpose))
        continue
      }

      seenPurposes[purpose] = struct{}{}
    }
  }

  return errors.Join(errs...)
}
