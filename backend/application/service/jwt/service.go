package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/Asheze1127/progress-checker/backend/entities"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

const TokenExpiry = 24 * time.Hour

var ErrInvalidToken = errors.New("invalid or expired token")

type TokenClaims struct {
	UserID   entities.UserID
	UserName string
	UserRole entities.UserRole
}

type JWTService struct {
	secret []byte
	now    func() time.Time
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret), now: time.Now}
}

func (s *JWTService) GenerateToken(user *entities.User) (string, error) {
	now := s.now()
	claims := jwtlib.MapClaims{
		"sub": string(user.ID), "name": user.Name, "role": string(user.Role),
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
	sub, _ := claims["sub"].(string)
	name, _ := claims["name"].(string)
	role, _ := claims["role"].(string)
	if sub == "" {
		return nil, ErrInvalidToken
	}
	userRole := entities.UserRole(role)
	if userRole != entities.UserRoleMentor && userRole != entities.UserRoleParticipant {
		return nil, ErrInvalidToken
	}
	return &TokenClaims{UserID: entities.UserID(sub), UserName: name, UserRole: userRole}, nil
}
