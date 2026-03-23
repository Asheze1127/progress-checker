package entities

import (
	"fmt"
	"net/mail"
	"strings"
)

func validateEmail(field, value string) error {
	email := strings.TrimSpace(value)
	if email == "" {
		return fmt.Errorf("%s is required", field)
	}

	parsedAddress, err := mail.ParseAddress(email)
	if err != nil || parsedAddress.Address != email {
		return fmt.Errorf("%s must be a valid email address", field)
	}

	return nil
}
