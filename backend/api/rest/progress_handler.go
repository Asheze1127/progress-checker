package rest

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// progressBodyResponse is the JSON representation of a progress body.
type progressBodyResponse struct {
	Phase       string `json:"phase"`
	SOS         bool   `json:"sos"`
	Comment     string `json:"comment"`
	SubmittedAt string `json:"submitted_at"`
}

// latestProgressResponse is the JSON representation of a progress log.
type latestProgressResponse struct {
	ID              string                 `json:"id"`
	ParticipantID   string                 `json:"participant_id"`
	ProgressBodies  []progressBodyResponse `json:"progress_bodies"`
}

// teamProgressResponse is the JSON representation of a team's progress.
type teamProgressResponse struct {
	TeamID         string                  `json:"team_id"`
	TeamName       string                  `json:"team_name"`
	LatestProgress *latestProgressResponse `json:"latest_progress"`
}

// progressListResponse is the top-level JSON response.
type progressListResponse struct {
	Data []teamProgressResponse `json:"data"`
}

// ProgressHandler handles HTTP requests for progress data.
type ProgressHandler struct {
	listProgressUC *usecase.ListProgressUseCase
}

// NewProgressHandler creates a new ProgressHandler.
func NewProgressHandler(listProgressUC *usecase.ListProgressUseCase) *ProgressHandler {
	return &ProgressHandler{listProgressUC: listProgressUC}
}

// HandleListProgress handles GET /api/v1/progress requests.
func (h *ProgressHandler) HandleListProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	setCORSHeaders(w)

	teamID := r.URL.Query().Get("team_id")

	results, err := h.listProgressUC.Execute(r.Context(), teamID)
	if err != nil {
		slog.Error("failed to list latest progress", slog.String("error", err.Error()))
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := toProgressListResponse(results)
	WriteJSON(w, http.StatusOK, resp)
}

// HandleProgressPreflight handles CORS preflight requests for the progress endpoint.
func (h *ProgressHandler) HandleProgressPreflight(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

// setCORSHeaders adds CORS headers to the response for cross-origin access.
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// toProgressListResponse converts domain objects to the API response format.
func toProgressListResponse(teamProgresses []entities.TeamProgress) progressListResponse {
	data := make([]teamProgressResponse, 0, len(teamProgresses))

	for _, tp := range teamProgresses {
		item := teamProgressResponse{
			TeamID:   string(tp.TeamID),
			TeamName: tp.TeamName,
		}

		if tp.LatestProgress != nil {
			item.LatestProgress = toLatestProgressResponse(tp.LatestProgress)
		}

		data = append(data, item)
	}

	return progressListResponse{Data: data}
}

// toLatestProgressResponse converts a ProgressLog entity to its API response format.
func toLatestProgressResponse(log *entities.ProgressLog) *latestProgressResponse {
	bodies := make([]progressBodyResponse, 0, len(log.ProgressBodies))
	for _, b := range log.ProgressBodies {
		bodies = append(bodies, progressBodyResponse{
			Phase:       string(b.Phase),
			SOS:         b.SOS,
			Comment:     b.Comment,
			SubmittedAt: b.SubmittedAt.Format(time.RFC3339),
		})
	}

	return &latestProgressResponse{
		ID:             string(log.ID),
		ParticipantID:  string(log.ParticipantID),
		ProgressBodies: bodies,
	}
}
