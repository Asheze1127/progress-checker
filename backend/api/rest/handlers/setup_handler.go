package handlers

import (
	"context"
	"errors"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
)

// SetupHandler handles password setup endpoints.
type SetupHandler struct {
	setupPasswordUC *usecase.SetupPasswordUseCase
}

// NewSetupHandler creates a new SetupHandler.
func NewSetupHandler(setupPasswordUC *usecase.SetupPasswordUseCase) *SetupHandler {
	return &SetupHandler{setupPasswordUC: setupPasswordUC}
}

func (h *SetupHandler) SetupPassword(ctx context.Context, request openapi.SetupPasswordRequestObject) (openapi.SetupPasswordResponseObject, error) {
	if request.Body.Token == "" || request.Body.Password == "" {
		return openapi.SetupPassword400JSONResponse{Error: "token and password are required"}, nil
	}

	err := h.setupPasswordUC.Execute(ctx, request.Body.Token, request.Body.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrPasswordTooShort) {
			return openapi.SetupPassword400JSONResponse{Error: err.Error()}, nil
		}
		if errors.Is(err, usecase.ErrSetupTokenInvalid) ||
			errors.Is(err, usecase.ErrSetupTokenExpired) ||
			errors.Is(err, usecase.ErrSetupTokenUsed) {
			return openapi.SetupPassword400JSONResponse{Error: "invalid or expired setup token"}, nil
		}
		return nil, err
	}

	return openapi.SetupPassword200JSONResponse{Message: "password set successfully"}, nil
}
