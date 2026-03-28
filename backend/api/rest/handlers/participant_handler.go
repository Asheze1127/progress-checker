package handlers

import (
	"context"
	"errors"
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
		if errors.Is(err, usecase.ErrNotAuthorized) || errors.Is(err, usecase.ErrNotAuthorizedForTeam) {
			return openapi.ListTeamParticipants403JSONResponse{Error: "not authorized"}, nil
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
		if errors.Is(err, usecase.ErrNotAuthorized) || errors.Is(err, usecase.ErrNotAuthorizedForTeam) {
			return openapi.RegisterParticipant403JSONResponse{Error: "not authorized"}, nil
		}
		if errors.Is(err, usecase.ErrTeamNotFound) || errors.Is(err, usecase.ErrUserAlreadyExists) {
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
