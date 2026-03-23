package entities

type StaffID string

type Staff struct {
	ID          StaffID
	SlackUserID *SlackUserID
	Name        string
	Email       string
}
