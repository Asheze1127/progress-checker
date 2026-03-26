package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// AuthHandler handles authentication-related HTTP endpoints.
type AuthHandler struct {
	loginUseCase *usecase.LoginUseCase
}

// NewAuthHandler creates a new AuthHandler with the given login use case.
func NewAuthHandler(loginUseCase *usecase.LoginUseCase) *AuthHandler {
	return &AuthHandler{loginUseCase: loginUseCase}
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

// errorResponse represents a JSON error response.
type errorResponse struct {
	Error string `json:"error"`
}

// HandleLogin handles POST /api/v1/auth/login requests.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" || req.Password == "" {
		writeJSONError(w, "email and password are required", http.StatusBadRequest)
		return
	}

	result, err := h.loginUseCase.Execute(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			writeJSONError(w, "invalid email or password", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, usecase.ErrUserNotMentor) {
			writeJSONError(w, "only mentors can log in", http.StatusForbidden)
			return
		}
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, loginResponse{
		Token: result.Token,
		User:  toUserResponse(result.User),
	}, http.StatusOK)
}

// toUserResponse converts an entity User to a userResponse.
func toUserResponse(user entities.User) userResponse {
	return userResponse{
		ID:   string(user.ID),
		Name: user.Name,
		Role: string(user.Role),
	}
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// writeJSONError writes a JSON error response with the given status code.
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse{Error: message})
}
