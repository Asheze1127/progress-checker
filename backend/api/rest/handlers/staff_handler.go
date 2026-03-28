package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// StaffHandler handles staff management endpoints.
type StaffHandler struct {
	staffLoginUC *usecase.StaffLoginUseCase
	createTeamUC *usecase.CreateTeamUseCase
	teamRepo     entities.TeamRepository
}

// NewStaffHandler creates a new StaffHandler.
func NewStaffHandler(
	staffLoginUC *usecase.StaffLoginUseCase,
	createTeamUC *usecase.CreateTeamUseCase,
	teamRepo entities.TeamRepository,
) *StaffHandler {
	return &StaffHandler{
		staffLoginUC: staffLoginUC,
		createTeamUC: createTeamUC,
		teamRepo:     teamRepo,
	}
}

func (h *StaffHandler) StaffLogin(ctx context.Context, request openapi.StaffLoginRequestObject) (openapi.StaffLoginResponseObject, error) {
	email := strings.TrimSpace(request.Body.Email)
	if email == "" || request.Body.Password == "" {
		return openapi.StaffLogin400JSONResponse{Error: "email and password are required"}, nil
	}

	result, err := h.staffLoginUC.Execute(ctx, email, request.Body.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidStaffCredentials) {
			return openapi.StaffLogin401JSONResponse{Error: "invalid email or password"}, nil
		}
		return nil, err
	}

	return openapi.StaffLogin200JSONResponse{
		Token: result.Token,
		Staff: openapi.StaffResponse{
			Id:   string(result.Staff.ID),
			Name: result.Staff.Name,
		},
	}, nil
}

func (h *StaffHandler) StaffListTeams(ctx context.Context, _ openapi.StaffListTeamsRequestObject) (openapi.StaffListTeamsResponseObject, error) {
	teams, err := h.teamRepo.List(ctx)
	if err != nil {
		return openapi.StaffListTeams500JSONResponse{Error: "failed to list teams"}, nil
	}

	items := make([]openapi.TeamResponse, len(teams))
	for i, t := range teams {
		items[i] = openapi.TeamResponse{
			Id:   string(t.ID),
			Name: t.Name,
		}
	}

	return openapi.StaffListTeams200JSONResponse{Teams: items}, nil
}

func (h *StaffHandler) StaffCreateTeam(ctx context.Context, request openapi.StaffCreateTeamRequestObject) (openapi.StaffCreateTeamResponseObject, error) {
	name := strings.TrimSpace(request.Body.Name)
	if name == "" {
		return openapi.StaffCreateTeam400JSONResponse{Error: "team name is required"}, nil
	}

	team, err := h.createTeamUC.Execute(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return openapi.StaffCreateTeam201JSONResponse{
		Team: openapi.TeamResponse{
			Id:   string(team.ID),
			Name: team.Name,
		},
	}, nil
}
