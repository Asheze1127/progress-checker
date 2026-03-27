package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	loginUseCase *usecase.LoginUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(loginUC *usecase.LoginUseCase) *AuthHandler {
	return &AuthHandler{loginUseCase: loginUC}
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(ctx context.Context, request openapi.LoginRequestObject) (openapi.LoginResponseObject, error) {
	email := strings.TrimSpace(request.Body.Email)
	if email == "" || request.Body.Password == "" {
		return openapi.Login400JSONResponse{Error: "email and password are required"}, nil
	}

	result, err := h.loginUseCase.Execute(ctx, email, request.Body.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return openapi.Login401JSONResponse{Error: "invalid email or password"}, nil
		}
		if errors.Is(err, usecase.ErrUserNotMentor) {
			return openapi.Login403JSONResponse{Error: "only mentors can log in"}, nil
		}
		return nil, err
	}

	return openapi.Login200JSONResponse{
		Token: result.Token,
		User: openapi.UserResponse{
			Id:   string(result.User.ID),
			Name: result.User.Name,
			Role: string(result.User.Role),
		},
	}, nil
}
