package handlers

import (
	"context"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
)

// ParticipantHandler handles participant registration endpoints.
type ParticipantHandler struct {
	registerParticipantUC    *usecase.RegisterParticipantUseCase
	listTeamParticipantsUC *usecase.ListTeamParticipantsUseCase
}

// NewParticipantHandler creates a new ParticipantHandler.
func NewParticipantHandler(
	registerParticipantUC *usecase.RegisterParticipantUseCase,
	listTeamParticipantsUC *usecase.ListTeamParticipantsUseCase,
) *ParticipantHandler {
	return &ParticipantHandler{
		registerParticipantUC:    registerParticipantUC,
		listTeamParticipantsUC: listTeamParticipantsUC,
	}
}

func (h *ParticipantHandler) ListTeamParticipants(ctx context.Context, request openapi.ListTeamParticipantsRequestObject) (openapi.ListTeamParticipantsResponseObject, error) {
	participants, err := h.listTeamParticipantsUC.Execute(ctx, request.TeamId)
	if err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			return openapi.ListTeamParticipants403JSONResponse{Error: err.Error()}, nil
		}
		return openapi.ListTeamParticipants500JSONResponse{Error: "failed to list participants"}, nil
	}

	resp := make([]openapi.ParticipantResponse, len(participants))
	for i, p := range participants {
		resp[i] = openapi.ParticipantResponse{
			Id:          string(p.ID),
			SlackUserId: string(p.SlackUserID),
			Name:        p.Name,
			Email:       p.Email,
		}
	}

	return openapi.ListTeamParticipants200JSONResponse{Participants: resp}, nil
}

func (h *ParticipantHandler) RegisterParticipant(ctx context.Context, request openapi.RegisterParticipantRequestObject) (openapi.RegisterParticipantResponseObject, error) {
	body := request.Body
	if strings.TrimSpace(body.SlackUserId) == "" {
		return openapi.RegisterParticipant400JSONResponse{Error: "slack_user_id is required"}, nil
	}
	if strings.TrimSpace(body.TeamId) == "" {
		return openapi.RegisterParticipant400JSONResponse{Error: "team_id is required"}, nil
	}

	result, err := h.registerParticipantUC.Execute(ctx, body.SlackUserId, body.TeamId)
	if err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			return openapi.RegisterParticipant403JSONResponse{Error: err.Error()}, nil
		}
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already") {
			return openapi.RegisterParticipant400JSONResponse{Error: err.Error()}, nil
		}
		return nil, err
	}

	return openapi.RegisterParticipant201JSONResponse{
		User: openapi.UserResponse{
			Id:   string(result.ID),
			Name: result.Name,
			Role: string(result.Role),
		},
	}, nil
}
