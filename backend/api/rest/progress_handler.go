package rest

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// ProgressHandler handles HTTP requests for progress data.
type ProgressHandler struct {
	listProgressUC *usecase.ListProgressUseCase
	allowedOrigins []string
}

// NewProgressHandler creates a new ProgressHandler.
func NewProgressHandler(listProgressUC *usecase.ListProgressUseCase, allowedOrigins []string) *ProgressHandler {
	return &ProgressHandler{
		listProgressUC: listProgressUC,
		allowedOrigins: allowedOrigins,
	}
}

// HandleListProgress handles GET /api/v1/progress requests.
func (h *ProgressHandler) HandleListProgress(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w, r)

	teamID := r.URL.Query().Get("team_id")

	results, err := h.listProgressUC.Execute(r.Context(), teamID)
	if err != nil {
		slog.Error("failed to list latest progress", slog.String("error", err.Error()))
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSON(w, http.StatusOK, toProgressListResponse(results))
}

// HandleProgressPreflight handles CORS preflight requests for the progress endpoint.
func (h *ProgressHandler) HandleProgressPreflight(w http.ResponseWriter, r *http.Request) {
	h.setCORSHeaders(w, r)
	w.WriteHeader(http.StatusNoContent)
}

// setCORSHeaders adds CORS headers to the response for cross-origin access.
func (h *ProgressHandler) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}

	for _, allowed := range h.allowedOrigins {
		if origin == allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			return
		}
	}
}

// toProgressListResponse converts domain objects to the API response format.
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

// toLatestProgressResponse converts a ProgressLog entity to its API response format.
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
