package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Asheze1127/progress-checker/backend/entities"
	"golang.org/x/crypto/bcrypt"
)

// mockSessionRepository implements SessionRepository for testing.
type mockSessionRepository struct {
	storeFunc            func(ctx context.Context, session Session) error
	findByTokenHashFunc  func(ctx context.Context, tokenHash string) (*Session, error)
	deleteByTokenHashFunc func(ctx context.Context, tokenHash string) error
}

func (m *mockSessionRepository) Store(ctx context.Context, session Session) error {
	if m.storeFunc != nil {
		return m.storeFunc(ctx, session)
	}
	return nil
}

func (m *mockSessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	if m.findByTokenHashFunc != nil {
		return m.findByTokenHashFunc(ctx, tokenHash)
	}
	return nil, ErrSessionNotFound
}

func (m *mockSessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	if m.deleteByTokenHashFunc != nil {
		return m.deleteByTokenHashFunc(ctx, tokenHash)
	}
	return nil
}

// mockUserRepository implements UserRepository for testing.
type mockUserRepository struct {
	findByEmailFunc func(ctx context.Context, email string) (*UserWithPassword, error)
	findByIDFunc    func(ctx context.Context, id entities.UserID) (*entities.User, error)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*UserWithPassword, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(ctx, email)
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) FindByID(ctx context.Context, id entities.UserID) (*entities.User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("user not found")
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hash)
}

func newTestAuthService(sessionRepo SessionRepository, userRepo UserRepository) *AuthService {
	svc := NewAuthService(sessionRepo, userRepo)
	svc.now = func() time.Time {
		return time.Date(2026, 3, 25, 12, 0, 0, 0, time.UTC)
	}
	svc.generateToken = func() (string, error) {
		return "test-token-abc123", nil
	}
	return svc
}

