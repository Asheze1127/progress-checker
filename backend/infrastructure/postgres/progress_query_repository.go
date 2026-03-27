package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/google/uuid"
)

// Compile-time check that ProgressQueryRepository implements the interface.
var _ entities.ProgressQueryRepository = (*ProgressQueryRepository)(nil)

// ProgressQueryRepository implements entities.ProgressQueryRepository using PostgreSQL.
type ProgressQueryRepository struct {
	queries *db.Queries
}

// NewProgressQueryRepository creates a new ProgressQueryRepository.
func NewProgressQueryRepository(database *sql.DB) *ProgressQueryRepository {
	return &ProgressQueryRepository{queries: db.New(database)}
}

// ListLatestByTeam returns the latest progress for all teams.
func (r *ProgressQueryRepository) ListLatestByTeam(ctx context.Context) ([]entities.TeamProgress, error) {
	rows, err := r.queries.GetLatestProgressByTeam(ctx)
	if err != nil {
		return nil, fmt.Errorf("query latest progress by team: %w", err)
	}
	return toTeamProgressList(rows), nil
}

// ListLatestByTeamID returns the latest progress filtered by team ID.
func (r *ProgressQueryRepository) ListLatestByTeamID(ctx context.Context, teamID entities.TeamID) ([]entities.TeamProgress, error) {
	uid, err := uuid.Parse(string(teamID))
	if err != nil {
		return nil, fmt.Errorf("parse team id: %w", err)
	}
	rows, err := r.queries.GetLatestProgressByTeamID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("query latest progress by team id: %w", err)
	}
	return toTeamProgressListByID(rows), nil
}

// progressRow is an intermediate representation shared by both sqlc query row types.
type progressRow struct {
	TeamID        uuid.UUID
	TeamName      string
	ProgressLogID uuid.NullUUID
	ParticipantID uuid.NullUUID
	LogCreatedAt  sql.NullTime
	BodyID        uuid.NullUUID
	Phase         db.NullProgressPhase
	Sos           sql.NullBool
	Comment       sql.NullString
	SubmittedAt   sql.NullTime
}

func toTeamProgressList(rows []db.GetLatestProgressByTeamRow) []entities.TeamProgress {
	converted := make([]progressRow, len(rows))
	for i, r := range rows {
		converted[i] = progressRow{
			TeamID:        r.TeamID,
			TeamName:      r.TeamName,
			ProgressLogID: r.ProgressLogID,
			ParticipantID: r.ParticipantID,
			LogCreatedAt:  r.LogCreatedAt,
			BodyID:        r.BodyID,
			Phase:         r.Phase,
			Sos:           r.Sos,
			Comment:       r.Comment,
			SubmittedAt:   r.SubmittedAt,
		}
	}
	return toTeamProgressListFrom(converted)
}

func toTeamProgressListByID(rows []db.GetLatestProgressByTeamIDRow) []entities.TeamProgress {
	converted := make([]progressRow, len(rows))
	for i, r := range rows {
		converted[i] = progressRow{
			TeamID:        r.TeamID,
			TeamName:      r.TeamName,
			ProgressLogID: r.ProgressLogID,
			ParticipantID: r.ParticipantID,
			LogCreatedAt:  r.LogCreatedAt,
			BodyID:        r.BodyID,
			Phase:         r.Phase,
			Sos:           r.Sos,
			Comment:       r.Comment,
			SubmittedAt:   r.SubmittedAt,
		}
	}
	return toTeamProgressListFrom(converted)
}

func toTeamProgressListFrom(rows []progressRow) []entities.TeamProgress {
	teamMap := make(map[uuid.UUID]*entities.TeamProgress)
	var orderedIDs []uuid.UUID

	for _, row := range rows {
		tp, exists := teamMap[row.TeamID]
		if !exists {
			tp = &entities.TeamProgress{
				TeamID:   entities.TeamID(row.TeamID.String()),
				TeamName: row.TeamName,
			}
			teamMap[row.TeamID] = tp
			orderedIDs = append(orderedIDs, row.TeamID)
		}

		if !row.BodyID.Valid {
			continue
		}

		if tp.LatestProgress == nil && row.ProgressLogID.Valid {
			tp.LatestProgress = &entities.ProgressLog{
				ID:            entities.ProgressLogID(row.ProgressLogID.UUID.String()),
				ParticipantID: entities.ParticipantID(row.ParticipantID.UUID.String()),
			}
		}

		if row.Phase.Valid {
			tp.LatestProgress.ProgressBodies = append(
				tp.LatestProgress.ProgressBodies,
				toProgressBody(row.Phase, row.Sos, row.Comment, row.SubmittedAt),
			)
		}
	}

	results := make([]entities.TeamProgress, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		results = append(results, *teamMap[id])
	}
	return results
}

func toProgressBody(phase db.NullProgressPhase, sos sql.NullBool, comment sql.NullString, submittedAt sql.NullTime) entities.ProgressBody {
	return entities.ProgressBody{
		Phase:       entities.ProgressPhase(phase.ProgressPhase),
		SOS:         sos.Bool,
		Comment:     comment.String,
		SubmittedAt: submittedAt.Time.UTC().Truncate(time.Microsecond),
	}
}
