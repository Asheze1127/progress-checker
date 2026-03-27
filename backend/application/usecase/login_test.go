package usecase

import (
  "context"
  "errors"
  "testing"

  "github.com/Asheze1127/progress-checker/backend/application/service/jwt"
  "github.com/Asheze1127/progress-checker/backend/application/service/password_hasher"
  "github.com/Asheze1127/progress-checker/backend/entities"
)

// mockUserRepository implements entities.UserRepository for testing.
type mockUserRepository struct {
  findByEmailFunc func(ctx context.Context, email string) (*entities.UserWithPassword, error)
  findByIDFunc    func(ctx context.Context, id entities.UserID) (*entities.User, error)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.UserWithPassword, error) {
  if m.findByEmailFunc != nil {
    return m.findByEmailFunc(ctx, email)
  }
  return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetByID(ctx context.Context, id entities.UserID) (*entities.User, error) {
  if m.findByIDFunc != nil {
    return m.findByIDFunc(ctx, id)
  }
  return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetByEmail(_ context.Context, _ string) (*entities.User, error) {
  return nil, errors.New("not implemented")
}

func (m *mockUserRepository) GetBySlackUserID(_ context.Context, _ entities.SlackUserID) (*entities.User, error) {
  return nil, errors.New("not implemented")
}

func hashPassword(t *testing.T, password string) string {
  t.Helper()
  hasher := passwordhasher.NewPasswordHasher()
  hash, err := hasher.Hash(password)
  if err != nil {
    t.Fatalf("failed to hash password: %v", err)
  }
  return hash
}

func newTestLoginUseCase(userRepo entities.UserRepository) *LoginUseCase {
  jwtSvc := jwt.NewJWTService("test-secret-key-for-testing-only")
  hasher := passwordhasher.NewPasswordHasher()
  return NewLoginUseCase(userRepo, jwtSvc, hasher)
}

func TestLoginUseCase_Execute_Success(t *testing.T) {
  passwordHash := hashPassword(t, "correct-password")

  userRepo := &mockUserRepository{
    findByEmailFunc: func(_ context.Context, email string) (*entities.UserWithPassword, error) {
      if email != "mentor@example.com" {
        return nil, errors.New("user not found")
      }
      return &entities.UserWithPassword{
        User: entities.User{
          ID:    "user-1",
          Name:  "Test Mentor",
          Email: "mentor@example.com",
          Role:  entities.UserRoleMentor,
        },
        PasswordHash: passwordHash,
      }, nil
    },
  }

  uc := newTestLoginUseCase(userRepo)
  result, err := uc.Execute(context.Background(), "mentor@example.com", "correct-password")
  if err != nil {
    t.Fatalf("Execute() error = %v", err)
  }

  if result.Token == "" {
    t.Error("expected non-empty token")
  }

  if result.User.ID != "user-1" {
    t.Errorf("User.ID = %q, want %q", result.User.ID, "user-1")
  }

  if result.User.Role != entities.UserRoleMentor {
    t.Errorf("User.Role = %q, want %q", result.User.Role, entities.UserRoleMentor)
  }
}

func TestLoginUseCase_Execute_InvalidCredentials(t *testing.T) {
  passwordHash := hashPassword(t, "correct-password")

  userRepo := &mockUserRepository{
    findByEmailFunc: func(_ context.Context, _ string) (*entities.UserWithPassword, error) {
      return &entities.UserWithPassword{
        User: entities.User{
          ID:   "user-1",
          Role: entities.UserRoleMentor,
        },
        PasswordHash: passwordHash,
      }, nil
    },
  }

  uc := newTestLoginUseCase(userRepo)
  _, err := uc.Execute(context.Background(), "mentor@example.com", "wrong-password")
  if !errors.Is(err, ErrInvalidCredentials) {
    t.Errorf("Execute() error = %v, want %v", err, ErrInvalidCredentials)
  }
}

func TestLoginUseCase_Execute_UserNotFound(t *testing.T) {
  userRepo := &mockUserRepository{
    findByEmailFunc: func(_ context.Context, _ string) (*entities.UserWithPassword, error) {
      return nil, errors.New("user not found")
    },
  }

  uc := newTestLoginUseCase(userRepo)
  _, err := uc.Execute(context.Background(), "nobody@example.com", "password")
  if !errors.Is(err, ErrInvalidCredentials) {
    t.Errorf("Execute() error = %v, want %v", err, ErrInvalidCredentials)
  }
}

func TestLoginUseCase_Execute_NotMentor(t *testing.T) {
  passwordHash := hashPassword(t, "correct-password")

  userRepo := &mockUserRepository{
    findByEmailFunc: func(_ context.Context, _ string) (*entities.UserWithPassword, error) {
      return &entities.UserWithPassword{
        User: entities.User{
          ID:   "user-1",
          Role: entities.UserRoleParticipant,
        },
        PasswordHash: passwordHash,
      }, nil
    },
  }

  uc := newTestLoginUseCase(userRepo)
  _, err := uc.Execute(context.Background(), "participant@example.com", "correct-password")
  if !errors.Is(err, ErrUserNotMentor) {
    t.Errorf("Execute() error = %v, want %v", err, ErrUserNotMentor)
  }
}

func TestLoginUseCase_Execute_EmptyEmail(t *testing.T) {
  uc := newTestLoginUseCase(&mockUserRepository{})
  _, err := uc.Execute(context.Background(), "", "password")
  if err == nil {
    t.Fatal("Execute() expected error for empty email, got nil")
  }
}

func TestLoginUseCase_Execute_EmptyPassword(t *testing.T) {
  uc := newTestLoginUseCase(&mockUserRepository{})
  _, err := uc.Execute(context.Background(), "mentor@example.com", "")
  if err == nil {
    t.Fatal("Execute() expected error for empty password, got nil")
  }
}
