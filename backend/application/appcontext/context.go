package appcontext

import (
	"context"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// Context keys shared between middleware and use cases.
// Using plain strings so Gin's context Value() lookup works correctly.
const (
	KeyUser  = "middleware.authenticated_user"
	KeyStaff = "middleware.authenticated_staff"
)

// UserFromContext retrieves the authenticated user from the request context.
func UserFromContext(ctx context.Context) *entities.User {
	user, ok := ctx.Value(KeyUser).(*entities.User)
	if !ok {
		return nil
	}
	return user
}

// StaffFromContext retrieves the authenticated staff from the request context.
func StaffFromContext(ctx context.Context) *entities.Staff {
	staff, ok := ctx.Value(KeyStaff).(*entities.Staff)
	if !ok {
		return nil
	}
	return staff
}
