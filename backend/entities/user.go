package entities

type UserRole string

const (
	UserRoleParticipant UserRole = "participant"
	UserRoleMentor      UserRole = "mentor"
)

// User represents a Slack workspace member (participant or mentor).
type User struct {
	ID          string
	SlackUserID string
	Name        string
	Email       string
	Role        UserRole
}
