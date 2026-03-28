package handlers

import (
	"context"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
)

type SlackHandler struct {
	listSlackUsersUC *usecase.ListSlackUsersUseCase
}

func NewSlackHandler(listSlackUsersUC *usecase.ListSlackUsersUseCase) *SlackHandler {
	return &SlackHandler{listSlackUsersUC: listSlackUsersUC}
}

func (h *SlackHandler) ListSlackUsers(ctx context.Context, _ openapi.ListSlackUsersRequestObject) (openapi.ListSlackUsersResponseObject, error) {
	users, err := h.listSlackUsersUC.Execute(ctx)
	if err != nil {
		return openapi.ListSlackUsers500JSONResponse{Error: "failed to fetch slack users"}, nil
	}

	resp := make([]openapi.SlackUserResponse, len(users))
	for i, u := range users {
		resp[i] = openapi.SlackUserResponse{
			Id:    u.ID,
			Name:  u.RealName,
			Email: u.Email,
		}
	}

	return openapi.ListSlackUsers200JSONResponse{Users: resp}, nil
}
