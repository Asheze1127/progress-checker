package usecase

import (
  "context"
  "errors"
  "fmt"

  "github.com/Asheze1127/progress-checker/backend/application/service/jwt"
  "github.com/Asheze1127/progress-checker/backend/application/service/password_hasher"
  "github.com/Asheze1127/progress-checker/backend/entities"
)

// ErrInvalidCredentials is returned when email or password is incorrect.
var ErrInvalidCredentials = errors.New("invalid email or password")

// ErrUserNotMentor is returned when the authenticated user does not have the mentor role.
var ErrUserNotMentor = errors.New("user is not a mentor")

// LoginResult holds the result of a successful login.
type LoginResult struct {
  Token string
  User  entities.User
}

// LoginUseCase handles the login workflow.
type LoginUseCase struct {
  userRepo entities.UserRepository
  jwt      *jwt.JWTService
  hasher   *passwordhasher.PasswordHasher
}

// NewLoginUseCase creates a new LoginUseCase with the given dependencies.
func NewLoginUseCase(
  userRepo entities.UserRepository,
  jwt *jwt.JWTService,
  hasher *passwordhasher.PasswordHasher,
) *LoginUseCase {
  return &LoginUseCase{
    userRepo: userRepo,
    jwt:      jwt,
    hasher:   hasher,
  }
}

// Execute authenticates a user by email and password, verifies mentor role,
// and generates a JWT token.
func (uc *LoginUseCase) Execute(ctx context.Context, email, password string) (*LoginResult, error) {
  if email == "" {
    return nil, fmt.Errorf("email is required")
  }
  if password == "" {
    return nil, fmt.Errorf("password is required")
  }

  userWithPw, err := uc.userRepo.FindByEmail(ctx, email)
  if err != nil {
    return nil, ErrInvalidCredentials
  }

  if err := uc.hasher.Verify(userWithPw.PasswordHash, password); err != nil {
    return nil, ErrInvalidCredentials
  }

  if !userWithPw.IsMentor() {
    return nil, ErrUserNotMentor
  }

  token, err := uc.jwt.GenerateToken(&userWithPw.User)
  if err != nil {
    return nil, fmt.Errorf("failed to generate token: %w", err)
  }

  return &LoginResult{
    Token: token,
    User:  userWithPw.User,
  }, nil
}
