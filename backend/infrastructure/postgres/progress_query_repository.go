package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

// Compile-time check that ProgressQueryRepository implements the interface.
var _ entities.ProgressQueryRepository = (*ProgressQueryRepository)(nil)

// ProgressQueryRepository implements entities.ProgressQueryRepository using PostgreSQL.
type ProgressQueryRepository struct {
	db *sql.DB
}

// NewProgressQueryRepository creates a new ProgressQueryRepository.
func NewProgressQueryRepository(db *sql.DB) *ProgressQueryRepository {
	return &ProgressQueryRepository{db: db}
}

const latestProgressQuery = `
SELECT
    t.id          AS team_id,
    t.name        AS team_name,
    pl.id         AS progress_log_id,
    pl.participant_id,
    pb.phase,
    pb.sos,
    COALESCE(pb.comment, '') AS comment,
    pb.submitted_at
FROM teams t
LEFT JOIN LATERAL (
    SELECT pl2.id, pl2.participant_id
    FROM progress_logs pl2
    JOIN participants p ON p.user_id = pl2.participant_id
    WHERE p.team_id = t.id
    ORDER BY pl2.created_at DESC
    LIMIT 1
) pl ON TRUE
LEFT JOIN progress_bodies pb ON pb.progress_log_id = pl.id
`

// ListLatestByTeam returns the latest progress for all teams.
func (r *ProgressQueryRepository) ListLatestByTeam(ctx context.Context) ([]entities.TeamProgress, error) {
	query := latestProgressQuery + `ORDER BY t.name, pb.submitted_at`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query latest progress by team: %w", err)
	}
	defer rows.Close()

	return scanTeamProgressRows(rows)
}

// ListLatestByTeamID returns the latest progress filtered by team ID.
func (r *ProgressQueryRepository) ListLatestByTeamID(ctx context.Context, teamID entities.TeamID) ([]entities.TeamProgress, error) {
	query := latestProgressQuery + `WHERE t.id = $1 ORDER BY t.name, pb.submitted_at`

	rows, err := r.db.QueryContext(ctx, query, string(teamID))
	if err != nil {
		return nil, fmt.Errorf("query latest progress by team id: %w", err)
	}
	defer rows.Close()

	return scanTeamProgressRows(rows)
}

// scanTeamProgressRows reads rows from the joined query and groups them
// into TeamProgress entities. Each team appears once, with all progress
// bodies collected into a single ProgressLog.
func scanTeamProgressRows(rows *sql.Rows) ([]entities.TeamProgress, error) {
	teamMap := make(map[entities.TeamID]*entities.TeamProgress)
	var orderedTeamIDs []entities.TeamID

	for rows.Next() {
		var (
			teamID        string
			teamName      string
			progressLogID sql.NullString
			participantID sql.NullString
			phase         sql.NullString
			sos           sql.NullBool
			comment       sql.NullString
			submittedAt   sql.NullTime
		)

		if err := rows.Scan(
			&teamID,
			&teamName,
			&progressLogID,
			&participantID,
			&phase,
			&sos,
			&comment,
			&submittedAt,
		); err != nil {
			return nil, fmt.Errorf("scan progress row: %w", err)
		}

		tid := entities.TeamID(teamID)

		tp, exists := teamMap[tid]
		if !exists {
			tp = &entities.TeamProgress{
				TeamID:   tid,
				TeamName: teamName,
			}
			teamMap[tid] = tp
			orderedTeamIDs = append(orderedTeamIDs, tid)
		}

		if !progressLogID.Valid {
			continue
		}

		if tp.LatestProgress == nil {
			tp.LatestProgress = &entities.ProgressLog{
				ID:            entities.ProgressLogID(progressLogID.String),
				ParticipantID: entities.ParticipantID(participantID.String),
			}
		}

		if phase.Valid {
			tp.LatestProgress.ProgressBodies = append(
				tp.LatestProgress.ProgressBodies,
				entities.ProgressBody{
					Phase:       entities.ProgressPhase(phase.String),
					SOS:         sos.Bool,
					Comment:     comment.String,
					SubmittedAt: submittedAt.Time.UTC().Truncate(time.Microsecond),
				},
			)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate progress rows: %w", err)
	}

	results := make([]entities.TeamProgress, 0, len(orderedTeamIDs))
	for _, tid := range orderedTeamIDs {
		results = append(results, *teamMap[tid])
	}

	return results, nil
}
