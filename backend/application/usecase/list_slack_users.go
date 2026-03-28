package usecase

import (
	"context"
	"fmt"
	"log/slog"

	slackinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/slack"
)

// SlackUserLister lists users from the Slack workspace.
type SlackUserLister interface {
	GetUsers(ctx context.Context) ([]slackinfra.SlackUserInfo, error)
}

type ListSlackUsersUseCase struct {
	slackClient SlackUserLister
}

func NewListSlackUsersUseCase(slackClient SlackUserLister) *ListSlackUsersUseCase {
	return &ListSlackUsersUseCase{slackClient: slackClient}
}

func (uc *ListSlackUsersUseCase) Execute(ctx context.Context) (result []slackinfra.SlackUserInfo, err error) {
	defer func() {
		attrs := []slog.Attr{slog.Int("count", len(result))}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
		slog.LogAttrs(ctx, slog.LevelDebug, "ListSlackUsersUseCase.Execute", attrs...)
	}()

	users, err := uc.slackClient.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list slack users: %w", err)
	}

	return users, nil
}
