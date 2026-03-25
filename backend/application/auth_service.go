package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/Asheze1127/progress-checker/backend/entities"
	"golang.org/x/crypto/bcrypt"
)

// SessionTokenLength is the number of random bytes used to generate a session token.
const SessionTokenLength = 32

// SessionExpiry is the duration after which a session expires.
const SessionExpiry = 24 * time.Hour

// ErrInvalidCredentials is returned when email or password is incorrect.
var ErrInvalidCredentials = errors.New("invalid email or password")

// ErrUserNotMentor is returned when the authenticated user does not have the mentor role.
var ErrUserNotMentor = errors.New("user is not a mentor")

// ErrSessionNotFound is returned when a session token is not found or has expired.
var ErrSessionNotFound = errors.New("session not found or expired")

// Session represents an authenticated session stored in the database.
type Session struct {
	TokenHash string
	UserID    entities.UserID
	ExpiresAt time.Time
	CreatedAt time.Time
}

// UserWithPassword extends User with a hashed password for credential verification.
type UserWithPassword struct {
	entities.User
	PasswordHash string
}

// SessionRepository defines the interface for session persistence.
type SessionRepository interface {
	// Store saves a new session to the database.
	Store(ctx context.Context, session Session) error
	// FindByTokenHash retrieves a session by its token hash.
	FindByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	// DeleteByTokenHash removes a session by its token hash.
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
}

// UserRepository defines the interface for user lookup.
type UserRepository interface {
	// FindByEmail retrieves a user with their password hash by email address.
	FindByEmail(ctx context.Context, email string) (*UserWithPassword, error)
	// FindByID retrieves a user by their ID.
	FindByID(ctx context.Context, id entities.UserID) (*entities.User, error)
}

// AuthService handles authentication-related use cases.
type AuthService struct {
	sessionRepo SessionRepository
	userRepo    UserRepository
	now         func() time.Time
	generateToken func() (string, error)
}

// NewAuthService creates a new AuthService with the given repositories.
func NewAuthService(sessionRepo SessionRepository, userRepo UserRepository) *AuthService {
	return &AuthService{
		sessionRepo:   sessionRepo,
		userRepo:      userRepo,
		now:           time.Now,
		generateToken: generateSecureToken,
	}
}

// LoginResult holds the result of a successful login.
type LoginResult struct {
	Token string
	User  entities.User
}

// Login authenticates a user by email and password, verifies mentor role,
// and creates a new session.
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	userWithPw, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userWithPw.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if userWithPw.Role != entities.UserRoleMentor {
		return nil, ErrUserNotMentor
	}

	token, err := s.generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	tokenHash := HashToken(token)
	now := s.now()

	session := Session{
		TokenHash: tokenHash,
		UserID:    userWithPw.ID,
		ExpiresAt: now.Add(SessionExpiry),
		CreatedAt: now,
	}

	if err := s.sessionRepo.Store(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return &LoginResult{
		Token: token,
		User:  userWithPw.User,
	}, nil
}

// Logout invalidates a session by removing it from the database.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}

	tokenHash := HashToken(token)
	return s.sessionRepo.DeleteByTokenHash(ctx, tokenHash)
}

// ValidateSession checks if a token corresponds to a valid, non-expired session
// and returns the associated user.
func (s *AuthService) ValidateSession(ctx context.Context, token string) (*entities.User, error) {
	if token == "" {
		return nil, ErrSessionNotFound
	}

	tokenHash := HashToken(token)

	session, err := s.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if s.now().After(session.ExpiresAt) {
		return nil, ErrSessionNotFound
	}

	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if user.Role != entities.UserRoleMentor {
		return nil, ErrUserNotMentor
	}

	return user, nil
}

// HashToken creates a SHA-256 hash of a token and returns it as a hex string.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// generateSecureToken creates a cryptographically random token encoded as base64.
func generateSecureToken() (string, error) {
	bytes := make([]byte, SessionTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
