package entities

import (
  "testing"
)

func TestQuestionValidate(t *testing.T) {
  tests := []struct {
    name           string
    question       Question
    wantErrStrings []string
  }{
    {
      name: "valid question with assigned_mentor status",
      question: Question{
        ID:             QuestionID("question-1"),
        ParticipantID:  ParticipantID("participant-1"),
        MentorIDs:      []MentorID{"mentor-1"},
        Title:          "Need help with build",
        SlackChannelID: SlackChannelID("channel-1"),
        Status:         QuestionStatusAssignedMentor,
        SlackThreadTS:  "1742731200.000100",
      },
    },
    {
      name: "empty id",
      question: Question{
        ParticipantID:  ParticipantID("participant-1"),
        Title:          "Need help with build",
        SlackChannelID: SlackChannelID("channel-1"),
        Status:         QuestionStatusOpen,
        SlackThreadTS:  "1742731200.000100",
      },
      wantErrStrings: []string{"question.id is required"},
    },
    {
      name: "empty participant_id",
      question: Question{
        ID:             QuestionID("question-1"),
        Title:          "Need help with build",
        SlackChannelID: SlackChannelID("channel-1"),
        Status:         QuestionStatusOpen,
        SlackThreadTS:  "1742731200.000100",
      },
      wantErrStrings: []string{"question.participant_id is required"},
    },
    {
      name: "empty title",
      question: Question{
        ID:             QuestionID("question-1"),
        ParticipantID:  ParticipantID("participant-1"),
        SlackChannelID: SlackChannelID("channel-1"),
        Status:         QuestionStatusOpen,
        SlackThreadTS:  "1742731200.000100",
      },
      wantErrStrings: []string{"question.title is required"},
    },
    {
      name: "empty slack_channel_id",
      question: Question{
        ID:            QuestionID("question-1"),
        ParticipantID: ParticipantID("participant-1"),
        Title:         "Need help with build",
        Status:        QuestionStatusOpen,
        SlackThreadTS: "1742731200.000100",
      },
      wantErrStrings: []string{"question.slack_channel_id is required"},
    },
    {
      name: "invalid status",
      question: Question{
        ID:             QuestionID("question-1"),
        ParticipantID:  ParticipantID("participant-1"),
        Title:          "Need help with build",
        SlackChannelID: SlackChannelID("channel-1"),
        Status:         QuestionStatus("invalid"),
        SlackThreadTS:  "1742731200.000100",
      },
      wantErrStrings: []string{"question.status must be one of open, in_progress, awaiting_user, assigned_mentor, resolved"},
    },
    {
      name: "empty status",
      question: Question{
        ID:             QuestionID("question-1"),
        ParticipantID:  ParticipantID("participant-1"),
        Title:          "Need help with build",
        SlackChannelID: SlackChannelID("channel-1"),
        SlackThreadTS:  "1742731200.000100",
      },
      wantErrStrings: []string{"question.status must be one of open, in_progress, awaiting_user, assigned_mentor, resolved"},
    },
    {
      name: "empty slack_thread_ts",
      question: Question{
        ID:             QuestionID("question-1"),
        ParticipantID:  ParticipantID("participant-1"),
        Title:          "Need help with build",
        SlackChannelID: SlackChannelID("channel-1"),
        Status:         QuestionStatusOpen,
      },
      wantErrStrings: []string{"question.slack_thread_ts is required"},
    },
    {
      name: "empty mentor_id in array",
      question: Question{
        ID:             QuestionID("question-1"),
        ParticipantID:  ParticipantID("participant-1"),
        MentorIDs:      []MentorID{"mentor-1", ""},
        Title:          "Need help with build",
        SlackChannelID: SlackChannelID("channel-1"),
        Status:         QuestionStatusOpen,
        SlackThreadTS:  "1742731200.000100",
      },
      wantErrStrings: []string{"question.mentor_ids[1] is required"},
    },
    {
      name: "duplicate mentor ids",
      question: Question{
        ID:             QuestionID("question-1"),
        ParticipantID:  ParticipantID("participant-1"),
        MentorIDs:      []MentorID{"mentor-1", "mentor-1"},
        Title:          "Need help with build",
        SlackChannelID: SlackChannelID("channel-1"),
        Status:         QuestionStatusOpen,
        SlackThreadTS:  "1742731200.000100",
      },
      wantErrStrings: []string{`question.mentor_ids contains duplicate value "mentor-1"`},
    },
    {
      name: "assigned_mentor requires mentor_ids",
      question: Question{
        ID:             QuestionID("question-1"),
        ParticipantID:  ParticipantID("participant-1"),
        Title:          "Need help with build",
        SlackChannelID: SlackChannelID("channel-1"),
        Status:         QuestionStatusAssignedMentor,
        SlackThreadTS:  "1742731200.000100",
      },
      wantErrStrings: []string{`question.mentor_ids must not be empty when status is "assigned_mentor"`},
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      err := tt.question.Validate()
      assertValidationResult(t, err, tt.wantErrStrings)
    })
  }
}

func TestCanTransitionTo(t *testing.T) {
  tests := []struct {
    name   string
    from   QuestionStatus
    to     QuestionStatus
    expect bool
  }{
    {"open to resolved", QuestionStatusOpen, QuestionStatusResolved, true},
    {"open to in_progress", QuestionStatusOpen, QuestionStatusInProgress, true},
    {"open to assigned_mentor", QuestionStatusOpen, QuestionStatusAssignedMentor, true},
    {"in_progress to resolved", QuestionStatusInProgress, QuestionStatusResolved, true},
    {"in_progress to assigned_mentor", QuestionStatusInProgress, QuestionStatusAssignedMentor, true},
    {"in_progress to open", QuestionStatusInProgress, QuestionStatusOpen, false},
    {"assigned_mentor to resolved", QuestionStatusAssignedMentor, QuestionStatusResolved, true},
    {"assigned_mentor to open", QuestionStatusAssignedMentor, QuestionStatusOpen, false},
    {"assigned_mentor to in_progress", QuestionStatusAssignedMentor, QuestionStatusInProgress, false},
    {"resolved to open", QuestionStatusResolved, QuestionStatusOpen, false},
    {"resolved to in_progress", QuestionStatusResolved, QuestionStatusInProgress, false},
    {"resolved to assigned_mentor", QuestionStatusResolved, QuestionStatusAssignedMentor, false},
    {"invalid status", QuestionStatus("invalid"), QuestionStatusResolved, false},
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      q := &Question{Status: tt.from}
      got := q.CanTransitionTo(tt.to)
      if got != tt.expect {
        t.Errorf("CanTransitionTo(%q -> %q) = %v, want %v", tt.from, tt.to, got, tt.expect)
      }
    })
  }
}
