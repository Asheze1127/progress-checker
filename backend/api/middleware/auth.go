package middleware

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/application/appcontext"
	"github.com/Asheze1127/progress-checker/backend/application/service/jwt"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// GinKeyUser is the Gin context key for storing the authenticated user.
const GinKeyUser = appcontext.KeyUser

// GinKeyStaff is the Gin context key for storing the authenticated staff.
const GinKeyStaff = appcontext.KeyStaff

// SecurityMiddleware returns an oapi-codegen MiddlewareFunc that dispatches
// authentication based on OpenAPI security scopes set by the generated code.
//   - BearerAuthScopes  → JWT Bearer token validation (mentor role required)
//   - InternalTokenAuthScopes → X-Internal-Token header validation
//   - Neither → pass through (public endpoint)
func SecurityMiddleware(jwtService *jwt.JWTService, internalToken string) openapi.MiddlewareFunc {
	if len(internalToken) < 32 {
		slog.Error("SecurityMiddleware: internalToken must be at least 32 bytes")
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "server misconfiguration"})
		}
	}

	return func(c *gin.Context) {
		if _, exists := c.Get(openapi.BearerAuthScopes); exists {
			handleBearerAuth(c, jwtService)
			return
		}

		if _, exists := c.Get(openapi.StaffBearerAuthScopes); exists {
			handleStaffBearerAuth(c, jwtService)
			return
		}

		if _, exists := c.Get(openapi.InternalTokenAuthScopes); exists {
			handleInternalTokenAuth(c, internalToken)
			return
		}

		// No security scope set — route is public (e.g., /api/v1/auth/login).
		// The generated code controls which scopes are set based on openapi.yml.
	}
}

// handleBearerAuth validates the JWT bearer token and stores the user in context.
func handleBearerAuth(c *gin.Context, jwtService *jwt.JWTService) {
	token := extractBearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
		return
	}

	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	if claims.RawRole != string(entities.UserRoleMentor) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions: mentor role required"})
		return
	}

	user := &entities.User{
		ID:   claims.UserID,
		Name: claims.UserName,
		Role: claims.UserRole,
	}

	c.Set(GinKeyUser, user)
}

// handleInternalTokenAuth validates the X-Internal-Token header.
func handleInternalTokenAuth(c *gin.Context, expectedToken string) {
	token := c.GetHeader("X-Internal-Token")
	if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
}

// handleStaffBearerAuth validates the JWT bearer token and stores the staff in context.
func handleStaffBearerAuth(c *gin.Context, jwtService *jwt.JWTService) {
	token := extractBearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
		return
	}

	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	if claims.RawRole != jwt.RoleStaff {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions: staff role required"})
		return
	}

	staff := &entities.Staff{
		ID:   entities.StaffID(string(claims.UserID)),
		Name: claims.UserName,
	}

	c.Set(GinKeyStaff, staff)
}

// StaffFromContext retrieves the authenticated staff from the request context.
func StaffFromContext(ctx context.Context) *entities.Staff {
	return appcontext.StaffFromContext(ctx)
}

// UserFromContext retrieves the authenticated user from the request context.
func UserFromContext(ctx context.Context) *entities.User {
	return appcontext.UserFromContext(ctx)
}

// extractBearerToken extracts the token from the Authorization header value.
// Returns empty string if the value is not in "Bearer <token>" format.
func extractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return ""
	}

	token := strings.TrimSpace(authHeader[len(bearerPrefix):])
	if token == "" {
		return ""
	}

	return token
}
