package entities

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
