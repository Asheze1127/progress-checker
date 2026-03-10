package entities

import "time"

// MentorAssignment represents which mentor is responsible for which team.
// Prepared for future scope-based access control.
type MentorAssignment struct {
	ID           string    `json:"id" db:"id"`
	TeamID       string    `json:"team_id" db:"team_id"`
	MentorUserID string    `json:"mentor_user_id" db:"mentor_user_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
