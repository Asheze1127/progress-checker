package entities

import (
	"errors"
	"fmt"
	"strings"
)

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

func (u User) Validate() error {
	var errs []error

	if strings.TrimSpace(string(u.ID)) == "" {
		errs = append(errs, fmt.Errorf("user.id is required"))
	}

	if strings.TrimSpace(string(u.SlackUserID)) == "" {
		errs = append(errs, fmt.Errorf("user.slack_user_id is required"))
	}

	if strings.TrimSpace(u.Name) == "" {
		errs = append(errs, fmt.Errorf("user.name is required"))
	}

	if err := validateEmail("user.email", u.Email); err != nil {
		errs = append(errs, err)
	}

	switch u.Role {
	case UserRoleParticipant, UserRoleMentor:
	default:
		errs = append(errs, fmt.Errorf("user.role must be one of participant, mentor"))
	}

	return errors.Join(errs...)
}
