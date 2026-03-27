package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// StaffHandler handles staff management endpoints.
type StaffHandler struct {
	staffLoginUC *usecase.StaffLoginUseCase
	createTeamUC *usecase.CreateTeamUseCase
	createUserUC *usecase.CreateUserUseCase
	teamRepo     entities.TeamRepository
	userRepo     entities.UserRepository
}

// NewStaffHandler creates a new StaffHandler.
func NewStaffHandler(
	staffLoginUC *usecase.StaffLoginUseCase,
	createTeamUC *usecase.CreateTeamUseCase,
	createUserUC *usecase.CreateUserUseCase,
	teamRepo entities.TeamRepository,
	userRepo entities.UserRepository,
) *StaffHandler {
	return &StaffHandler{
		staffLoginUC: staffLoginUC,
		createTeamUC: createTeamUC,
		createUserUC: createUserUC,
		teamRepo:     teamRepo,
		userRepo:     userRepo,
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
		return nil, err
	}

	return openapi.StaffCreateTeam201JSONResponse{
		Team: openapi.TeamResponse{
			Id:   string(team.ID),
			Name: team.Name,
		},
	}, nil
}

func (h *StaffHandler) StaffListUsers(ctx context.Context, _ openapi.StaffListUsersRequestObject) (openapi.StaffListUsersResponseObject, error) {
	users, err := h.userRepo.List(ctx)
	if err != nil {
		return openapi.StaffListUsers500JSONResponse{Error: "failed to list users"}, nil
	}

	items := make([]openapi.UserResponse, len(users))
	for i, u := range users {
		items[i] = openapi.UserResponse{
			Id:   string(u.ID),
			Name: u.Name,
			Role: string(u.Role),
		}
	}

	return openapi.StaffListUsers200JSONResponse{Users: items}, nil
}

func (h *StaffHandler) StaffCreateUser(ctx context.Context, request openapi.StaffCreateUserRequestObject) (openapi.StaffCreateUserResponseObject, error) {
	body := request.Body
	if strings.TrimSpace(body.SlackUserId) == "" {
		return openapi.StaffCreateUser400JSONResponse{Error: "slack_user_id is required"}, nil
	}
	if strings.TrimSpace(body.Name) == "" {
		return openapi.StaffCreateUser400JSONResponse{Error: "name is required"}, nil
	}
	if strings.TrimSpace(body.Email) == "" {
		return openapi.StaffCreateUser400JSONResponse{Error: "email is required"}, nil
	}

	role := entities.UserRole(body.Role)
	if role != entities.UserRoleParticipant && role != entities.UserRoleMentor {
		return openapi.StaffCreateUser400JSONResponse{Error: "role must be participant or mentor"}, nil
	}

	result, err := h.createUserUC.Execute(
		ctx,
		body.SlackUserId,
		body.Name,
		body.Email,
		role,
		entities.TeamID(body.TeamId),
	)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "is required") {
			return openapi.StaffCreateUser400JSONResponse{Error: err.Error()}, nil
		}
		return nil, err
	}

	return openapi.StaffCreateUser201JSONResponse{
		User: openapi.UserResponse{
			Id:   string(result.User.ID),
			Name: result.User.Name,
			Role: string(result.User.Role),
		},
		SetupUrl: result.SetupURL,
	}, nil
}
