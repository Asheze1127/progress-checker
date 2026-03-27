package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/samber/do/v2"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// AuthHandler handles authentication-related HTTP endpoints.
type AuthHandler struct {
	loginUseCase *usecase.LoginUseCase
}

// NewAuthHandler creates a new AuthHandler via DI container.
func NewAuthHandler(i do.Injector) (*AuthHandler, error) {
	loginUseCase := do.MustInvoke[*usecase.LoginUseCase](i)
	return &AuthHandler{loginUseCase: loginUseCase}, nil
}

// loginRequest represents the JSON body of a login request.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginResponse represents the JSON response of a successful login.
type loginResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

// userResponse represents the user data returned in API responses.
type userResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// HandleLogin handles POST /api/v1/auth/login requests.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req loginRequest
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

	WriteJSON(w, http.StatusOK, loginResponse{
		Token: result.Token,
		User:  toUserResponse(result.User),
	})
}

// toUserResponse converts an entity User to a userResponse.
func toUserResponse(user entities.User) userResponse {
	return userResponse{
		ID:   string(user.ID),
		Name: user.Name,
		Role: string(user.Role),
	}
}
