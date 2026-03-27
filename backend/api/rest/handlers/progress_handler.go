package handlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// ProgressHandler handles progress-related endpoints.
type ProgressHandler struct {
	listProgressUC *usecase.ListProgressUseCase
}

// NewProgressHandler creates a new ProgressHandler.
func NewProgressHandler(listProgressUC *usecase.ListProgressUseCase) *ProgressHandler {
	return &ProgressHandler{listProgressUC: listProgressUC}
}

// ListProgress handles GET /api/v1/progress.
func (h *ProgressHandler) ListProgress(ctx context.Context, request openapi.ListProgressRequestObject) (openapi.ListProgressResponseObject, error) {
	var teamID string
	if request.Params.TeamId != nil {
		teamID = *request.Params.TeamId
	}

	results, err := h.listProgressUC.Execute(ctx, teamID)
	if err != nil {
		slog.Error("failed to list latest progress", slog.String("error", err.Error()))
		return openapi.ListProgress500JSONResponse{Error: "internal server error"}, nil
	}

	return openapi.ListProgress200JSONResponse(toProgressListResponse(results)), nil
}

func toProgressListResponse(teamProgresses []entities.TeamProgress) openapi.ProgressListResponse {
	data := make([]openapi.TeamProgressResponse, 0, len(teamProgresses))

	for _, tp := range teamProgresses {
		item := openapi.TeamProgressResponse{
			TeamId:   string(tp.TeamID),
			TeamName: tp.TeamName,
		}

		if tp.LatestProgress != nil {
			item.LatestProgress = toLatestProgressResponse(tp.LatestProgress)
		}

		data = append(data, item)
	}

	return openapi.ProgressListResponse{Data: data}
}

func toLatestProgressResponse(log *entities.ProgressLog) *openapi.LatestProgressResponse {
	bodies := make([]openapi.ProgressBodyResponse, 0, len(log.ProgressBodies))
	for _, b := range log.ProgressBodies {
		bodies = append(bodies, openapi.ProgressBodyResponse{
			Phase:       string(b.Phase),
			Sos:         b.SOS,
			Comment:     b.Comment,
			SubmittedAt: b.SubmittedAt.UTC().Truncate(time.Second),
		})
	}

	return &openapi.LatestProgressResponse{
		Id:             string(log.ID),
		ParticipantId:  string(log.ParticipantID),
		ProgressBodies: bodies,
	}
}
