package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/application/appcontext"
	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	slackinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/slack"
	"github.com/google/uuid"
)

var (
	ErrTeamNotFound    = errors.New("team not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type RegisterParticipantUseCase struct {
	userRepo        entities.UserRepository
	teamRepo        entities.TeamRepository
	mentorRepo      entities.MentorRepository
	slackClient     SlackUserInfoFetcher
	database        *sql.DB
}

// SlackUserInfoFetcher is already defined in create_mentor.go but we need it here too.
// Since both are in the same package, the interface is shared.

func NewRegisterParticipantUseCase(
	userRepo entities.UserRepository,
	teamRepo entities.TeamRepository,
	mentorRepo entities.MentorRepository,
	slackClient SlackUserInfoFetcher,
	database *sql.DB,
) *RegisterParticipantUseCase {
	return &RegisterParticipantUseCase{
		userRepo:    userRepo,
		teamRepo:    teamRepo,
		mentorRepo:  mentorRepo,
		slackClient: slackClient,
		database:    database,
	}
}

func (uc *RegisterParticipantUseCase) Execute(ctx context.Context, slackUserID, teamID string) (result *entities.User, err error) {
	defer func() {
		attrs := []slog.Attr{slog.String("slack_user_id", slackUserID), slog.String("team_id", teamID)}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
		slog.LogAttrs(ctx, slog.LevelDebug, "RegisterParticipantUseCase.Execute", attrs...)
	}()

	if strings.TrimSpace(slackUserID) == "" {
		return nil, fmt.Errorf("slack_user_id is required")
	}
	if strings.TrimSpace(teamID) == "" {
		return nil, fmt.Errorf("team_id is required")
	}

	// Get the authenticated mentor from context
	mentorUser := appcontext.UserFromContext(ctx)
	if mentorUser == nil {
		return nil, ErrNotAuthorized
	}

	// Verify the mentor is assigned to this team
	mentor, err := uc.mentorRepo.GetByUserID(ctx, mentorUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentor: %w", err)
	}

	if !mentor.BelongsToTeam(entities.TeamID(teamID)) {
		return nil, ErrNotAuthorizedForTeam
	}

	// Verify team exists
	if _, err := uc.teamRepo.GetByID(ctx, entities.TeamID(teamID)); err != nil {
		return nil, ErrTeamNotFound
	}

	// Check if user already exists
	if existing, existErr := uc.userRepo.GetBySlackUserID(ctx, entities.SlackUserID(slackUserID)); existErr == nil && existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// Fetch user info from Slack
	var slackUser *slackinfra.SlackUserInfo
	slackUser, err = uc.slackClient.GetUserInfo(ctx, slackUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch slack user info: %w", err)
	}

	name := slackUser.RealName
	if name == "" {
		name = slackUserID
	}

	email := slackUser.Email
	if email == "" {
		return nil, fmt.Errorf("slack user has no email address configured")
	}

	teamUUID, err := uuid.Parse(teamID)
	if err != nil {
		return nil, fmt.Errorf("invalid team ID: %w", err)
	}

	// Execute both DB writes in a single transaction
	tx, err := uc.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := db.New(tx)

	userRow, err := qtx.CreateUser(ctx, db.CreateUserParams{
		SlackUserID:  slackUserID,
		Name:         name,
		Email:        email,
		Role:         string(entities.UserRoleParticipant),
		PasswordHash: "",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create participant: %w", err)
	}

	if err := qtx.CreateParticipant(ctx, db.CreateParticipantParams{
		UserID: userRow.ID,
		TeamID: teamUUID,
	}); err != nil {
		return nil, fmt.Errorf("failed to create participant record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	createdUser := &entities.User{
		ID:          entities.UserID(userRow.ID.String()),
		SlackUserID: entities.SlackUserID(userRow.SlackUserID),
		Name:        userRow.Name,
		Email:       userRow.Email,
		Role:        entities.UserRole(userRow.Role),
	}

	return createdUser, nil
}
