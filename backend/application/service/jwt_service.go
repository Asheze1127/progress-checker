package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/samber/do/v2"

	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/Asheze1127/progress-checker/backend/util"
	"github.com/golang-jwt/jwt/v5"
)

// TokenExpiry is the duration after which a JWT token expires.
const TokenExpiry = 24 * time.Hour

// ErrInvalidToken is returned when a JWT token is invalid or expired.
var ErrInvalidToken = errors.New("invalid or expired token")

// TokenClaims holds the custom claims extracted from a JWT token.
type TokenClaims struct {
	UserID   entities.UserID
	UserName string
	UserRole entities.UserRole
}

// JWTService handles JWT token generation and validation.
type JWTService struct {
	secret []byte
	now    func() time.Time
}

// NewJWTService creates a new JWTService via DI container.
// It reads the JWT secret from the Config registered in the injector.
func NewJWTService(i do.Injector) (*JWTService, error) {
	cfg := do.MustInvoke[*util.Config](i)
	return newJWTService(cfg.JWTSecret), nil
}

// newJWTService creates a JWTService with the given secret string.
func newJWTService(secret string) *JWTService {
	return &JWTService{
		secret: []byte(secret),
		now:    time.Now,
	}
}

// GenerateToken creates a new JWT token for the given user.
func (s *JWTService) GenerateToken(user *entities.User) (string, error) {
	now := s.now()
	claims := jwt.MapClaims{
		"sub":  string(user.ID),
		"name": user.Name,
		"role": string(user.Role),
		"iat":  now.Unix(),
		"exp":  now.Add(TokenExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken parses and validates a JWT token string, returning the claims.
func (s *JWTService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	}, jwt.WithTimeFunc(s.now))
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	sub, _ := claims["sub"].(string)
	name, _ := claims["name"].(string)
	role, _ := claims["role"].(string)

	if sub == "" {
		return nil, ErrInvalidToken
	}

	userRole := entities.UserRole(role)
	if !isValidUserRole(userRole) {
		return nil, ErrInvalidToken
	}

	return &TokenClaims{
		UserID:   entities.UserID(sub),
		UserName: name,
		UserRole: userRole,
	}, nil
}

// isValidUserRole checks if the role is a known user role.
func isValidUserRole(role entities.UserRole) bool {
	switch role {
	case entities.UserRoleMentor, entities.UserRoleParticipant:
		return true
	default:
		return false
	}
}
