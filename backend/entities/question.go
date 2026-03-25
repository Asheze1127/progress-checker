package entities

import "time"

// QuestionStatus represents the mentor-facing handling state of a question.
type QuestionStatus string

const (
	QuestionStatusOpen       QuestionStatus = "open"
	QuestionStatusInProgress QuestionStatus = "in_progress"
	QuestionStatusResolved   QuestionStatus = "resolved"
)

type QuestionSessionStatus string

const (
	SessionStatusAwaitingAI   QuestionSessionStatus = "awaiting_ai"
	SessionStatusAwaitingUser QuestionSessionStatus = "awaiting_user"
	SessionStatusEscalated    QuestionSessionStatus = "escalated"
)

// Question represents a question submitted via /question.
type Question struct {
	ID               string
	TeamID           string
	AskedByUserID    string
	AssignedMentorID *string
	Title            string
	// Body stores the initial question text submitted from Slack.
	// To avoid rate restrictions
	Body          string
	Status        QuestionStatus
	SlackThreadTS string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// QuestionSession tracks the AI conversation state for a question thread.
// Used to determine whether a Slack reply is a new question or a follow-up.
type QuestionSession struct {
	ID         string
	QuestionID string
	// Status is independent from QuestionStatus and only tracks active AI follow-up flow.
	Status QuestionSessionStatus
	// 3 rounds of AI follow-up before escalating to a mentor
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
