package entities

type UserRole string

const (
	UserRoleParticipant UserRole = "participant"
	UserRoleMentor      UserRole = "mentor"
)

type UserID string

type SlackUserID string

type User struct {
	ID          UserID
	SlackUserID SlackUserID
	Name        string
	Email       string
	Role        UserRole
}
