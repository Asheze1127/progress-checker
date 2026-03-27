package rest

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/Asheze1127/progress-checker/backend/api/openapi"
	githubsvc "github.com/Asheze1127/progress-checker/backend/application/service/github"
	"github.com/Asheze1127/progress-checker/backend/application/usecase"
	"github.com/Asheze1127/progress-checker/backend/entities"
)

// Compile-time check that StrictHandler implements openapi.StrictServerInterface.
var _ openapi.StrictServerInterface = (*StrictHandler)(nil)

// StrictHandler implements openapi.StrictServerInterface with typed request/response objects.
type StrictHandler struct {
	loginUseCase   *usecase.LoginUseCase
	listProgressUC *usecase.ListProgressUseCase
	githubService  *githubsvc.GitHubService
}

// NewStrictHandler creates a new StrictHandler.
func NewStrictHandler(
	loginUC *usecase.LoginUseCase,
	listProgressUC *usecase.ListProgressUseCase,
	githubService *githubsvc.GitHubService,
) *StrictHandler {
	return &StrictHandler{
		loginUseCase:   loginUC,
		listProgressUC: listProgressUC,
		githubService:  githubService,
	}
}

// Login handles POST /api/v1/auth/login.
func (h *StrictHandler) Login(ctx context.Context, request openapi.LoginRequestObject) (openapi.LoginResponseObject, error) {
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

// ListProgress handles GET /api/v1/progress.
func (h *StrictHandler) ListProgress(ctx context.Context, request openapi.ListProgressRequestObject) (openapi.ListProgressResponseObject, error) {
	var teamID string
	if request.Params.TeamId != nil {
		teamID = *request.Params.TeamId
	}

	results, err := h.listProgressUC.Execute(ctx, teamID)
	if err != nil {
		slog.Error("failed to list latest progress", slog.String("error", err.Error()))
		return openapi.ListProgress500JSONResponse{Error: "internal server error"}, nil
	}

	return openapi.ListProgress200JSONResponse(toProgressListResponse(results)), nil
}

// ListRepositories handles GET /api/v1/teams/{teamId}/github-repos.
func (h *StrictHandler) ListRepositories(ctx context.Context, request openapi.ListRepositoriesRequestObject) (openapi.ListRepositoriesResponseObject, error) {
	repos, err := h.githubService.ListRepositories(ctx, request.TeamId)
	if err != nil {
		slog.Error("failed to list repositories", slog.String("error", err.Error()))
		return openapi.ListRepositories400JSONResponse{Error: "failed to list repositories"}, nil
	}

	items := make([]openapi.RepoItem, 0, len(repos))
	for _, repo := range repos {
		items = append(items, openapi.RepoItem{
			Id:       string(repo.ID),
			Owner:    repo.Owner,
			RepoName: repo.RepoName,
		})
	}

	return openapi.ListRepositories200JSONResponse{Repos: items}, nil
}

// RegisterRepository handles POST /api/v1/teams/{teamId}/github-repos.
func (h *StrictHandler) RegisterRepository(ctx context.Context, request openapi.RegisterRepositoryRequestObject) (openapi.RegisterRepositoryResponseObject, error) {
	if err := h.githubService.RegisterRepository(ctx, request.TeamId, request.Body.GithubRepoUrl, request.Body.PersonalAccessToken); err != nil {
		slog.Error("failed to register repository", slog.String("error", err.Error()))
		return openapi.RegisterRepository400JSONResponse{Error: "failed to register repository"}, nil
	}

	return openapi.RegisterRepository201JSONResponse{Message: "repository registered successfully"}, nil
}

// RemoveRepository handles DELETE /api/v1/teams/{teamId}/github-repos/{repoId}.
func (h *StrictHandler) RemoveRepository(ctx context.Context, request openapi.RemoveRepositoryRequestObject) (openapi.RemoveRepositoryResponseObject, error) {
	if err := h.githubService.RemoveRepository(ctx, request.TeamId, request.RepoId); err != nil {
		slog.Error("failed to remove repository", slog.String("error", err.Error()))
		return openapi.RemoveRepository400JSONResponse{Error: "failed to remove repository"}, nil
	}

	return openapi.RemoveRepository200JSONResponse{Message: "repository removed successfully"}, nil
}

// UpdateToken handles PUT /api/v1/teams/{teamId}/github-repos/{repoId}/token.
func (h *StrictHandler) UpdateToken(ctx context.Context, request openapi.UpdateTokenRequestObject) (openapi.UpdateTokenResponseObject, error) {
	if err := h.githubService.UpdateToken(ctx, request.TeamId, request.RepoId, request.Body.PersonalAccessToken); err != nil {
		slog.Error("failed to update token", slog.String("error", err.Error()))
		return openapi.UpdateToken400JSONResponse{Error: "failed to update token"}, nil
	}

	return openapi.UpdateToken200JSONResponse{Message: "token updated successfully"}, nil
}

// CreateIssue handles POST /internal/issues.
func (h *StrictHandler) CreateIssue(ctx context.Context, request openapi.CreateIssueRequestObject) (openapi.CreateIssueResponseObject, error) {
	issueURL, err := h.githubService.CreateIssue(ctx, request.Body.ChannelId, request.Body.Title, request.Body.Body)
	if err != nil {
		slog.Error("failed to create issue", slog.String("error", err.Error()))
		return nil, err
	}

	return openapi.CreateIssue201JSONResponse{IssueUrl: issueURL}, nil
}

// toProgressListResponse converts domain objects to the API response format.
func toProgressListResponse(teamProgresses []entities.TeamProgress) openapi.ProgressListResponse {
	data := make([]openapi.TeamProgressResponse, 0, len(teamProgresses))

	for _, tp := range teamProgresses {
		item := openapi.TeamProgressResponse{
			TeamId:   string(tp.TeamID),
			TeamName: tp.TeamName,
		}

		if tp.LatestProgress != nil {
			item.LatestProgress = toLatestProgressResponse(tp.LatestProgress)
		}

		data = append(data, item)
	}

	return openapi.ProgressListResponse{Data: data}
}

// toLatestProgressResponse converts a ProgressLog entity to its API response format.
func toLatestProgressResponse(log *entities.ProgressLog) *openapi.LatestProgressResponse {
	bodies := make([]openapi.ProgressBodyResponse, 0, len(log.ProgressBodies))
	for _, b := range log.ProgressBodies {
		bodies = append(bodies, openapi.ProgressBodyResponse{
			Phase:       string(b.Phase),
			Sos:         b.SOS,
			Comment:     b.Comment,
			SubmittedAt: b.SubmittedAt.UTC().Truncate(time.Second),
		})
	}

	return &openapi.LatestProgressResponse{
		Id:             string(log.ID),
		ParticipantId:  string(log.ParticipantID),
		ProgressBodies: bodies,
	}
}
