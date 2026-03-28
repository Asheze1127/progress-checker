package handlers

import (
	"context"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
)

type TeamHandler struct {
	listMentorTeamsUC *usecase.ListMentorTeamsUseCase
}

func NewTeamHandler(listMentorTeamsUC *usecase.ListMentorTeamsUseCase) *TeamHandler {
	return &TeamHandler{listMentorTeamsUC: listMentorTeamsUC}
}

func (h *TeamHandler) ListMentorTeams(ctx context.Context, _ openapi.ListMentorTeamsRequestObject) (openapi.ListMentorTeamsResponseObject, error) {
	teams, err := h.listMentorTeamsUC.Execute(ctx)
	if err != nil {
		return openapi.ListMentorTeams500JSONResponse{Error: "failed to list teams"}, nil
	}

	resp := make([]openapi.TeamResponse, len(teams))
	for i, t := range teams {
		resp[i] = openapi.TeamResponse{
			Id:   string(t.ID),
			Name: t.Name,
		}
	}

	return openapi.ListMentorTeams200JSONResponse{Teams: resp}, nil
}
