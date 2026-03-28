package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/Asheze1127/progress-checker/backend/entities"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

const TokenExpiry = 2 * time.Hour

var (
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrWeakSecret    = errors.New("JWT secret must be at least 32 bytes")
)

type TokenClaims struct {
	UserID   entities.UserID
	UserName string
	UserRole entities.UserRole
	// RawRole holds the role string from the JWT claim (e.g. "staff").
	// Use this when the role may not be a valid entities.UserRole.
	RawRole string
}

type JWTService struct {
	secret []byte
	now    func() time.Time
}

func NewJWTService(secret string) (*JWTService, error) {
	if len(secret) < 32 {
		return nil, ErrWeakSecret
	}
	return &JWTService{secret: []byte(secret), now: time.Now}, nil
}

// RoleStaff is the role claim value for staff tokens.
const RoleStaff = "staff"

func (s *JWTService) GenerateToken(user *entities.User) (string, error) {
	return s.generateTokenWithClaims(string(user.ID), user.Name, string(user.Role))
}

// GenerateStaffToken creates a JWT for a staff member.
func (s *JWTService) GenerateStaffToken(staff *entities.Staff) (string, error) {
	return s.generateTokenWithClaims(string(staff.ID), staff.Name, RoleStaff)
}

func (s *JWTService) generateTokenWithClaims(sub, name, role string) (string, error) {
	now := s.now()
	claims := jwtlib.MapClaims{
		"sub": sub, "name": name, "role": role,
		"iat": now.Unix(), "exp": now.Add(TokenExpiry).Unix(),
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signedToken, nil
}

func (s *JWTService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwtlib.Parse(tokenString, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	}, jwtlib.WithTimeFunc(s.now))
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(jwtlib.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return nil, ErrInvalidToken
	}
	name, ok := claims["name"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}
	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return nil, ErrInvalidToken
	}
	userRole := entities.UserRole(role)
	if userRole != entities.UserRoleMentor && userRole != entities.UserRoleParticipant && role != RoleStaff {
		return nil, ErrInvalidToken
	}
	return &TokenClaims{UserID: entities.UserID(sub), UserName: name, UserRole: userRole, RawRole: role}, nil
}
