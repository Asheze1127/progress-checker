package entities

import "time"

// MentorAssignment represents which mentor is responsible for which team.
// Prepared for future scope-based access control.
type MentorAssignment struct {
	ID        string    `json:"id" db:"id"`
	MentorID  string    `json:"mentor_id" db:"mentor_id"`
	TeamID    string    `json:"team_id" db:"team_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
