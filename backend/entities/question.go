package entities

import (
	"errors"
	"fmt"
	"strings"
)

type QuestionStatus string

const (
	QuestionStatusOpen           QuestionStatus = "open"
	QuestionStatusInProgress     QuestionStatus = "in_progress"
	QuestionStatusAssignedMentor QuestionStatus = "assigned_mentor"
	QuestionStatusResolved       QuestionStatus = "resolved"
)

type QuestionID string

type Question struct {
	ID             QuestionID
	ParticipantID  ParticipantID
	MentorIDs      []MentorID
	Title          string
	SlackChannelID SlackChannelID
	Status         QuestionStatus
	SlackThreadTS  string
}

func (q Question) Validate() error {
	var errs []error

	if strings.TrimSpace(string(q.ID)) == "" {
		errs = append(errs, fmt.Errorf("question.id is required"))
	}

	if strings.TrimSpace(string(q.ParticipantID)) == "" {
		errs = append(errs, fmt.Errorf("question.participant_id is required"))
	}

	if strings.TrimSpace(q.Title) == "" {
		errs = append(errs, fmt.Errorf("question.title is required"))
	}

	if strings.TrimSpace(string(q.SlackChannelID)) == "" {
		errs = append(errs, fmt.Errorf("question.slack_channel_id is required"))
	}

	switch q.Status {
	case QuestionStatusOpen, QuestionStatusInProgress, QuestionStatusAssignedMentor, QuestionStatusResolved:
	default:
		errs = append(errs, fmt.Errorf("question.status must be one of open, in_progress, assigned_mentor, resolved"))
	}

	if strings.TrimSpace(q.SlackThreadTS) == "" {
		errs = append(errs, fmt.Errorf("question.slack_thread_ts is required"))
	}

	seenMentorIDs := make(map[MentorID]struct{}, len(q.MentorIDs))

	for i, mentorID := range q.MentorIDs {
		if strings.TrimSpace(string(mentorID)) == "" {
			errs = append(errs, fmt.Errorf("question.mentor_ids[%d] is required", i))
		}

		if _, ok := seenMentorIDs[mentorID]; ok {
			errs = append(errs, fmt.Errorf("question.mentor_ids contains duplicate value %q", mentorID))
			continue
		}

		seenMentorIDs[mentorID] = struct{}{}
	}

	if q.Status == QuestionStatusAssignedMentor && len(q.MentorIDs) == 0 {
		errs = append(errs, fmt.Errorf("question.mentor_ids must not be empty when status is %q", q.Status))
	}

	return errors.Join(errs...)
}

// validTransitions defines the allowed status transitions for a question.
var validTransitions = map[QuestionStatus][]QuestionStatus{
	QuestionStatusOpen:           {QuestionStatusResolved, QuestionStatusInProgress, QuestionStatusAssignedMentor},
	QuestionStatusInProgress:     {QuestionStatusResolved, QuestionStatusAssignedMentor},
	QuestionStatusAssignedMentor: {QuestionStatusResolved},
	QuestionStatusResolved:       {},
}

// CanTransitionTo returns true if the question can move from its current status
// to the given target status.
func (q *Question) CanTransitionTo(status QuestionStatus) bool {
	allowed, ok := validTransitions[q.Status]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == status {
			return true
		}
	}
	return false
}
