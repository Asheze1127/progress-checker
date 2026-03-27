package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/application/service/jwt"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

const testJWTSecret = "test-secret-key-for-middleware"

func generateTestToken(t *testing.T, user *entities.User) string {
	t.Helper()
	jwtSvc := jwt.NewJWTService(testJWTSecret)
	token, err := jwtSvc.GenerateToken(user)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}
	return token
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	jwtSvc := jwt.NewJWTService(testJWTSecret)

	mentorUser := &entities.User{
		ID:   "user-1",
		Name: "Test Mentor",
		Role: entities.UserRoleMentor,
	}
	token := generateTestToken(t, mentorUser)

	var capturedUser *entities.User
	handler := AuthMiddleware(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	if capturedUser == nil {
		t.Fatal("expected user in context, got nil")
	}

	if capturedUser.ID != "user-1" {
		t.Errorf("user.ID = %q, want %q", capturedUser.ID, "user-1")
	}
}

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	jwtSvc := jwt.NewJWTService(testJWTSecret)

	handler := AuthMiddleware(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_InvalidBearerFormat(t *testing.T) {
	jwtSvc := jwt.NewJWTService(testJWTSecret)

	handler := AuthMiddleware(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_EmptyBearerToken(t *testing.T) {
	jwtSvc := jwt.NewJWTService(testJWTSecret)

	handler := AuthMiddleware(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtSvc := jwt.NewJWTService(testJWTSecret)

	handler := AuthMiddleware(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_NonMentorRole(t *testing.T) {
	jwtSvc := jwt.NewJWTService(testJWTSecret)

	participantUser := &entities.User{
		ID:   "user-2",
		Name: "Participant",
		Role: entities.UserRoleParticipant,
	}
	token := generateTestToken(t, participantUser)

	handler := AuthMiddleware(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestUserFromContext_NoUser(t *testing.T) {
	ctx := context.Background()
	user := UserFromContext(ctx)

	if user != nil {
		t.Errorf("expected nil user, got %+v", user)
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
	}{
		{
			name:      "valid bearer token",
			header:    "Bearer my-token",
			wantToken: "my-token",
		},
		{
			name:      "empty header",
			header:    "",
			wantToken: "",
		},
		{
			name:      "basic auth",
			header:    "Basic abc123",
			wantToken: "",
		},
		{
			name:      "bearer with no token",
			header:    "Bearer ",
			wantToken: "",
		},
		{
			name:      "bearer with extra spaces",
			header:    "Bearer   my-token  ",
			wantToken: "my-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			got := extractBearerToken(req)
			if got != tt.wantToken {
				t.Errorf("extractBearerToken() = %q, want %q", got, tt.wantToken)
			}
		})
	}
}
