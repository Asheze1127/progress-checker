package service

import (
	"github.com/samber/do/v2"
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost is the cost factor for bcrypt hashing.
// OWASP recommends 12+ for improved resistance to brute-force attacks.
const bcryptCost = 12

// PasswordHasher handles password hashing and verification using bcrypt.
type PasswordHasher struct{}

// NewPasswordHasher creates a new PasswordHasher via DI container.
func NewPasswordHasher(_ do.Injector) (*PasswordHasher, error) {
	return &PasswordHasher{}, nil
}

// Hash creates a bcrypt hash of the given password.
func (h *PasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Verify compares a bcrypt hashed password with a plain text password.
// Returns nil on success, or an error if they don't match.
func (h *PasswordHasher) Verify(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}
