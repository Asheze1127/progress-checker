package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/application/service/jwt"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

const testJWTSecret = "test-secret-key-for-middleware-32b"
const testInternalToken = "test-internal-token-value-32bytes!"

func init() {
	gin.SetMode(gin.TestMode)
}

func generateTestToken(t *testing.T, user *entities.User) string {
	t.Helper()
	jwtSvc, err := jwt.NewJWTService(testJWTSecret)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}
	token, err := jwtSvc.GenerateToken(user)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}
	return token
}

// callSecurityMiddleware creates a Gin context with the given scopes set,
// runs SecurityMiddleware, and returns the recorder and context.
func callSecurityMiddleware(
	t *testing.T,
	req *http.Request,
	setScopes func(c *gin.Context),
) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	jwtSvc, err := jwt.NewJWTService(testJWTSecret)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}
	mw := SecurityMiddleware(jwtSvc, testInternalToken)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	if setScopes != nil {
		setScopes(c)
	}

	mw(c)
	return rec, c
}

func TestSecurityMiddleware_BearerAuth_ValidToken(t *testing.T) {
	mentorUser := &entities.User{
		ID:   "user-1",
		Name: "Test Mentor",
		Role: entities.UserRoleMentor,
	}
	token := generateTestToken(t, mentorUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec, c := callSecurityMiddleware(t, req, func(c *gin.Context) {
		c.Set(openapi.BearerAuthScopes, []string{})
	})

	if rec.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
	}
	if c.IsAborted() {
		t.Error("expected request to not be aborted")
	}

	user := UserFromContext(c)
	if user == nil {
		t.Fatal("expected user in context, got nil")
	}
	if user.ID != "user-1" {
		t.Errorf("user.ID = %q, want %q", user.ID, "user-1")
	}
}

func TestSecurityMiddleware_BearerAuth_MissingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)

	rec, c := callSecurityMiddleware(t, req, func(c *gin.Context) {
		c.Set(openapi.BearerAuthScopes, []string{})
	})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if !c.IsAborted() {
		t.Error("expected request to be aborted")
	}
}

func TestSecurityMiddleware_BearerAuth_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	rec, c := callSecurityMiddleware(t, req, func(c *gin.Context) {
		c.Set(openapi.BearerAuthScopes, []string{})
	})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if !c.IsAborted() {
		t.Error("expected request to be aborted")
	}
}

func TestSecurityMiddleware_BearerAuth_NonMentorRole(t *testing.T) {
	participantUser := &entities.User{
		ID:   "user-2",
		Name: "Participant",
		Role: entities.UserRoleParticipant,
	}
	token := generateTestToken(t, participantUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec, c := callSecurityMiddleware(t, req, func(c *gin.Context) {
		c.Set(openapi.BearerAuthScopes, []string{})
	})

	if rec.Code != http.StatusForbidden {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if !c.IsAborted() {
		t.Error("expected request to be aborted")
	}
}

func TestSecurityMiddleware_InternalToken_Valid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/internal/issues", nil)
	req.Header.Set("X-Internal-Token", testInternalToken)

	rec, c := callSecurityMiddleware(t, req, func(c *gin.Context) {
		c.Set(openapi.InternalTokenAuthScopes, []string{})
	})

	if rec.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
	}
	if c.IsAborted() {
		t.Error("expected request to not be aborted")
	}
}

func TestSecurityMiddleware_InternalToken_Invalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/internal/issues", nil)
	req.Header.Set("X-Internal-Token", "wrong-token")

	rec, c := callSecurityMiddleware(t, req, func(c *gin.Context) {
		c.Set(openapi.InternalTokenAuthScopes, []string{})
	})

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if !c.IsAborted() {
		t.Error("expected request to be aborted")
	}
}

func TestSecurityMiddleware_NoScopes_PassThrough(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)

	rec, c := callSecurityMiddleware(t, req, nil)

	if rec.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
	}
	if c.IsAborted() {
		t.Error("expected request to not be aborted (public endpoint)")
	}
}

func generateTestStaffToken(t *testing.T, staff *entities.Staff) string {
	t.Helper()
	jwtSvc, err := jwt.NewJWTService(testJWTSecret)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}
	token, err := jwtSvc.GenerateStaffToken(staff)
	if err != nil {
		t.Fatalf("failed to generate test staff token: %v", err)
	}
	return token
}

func TestSecurityMiddleware_StaffBearerAuth_ValidToken(t *testing.T) {
	staff := &entities.Staff{
		ID:   "staff-1",
		Name: "Test Staff",
	}
	token := generateTestStaffToken(t, staff)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/staff/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec, c := callSecurityMiddleware(t, req, func(c *gin.Context) {
		c.Set(openapi.StaffBearerAuthScopes, []string{})
	})

	if rec.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
	}
	if c.IsAborted() {
		t.Error("expected request to not be aborted")
	}

	s := StaffFromContext(c)
	if s == nil {
		t.Fatal("expected staff in context, got nil")
	}
	if string(s.ID) != "staff-1" {
		t.Errorf("staff.ID = %q, want %q", s.ID, "staff-1")
	}
}

func TestSecurityMiddleware_StaffBearerAuth_NonStaffToken(t *testing.T) {
	mentorUser := &entities.User{
		ID:   "user-1",
		Name: "Test Mentor",
		Role: entities.UserRoleMentor,
	}
	token := generateTestToken(t, mentorUser)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/staff/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec, c := callSecurityMiddleware(t, req, func(c *gin.Context) {
		c.Set(openapi.StaffBearerAuthScopes, []string{})
	})

	if rec.Code != http.StatusForbidden {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if !c.IsAborted() {
		t.Error("expected request to be aborted")
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
			got := extractBearerToken(tt.header)
			if got != tt.wantToken {
				t.Errorf("extractBearerToken() = %q, want %q", got, tt.wantToken)
			}
		})
	}
}
