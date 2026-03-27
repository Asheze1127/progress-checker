package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
)

// AuthHandler handles authentication-related HTTP endpoints.
type AuthHandler struct {
	loginUseCase *usecase.LoginUseCase
}

// NewAuthHandler creates a new AuthHandler with the given login use case.
func NewAuthHandler(loginUseCase *usecase.LoginUseCase) *AuthHandler {
	return &AuthHandler{loginUseCase: loginUseCase}
}

// HandleLogin handles POST /api/v1/auth/login requests.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req openapi.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.loginUseCase.Execute(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			WriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		if errors.Is(err, usecase.ErrUserNotMentor) {
			WriteError(w, http.StatusForbidden, "only mentors can log in")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSON(w, http.StatusOK, openapi.LoginResponse{
		Token: result.Token,
		User: openapi.UserResponse{
			Id:   string(result.User.ID),
			Name: result.User.Name,
			Role: string(result.User.Role),
		},
	})
}
