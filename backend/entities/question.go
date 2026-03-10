package entities

import "time"

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
	SessionStatusResolved     QuestionSessionStatus = "resolved"
)

// Question represents a question submitted via /question.
type Question struct {
	ID             string         `json:"id" db:"id"`
	TeamID         string         `json:"team_id" db:"team_id"`
	UserID         string         `json:"user_id" db:"user_id"`
	Title          string         `json:"title" db:"title"`
	Body           string         `json:"body" db:"body"`
	Status         QuestionStatus `json:"status" db:"status"`
	SlackThreadTS  string         `json:"slack_thread_ts" db:"slack_thread_ts"`
	SlackChannelID string         `json:"slack_channel_id" db:"slack_channel_id"`
	AssignedTo     *string        `json:"assigned_to" db:"assigned_to"`
	GithubIssueURL *string       `json:"github_issue_url" db:"github_issue_url"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

// QuestionSession tracks the AI conversation state for a question thread.
// Used to determine whether a Slack reply is a new question or a follow-up.
type QuestionSession struct {
	ID           string                `json:"id" db:"id"`
	QuestionID   string                `json:"question_id" db:"question_id"`
	Status       QuestionSessionStatus `json:"status" db:"status"`
	// 3 rounds of AI follow-up before escalating to a mentor
	MaxFollowUps int                   `json:"max_follow_ups" db:"max_follow_ups"`
	CurrentRound int                   `json:"current_round" db:"current_round"`
	CreatedAt    time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at" db:"updated_at"`
}
