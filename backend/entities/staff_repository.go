package entities

import "context"

// StaffRepository defines the interface for querying staff.
type StaffRepository interface {
	FindByEmail(ctx context.Context, email string) (*StaffWithPassword, error)
	GetByID(ctx context.Context, id StaffID) (*Staff, error)
}
