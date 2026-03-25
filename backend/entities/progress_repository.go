package entities

import "context"

// ProgressRepository defines the interface for persisting progress logs.
type ProgressRepository interface {
	Save(ctx context.Context, log *ProgressLog) error
}
