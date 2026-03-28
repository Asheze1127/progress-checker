package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	slackinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/slack"
	"github.com/google/uuid"
)

var (
	ErrNotStaff = errors.New("caller is not a registered staff member")
)

const setupTokenExpiry = 24 * time.Hour

type CreateMentorResult struct {
	User     entities.User
	SetupURL string
}

// SlackUserInfoFetcher fetches user info from Slack.
type SlackUserInfoFetcher interface {
	GetUserInfo(ctx context.Context, userID string) (*slackinfra.SlackUserInfo, error)
}

type CreateMentorUseCase struct {
	staffRepo       entities.StaffRepository
	userRepo        entities.UserRepository
	teamRepo        entities.TeamRepository
	slackClient     SlackUserInfoFetcher
	database        *sql.DB
	frontendBaseURL string
	now             func() time.Time
}

func NewCreateMentorUseCase(
	staffRepo entities.StaffRepository,
	userRepo entities.UserRepository,
	teamRepo entities.TeamRepository,
	slackClient SlackUserInfoFetcher,
	database *sql.DB,
	frontendBaseURL string,
) *CreateMentorUseCase {
	return &CreateMentorUseCase{
		staffRepo:       staffRepo,
		userRepo:        userRepo,
		teamRepo:        teamRepo,
		slackClient:     slackClient,
		database:        database,
		frontendBaseURL: frontendBaseURL,
		now:             time.Now,
	}
}

func (uc *CreateMentorUseCase) Execute(ctx context.Context, callerSlackID, mentorSlackID, teamName string) (result *CreateMentorResult, err error) {
	defer func() {
		attrs := []slog.Attr{slog.String("caller_slack_id", callerSlackID), slog.String("mentor_slack_id", mentorSlackID), slog.String("team_name", teamName)}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
		slog.LogAttrs(ctx, slog.LevelDebug, "CreateMentorUseCase.Execute", attrs...)
	}()

	if strings.TrimSpace(callerSlackID) == "" {
		return nil, fmt.Errorf("caller slack ID is required")
	}
	if strings.TrimSpace(mentorSlackID) == "" {
		return nil, fmt.Errorf("mentor slack ID is required")
	}
	if strings.TrimSpace(teamName) == "" {
		return nil, fmt.Errorf("team name is required")
	}

	// Verify the caller is staff by fetching their email from Slack
	callerSlackUser, err := uc.slackClient.GetUserInfo(ctx, callerSlackID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch caller's slack profile: %w", err)
	}
	if callerSlackUser.Email == "" {
		return nil, fmt.Errorf("caller has no email address in Slack profile")
	}
	if _, err := uc.staffRepo.FindByEmail(ctx, callerSlackUser.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotStaff
		}
		return nil, fmt.Errorf("failed to verify staff status: %w", err)
	}

	// Find the team by name
	team, err := uc.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return nil, ErrTeamNotFound
	}

	// Fetch mentor's Slack profile
	slackUser, err := uc.slackClient.GetUserInfo(ctx, mentorSlackID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch slack user info: %w", err)
	}

	// Check if user already exists
	_, existErr := uc.userRepo.GetBySlackUserID(ctx, entities.SlackUserID(mentorSlackID))
	if existErr == nil {
		return nil, fmt.Errorf("user with Slack ID %s already exists", mentorSlackID)
	}
	if !errors.Is(existErr, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing user: %w", existErr)
	}

	name := slackUser.RealName
	if name == "" {
		name = mentorSlackID
	}

	email := slackUser.Email
	if email == "" {
		return nil, fmt.Errorf("slack user has no email address configured")
	}

	// Generate setup token before the transaction
	rawToken, err := generateSetupToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate setup token: %w", err)
	}
	tokenHash := hashToken(rawToken)
	expiresAt := uc.now().Add(setupTokenExpiry)

	teamID, err := uuid.Parse(string(team.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to parse team ID: %w", err)
	}

	// Execute all DB writes in a single transaction
	tx, err := uc.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := db.New(tx)

	// Create user
	userRow, err := qtx.CreateUser(ctx, db.CreateUserParams{
		SlackUserID:  mentorSlackID,
		Name:         name,
		Email:        email,
		Role:         string(entities.UserRoleMentor),
		PasswordHash: "",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create mentor user: %w", err)
	}

	// Create mentor record
	if err := qtx.CreateMentor(ctx, userRow.ID); err != nil {
		return nil, fmt.Errorf("failed to create mentor record: %w", err)
	}

	// Assign team
	if err := qtx.CreateMentorTeamAssignment(ctx, db.CreateMentorTeamAssignmentParams{
		MentorUserID: userRow.ID,
		TeamID:       teamID,
	}); err != nil {
		return nil, fmt.Errorf("failed to assign mentor to team: %w", err)
	}

	// Create setup token
	if _, err := qtx.CreateSetupToken(ctx, db.CreateSetupTokenParams{
		UserID:    userRow.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, fmt.Errorf("failed to store setup token: %w", err)
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

	setupURL := fmt.Sprintf("%s/setup?token=%s", strings.TrimRight(uc.frontendBaseURL, "/"), rawToken)

	return &CreateMentorResult{
		User:     *createdUser,
		SetupURL: setupURL,
	}, nil
}
