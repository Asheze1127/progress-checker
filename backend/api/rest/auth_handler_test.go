package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Asheze1127/progress-checker/backend/application/service"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// testUserRepository implements entities.UserRepository for testing.
type testUserRepository struct {
	users map[string]*entities.UserWithPassword
}

func newTestUserRepository() *testUserRepository {
	return &testUserRepository{users: make(map[string]*entities.UserWithPassword)}
}

func (r *testUserRepository) FindByEmail(_ context.Context, email string) (*entities.UserWithPassword, error) {
	user, ok := r.users[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *testUserRepository) FindByID(_ context.Context, id entities.UserID) (*entities.User, error) {
	for _, u := range r.users {
		if u.User.ID == id {
			return &u.User, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *testUserRepository) addMentor(email, password, name string) {
	hasher := service.NewPasswordHasher()
	hash, _ := hasher.Hash(password)
	r.users[email] = &entities.UserWithPassword{
		User: entities.User{
			ID:    entities.UserID("user-1"),
			Name:  name,
			Email: email,
			Role:  entities.UserRoleMentor,
		},
		PasswordHash: hash,
	}
}

func (r *testUserRepository) addParticipant(email, password, name string) {
	hasher := service.NewPasswordHasher()
	hash, _ := hasher.Hash(password)
	r.users[email] = &entities.UserWithPassword{
		User: entities.User{
			ID:    entities.UserID("user-2"),
			Name:  name,
			Email: email,
			Role:  entities.UserRoleParticipant,
		},
		PasswordHash: hash,
	}
}

func setupAuthHandler() (*AuthHandler, *testUserRepository) {
	userRepo := newTestUserRepository()
	jwtSvc := service.NewJWTService("test-secret-key")
	hasher := service.NewPasswordHasher()
	loginUC := usecase.NewLoginUseCase(userRepo, jwtSvc, hasher)
	handler := NewAuthHandler(loginUC)
	return handler, userRepo
}

func TestHandleLogin_Success(t *testing.T) {
	handler, userRepo := setupAuthHandler()
	userRepo.addMentor("mentor@example.com", "password123", "Test Mentor")

	body := `{"email":"mentor@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.HandleLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	respBody := rec.Body.String()
	if !strings.Contains(respBody, `"token"`) {
		t.Errorf("response should contain token, got %s", respBody)
	}
	if !strings.Contains(respBody, `"user"`) {
		t.Errorf("response should contain user, got %s", respBody)
	}
}

func TestHandleLogin_InvalidCredentials(t *testing.T) {
	handler, userRepo := setupAuthHandler()
	userRepo.addMentor("mentor@example.com", "password123", "Test Mentor")

	body := `{"email":"mentor@example.com","password":"wrong-password"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.HandleLogin(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandleLogin_UserNotFound(t *testing.T) {
	handler, _ := setupAuthHandler()

	body := `{"email":"nobody@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.HandleLogin(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandleLogin_NotMentor(t *testing.T) {
	handler, userRepo := setupAuthHandler()
	userRepo.addParticipant("participant@example.com", "password123", "Participant")

	body := `{"email":"participant@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.HandleLogin(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestHandleLogin_EmptyBody(t *testing.T) {
	handler, _ := setupAuthHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.HandleLogin(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleLogin_MissingFields(t *testing.T) {
	handler, _ := setupAuthHandler()

	tests := []struct {
		name string
		body string
	}{
		{name: "missing email", body: `{"password":"password123"}`},
		{name: "missing password", body: `{"email":"mentor@example.com"}`},
		{name: "empty email", body: `{"email":"","password":"password123"}`},
		{name: "empty password", body: `{"email":"mentor@example.com","password":""}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleLogin(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleLogin_WrongMethod(t *testing.T) {
	handler, _ := setupAuthHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	rec := httptest.NewRecorder()

	handler.HandleLogin(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
