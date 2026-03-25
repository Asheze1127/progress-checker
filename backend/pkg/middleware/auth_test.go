package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// mockSessionValidator implements SessionValidator for testing.
type mockSessionValidator struct {
	validateFunc func(ctx context.Context, token string) (*entities.User, error)
}

func (m *mockSessionValidator) ValidateSession(ctx context.Context, token string) (*entities.User, error) {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, token)
	}
	return nil, errors.New("session not found or expired")
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	validator := &mockSessionValidator{
		validateFunc: func(_ context.Context, token string) (*entities.User, error) {
			if token != "valid-token" {
				return nil, errors.New("session not found or expired")
			}
			return &entities.User{
				ID:   "user-1",
				Name: "Test Mentor",
				Role: entities.UserRoleMentor,
			}, nil
		},
	}

	var capturedUser *entities.User
	handler := AuthMiddleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
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
	validator := &mockSessionValidator{}

	handler := AuthMiddleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	validator := &mockSessionValidator{}

	handler := AuthMiddleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	validator := &mockSessionValidator{}

	handler := AuthMiddleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	validator := &mockSessionValidator{
		validateFunc: func(_ context.Context, _ string) (*entities.User, error) {
			return nil, errors.New("session not found or expired")
		},
	}

	handler := AuthMiddleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	validator := &mockSessionValidator{
		validateFunc: func(_ context.Context, _ string) (*entities.User, error) {
			return nil, errors.New("user is not a mentor")
		},
	}

	handler := AuthMiddleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer some-token")
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
