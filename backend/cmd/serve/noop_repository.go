package serve

import (
	"context"
	"log"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// noopProgressRepository is a temporary implementation that logs but does not persist.
// This will be replaced when database infrastructure is implemented.
type noopProgressRepository struct{}

func (r *noopProgressRepository) Save(_ context.Context, progressLog *entities.ProgressLog) error {
	log.Printf("INFO: progress log saved (noop): id=%s participant=%s", progressLog.ID, progressLog.ParticipantID)
	return nil
}