func TestAuthService_Login_Success(t *testing.T) {
	passwordHash := hashPassword(t, "correct-password")

	userRepo := &mockUserRepository{
		findByEmailFunc: func(_ context.Context, email string) (*UserWithPassword, error) {
			if email != "mentor@example.com" {
				return nil, errors.New("user not found")
			}
			return &UserWithPassword{
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

	var storedSession Session
	sessionRepo := &mockSessionRepository{
		storeFunc: func(_ context.Context, session Session) error {
			storedSession = session
			return nil
		},
	}

	svc := newTestAuthService(sessionRepo, userRepo)

	result, err := svc.Login(context.Background(), "mentor@example.com", "correct-password")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if result.Token != "test-token-abc123" {
		t.Errorf("Token = %q, want %q", result.Token, "test-token-abc123")
	}

	if result.User.ID != "user-1" {
		t.Errorf("User.ID = %q, want %q", result.User.ID, "user-1")
	}

	if result.User.Role != entities.UserRoleMentor {
		t.Errorf("User.Role = %q, want %q", result.User.Role, entities.UserRoleMentor)
	}

	expectedTokenHash := HashToken("test-token-abc123")
	if storedSession.TokenHash != expectedTokenHash {
		t.Errorf("stored session TokenHash = %q, want %q", storedSession.TokenHash, expectedTokenHash)
	}

	if storedSession.UserID != "user-1" {
		t.Errorf("stored session UserID = %q, want %q", storedSession.UserID, "user-1")
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	passwordHash := hashPassword(t, "correct-password")

	userRepo := &mockUserRepository{
		findByEmailFunc: func(_ context.Context, _ string) (*UserWithPassword, error) {
			return &UserWithPassword{
				User: entities.User{
					ID:   "user-1",
					Role: entities.UserRoleMentor,
				},
				PasswordHash: passwordHash,
			}, nil
		},
	}

	sessionRepo := &mockSessionRepository{}
	svc := newTestAuthService(sessionRepo, userRepo)

	_, err := svc.Login(context.Background(), "mentor@example.com", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepository{
		findByEmailFunc: func(_ context.Context, _ string) (*UserWithPassword, error) {
			return nil, errors.New("user not found")
		},
	}

	sessionRepo := &mockSessionRepository{}
	svc := newTestAuthService(sessionRepo, userRepo)

	_, err := svc.Login(context.Background(), "nobody@example.com", "password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestAuthService_Login_NotMentor(t *testing.T) {
	passwordHash := hashPassword(t, "correct-password")

	userRepo := &mockUserRepository{
		findByEmailFunc: func(_ context.Context, _ string) (*UserWithPassword, error) {
			return &UserWithPassword{
				User: entities.User{
					ID:   "user-1",
					Role: entities.UserRoleParticipant,
				},
				PasswordHash: passwordHash,
			}, nil
		},
	}

	sessionRepo := &mockSessionRepository{}
	svc := newTestAuthService(sessionRepo, userRepo)

	_, err := svc.Login(context.Background(), "participant@example.com", "correct-password")
	if !errors.Is(err, ErrUserNotMentor) {
		t.Errorf("Login() error = %v, want %v", err, ErrUserNotMentor)
	}
}

func TestAuthService_Login_EmptyEmail(t *testing.T) {
	svc := newTestAuthService(&mockSessionRepository{}, &mockUserRepository{})

	_, err := svc.Login(context.Background(), "", "password")
	if err == nil {
		t.Fatal("Login() expected error for empty email, got nil")
	}
}

func TestAuthService_Login_EmptyPassword(t *testing.T) {
	svc := newTestAuthService(&mockSessionRepository{}, &mockUserRepository{})

	_, err := svc.Login(context.Background(), "mentor@example.com", "")
	if err == nil {
		t.Fatal("Login() expected error for empty password, got nil")
	}
}

func TestAuthService_Logout_Success(t *testing.T) {
	var deletedHash string
	sessionRepo := &mockSessionRepository{
		deleteByTokenHashFunc: func(_ context.Context, tokenHash string) error {
			deletedHash = tokenHash
			return nil
		},
	}

	svc := newTestAuthService(sessionRepo, &mockUserRepository{})

	err := svc.Logout(context.Background(), "test-token")
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	expectedHash := HashToken("test-token")
	if deletedHash != expectedHash {
		t.Errorf("deleted hash = %q, want %q", deletedHash, expectedHash)
	}
}

func TestAuthService_Logout_EmptyToken(t *testing.T) {
	svc := newTestAuthService(&mockSessionRepository{}, &mockUserRepository{})

	err := svc.Logout(context.Background(), "")
	if err == nil {
		t.Fatal("Logout() expected error for empty token, got nil")
	}
}

func TestAuthService_ValidateSession_Success(t *testing.T) {
	token := "valid-token"
	tokenHash := HashToken(token)

	sessionRepo := &mockSessionRepository{
		findByTokenHashFunc: func(_ context.Context, hash string) (*Session, error) {
			if hash != tokenHash {
				return nil, ErrSessionNotFound
			}
			return &Session{
				TokenHash: tokenHash,
				UserID:    "user-1",
				ExpiresAt: time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
			}, nil
		},
	}

	userRepo := &mockUserRepository{
		findByIDFunc: func(_ context.Context, id entities.UserID) (*entities.User, error) {
			if id != "user-1" {
				return nil, errors.New("user not found")
			}
			return &entities.User{
				ID:   "user-1",
				Name: "Test Mentor",
				Role: entities.UserRoleMentor,
			}, nil
		},
	}

	svc := newTestAuthService(sessionRepo, userRepo)

	user, err := svc.ValidateSession(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateSession() error = %v", err)
	}

	if user.ID != "user-1" {
		t.Errorf("User.ID = %q, want %q", user.ID, "user-1")
	}
}

func TestAuthService_ValidateSession_Expired(t *testing.T) {
	token := "expired-token"
	tokenHash := HashToken(token)

	sessionRepo := &mockSessionRepository{
		findByTokenHashFunc: func(_ context.Context, hash string) (*Session, error) {
			if hash != tokenHash {
				return nil, ErrSessionNotFound
			}
			return &Session{
				TokenHash: tokenHash,
				UserID:    "user-1",
				ExpiresAt: time.Date(2026, 3, 24, 12, 0, 0, 0, time.UTC), // expired
			}, nil
		},
	}

	svc := newTestAuthService(sessionRepo, &mockUserRepository{})

	_, err := svc.ValidateSession(context.Background(), token)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("ValidateSession() error = %v, want %v", err, ErrSessionNotFound)
	}
}

func TestAuthService_ValidateSession_EmptyToken(t *testing.T) {
	svc := newTestAuthService(&mockSessionRepository{}, &mockUserRepository{})

	_, err := svc.ValidateSession(context.Background(), "")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("ValidateSession() error = %v, want %v", err, ErrSessionNotFound)
	}
}

func TestAuthService_ValidateSession_NotMentor(t *testing.T) {
	token := "valid-token"
	tokenHash := HashToken(token)

	sessionRepo := &mockSessionRepository{
		findByTokenHashFunc: func(_ context.Context, hash string) (*Session, error) {
			if hash != tokenHash {
				return nil, ErrSessionNotFound
			}
			return &Session{
				TokenHash: tokenHash,
				UserID:    "user-2",
				ExpiresAt: time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
			}, nil
		},
	}

	userRepo := &mockUserRepository{
		findByIDFunc: func(_ context.Context, id entities.UserID) (*entities.User, error) {
			return &entities.User{
				ID:   "user-2",
				Name: "Participant",
				Role: entities.UserRoleParticipant,
			}, nil
		},
	}

	svc := newTestAuthService(sessionRepo, userRepo)

	_, err := svc.ValidateSession(context.Background(), token)
	if !errors.Is(err, ErrUserNotMentor) {
		t.Errorf("ValidateSession() error = %v, want %v", err, ErrUserNotMentor)
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	hash1 := HashToken("same-token")
	hash2 := HashToken("same-token")

	if hash1 != hash2 {
		t.Errorf("HashToken not deterministic: %q != %q", hash1, hash2)
	}
}

func TestHashToken_DifferentForDifferentTokens(t *testing.T) {
	hash1 := HashToken("token-a")
	hash2 := HashToken("token-b")

	if hash1 == hash2 {
		t.Error("HashToken returned same hash for different tokens")
	}
}
