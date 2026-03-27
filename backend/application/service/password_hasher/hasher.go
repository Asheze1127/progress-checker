package passwordhasher

import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12

type PasswordHasher struct{}

func NewPasswordHasher() *PasswordHasher { return &PasswordHasher{} }

func (h *PasswordHasher) Hash(password string) (string, error) {
  hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
  if err != nil {
    return "", err
  }
  return string(hash), nil
}

func (h *PasswordHasher) Verify(hashed, plain string) error {
  return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}
