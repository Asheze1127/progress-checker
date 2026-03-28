package entities

import (
	"errors"
	"fmt"
	"strings"
)

type StaffID string

type Staff struct {
	ID          StaffID
	SlackUserID *SlackUserID
	Name        string
	Email       string
}

func (s Staff) Validate() error {
	var errs []error

	if strings.TrimSpace(string(s.ID)) == "" {
		errs = append(errs, fmt.Errorf("staff.id is required"))
	}

	if strings.TrimSpace(s.Name) == "" {
		errs = append(errs, fmt.Errorf("staff.name is required"))
	}

	if err := validateEmail("staff.email", s.Email); err != nil {
		errs = append(errs, err)
	}

	if s.SlackUserID != nil && strings.TrimSpace(string(*s.SlackUserID)) == "" {
		errs = append(errs, fmt.Errorf("staff.slack_user_id is required"))
	}

	return errors.Join(errs...)
}

// StaffWithPassword extends Staff with a hashed password for credential verification.
type StaffWithPassword struct {
	Staff
	PasswordHash string
}
