package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// contextKey is an unexported type used for context keys to avoid collisions.
type contextKey string

// userContextKey is the context key for storing the authenticated user.
const userContextKey contextKey = "authenticated_user"

// SessionValidator defines the interface for validating session tokens.
type SessionValidator interface {
	// ValidateSession checks a token and returns the associated user or an error.
	ValidateSession(ctx context.Context, token string) (*entities.User, error)
}

// AuthMiddleware creates an HTTP middleware that authenticates requests using
// Bearer tokens in the Authorization header.
func AuthMiddleware(validator SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			user, err := validator.ValidateSession(r.Context(), token)
			if err != nil {
				if isForbiddenError(err) {
					http.Error(w, `{"error":"insufficient permissions: mentor role required"}`, http.StatusForbidden)
					return
				}
				http.Error(w, `{"error":"invalid or expired session"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext retrieves the authenticated user from the request context.
// Returns nil if no user is present.
func UserFromContext(ctx context.Context) *entities.User {
	user, ok := ctx.Value(userContextKey).(*entities.User)
	if !ok {
		return nil
	}
	return user
}

// extractBearerToken extracts the token from the Authorization header.
// Returns empty string if the header is missing or not in "Bearer <token>" format.
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
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

// isForbiddenError checks if an error indicates a permission issue (non-mentor role).
func isForbiddenError(err error) bool {
	return strings.Contains(err.Error(), "not a mentor")
}
