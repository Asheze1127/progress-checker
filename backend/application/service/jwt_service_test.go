package service

import (
	"testing"
	"time"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

func TestJWTService_GenerateAndValidateToken(t *testing.T) {
	svc := newJWTService("test-secret")

	user := &entities.User{
		ID:   "user-1",
		Name: "Test Mentor",
		Role: entities.UserRoleMentor,
	}

	token, err := svc.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-1")
	}

	if claims.UserName != "Test Mentor" {
		t.Errorf("UserName = %q, want %q", claims.UserName, "Test Mentor")
	}

	if claims.UserRole != entities.UserRoleMentor {
		t.Errorf("UserRole = %q, want %q", claims.UserRole, entities.UserRoleMentor)
	}
}

func TestJWTService_ValidateToken_InvalidToken(t *testing.T) {
	svc := newJWTService("test-secret")

	_, err := svc.ValidateToken("not-a-valid-token")
	if err == nil {
		t.Fatal("ValidateToken() expected error for invalid token, got nil")
	}
}

func TestJWTService_ValidateToken_WrongSecret(t *testing.T) {
	svc1 := newJWTService("secret-one")
	svc2 := newJWTService("secret-two")

	user := &entities.User{
		ID:   "user-1",
		Name: "Test",
		Role: entities.UserRoleMentor,
	}

	token, err := svc1.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	_, err = svc2.ValidateToken(token)
	if err == nil {
		t.Fatal("ValidateToken() expected error for wrong secret, got nil")
	}
}

func TestJWTService_ValidateToken_ExpiredToken(t *testing.T) {
	svc := newJWTService("test-secret")
	// Override now to generate an already-expired token
	pastTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return pastTime }

	user := &entities.User{
		ID:   "user-1",
		Name: "Test",
		Role: entities.UserRoleMentor,
	}

	token, err := svc.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Reset to current time for validation
	svc.now = time.Now

	_, err = svc.ValidateToken(token)
	if err == nil {
		t.Fatal("ValidateToken() expected error for expired token, got nil")
	}
}

func TestJWTService_ValidateToken_EmptyString(t *testing.T) {
	svc := newJWTService("test-secret")

	_, err := svc.ValidateToken("")
	if err == nil {
		t.Fatal("ValidateToken() expected error for empty token, got nil")
	}
}
