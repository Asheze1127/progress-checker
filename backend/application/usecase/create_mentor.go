package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Asheze1127/progress-checker/backend/entities"
	slackinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/slack"
)

const setupTokenExpiry = 48 * time.Hour

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
	setupTokenRepo  entities.SetupTokenRepository
	mentorRepo      entities.MentorRepository
	slackClient     SlackUserInfoFetcher
	frontendBaseURL string
	now             func() time.Time
}

func NewCreateMentorUseCase(
	staffRepo entities.StaffRepository,
	userRepo entities.UserRepository,
	teamRepo entities.TeamRepository,
	setupTokenRepo entities.SetupTokenRepository,
	mentorRepo entities.MentorRepository,
	slackClient SlackUserInfoFetcher,
	frontendBaseURL string,
) *CreateMentorUseCase {
	return &CreateMentorUseCase{
		staffRepo:       staffRepo,
		userRepo:        userRepo,
		teamRepo:        teamRepo,
		setupTokenRepo:  setupTokenRepo,
		mentorRepo:      mentorRepo,
		slackClient:     slackClient,
		frontendBaseURL: frontendBaseURL,
		now:             time.Now,
	}
}

func (uc *CreateMentorUseCase) Execute(ctx context.Context, callerSlackID, mentorSlackID, teamName string) (*CreateMentorResult, error) {
	if strings.TrimSpace(callerSlackID) == "" {
		return nil, fmt.Errorf("caller slack ID is required")
	}
	if strings.TrimSpace(mentorSlackID) == "" {
		return nil, fmt.Errorf("mentor slack ID is required")
	}
	if strings.TrimSpace(teamName) == "" {
		return nil, fmt.Errorf("team name is required")
	}

	// Verify the caller is staff
	if _, err := uc.staffRepo.FindBySlackUserID(ctx, entities.SlackUserID(callerSlackID)); err != nil {
		return nil, fmt.Errorf("caller is not a registered staff member")
	}

	// Find the team by name
	team, err := uc.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("team %q not found", teamName)
	}

	// Fetch mentor's Slack profile
	slackUser, err := uc.slackClient.GetUserInfo(ctx, mentorSlackID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch slack user info: %w", err)
	}

	// Check if user already exists
	if existing, err := uc.userRepo.GetBySlackUserID(ctx, entities.SlackUserID(mentorSlackID)); err == nil && existing != nil {
		return nil, fmt.Errorf("user with Slack ID %s already exists", mentorSlackID)
	}

	name := slackUser.RealName
	if name == "" {
		name = mentorSlackID
	}

	email := slackUser.Email
	if email == "" {
		return nil, fmt.Errorf("slack user has no email address configured")
	}

	// Create the user with mentor role
	user := &entities.User{
		SlackUserID: entities.SlackUserID(mentorSlackID),
		Name:        name,
		Email:       email,
		Role:        entities.UserRoleMentor,
	}

	createdUser, err := uc.userRepo.Create(ctx, user, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create mentor user: %w", err)
	}

	// Create mentor record + team assignment
	if err := uc.mentorRepo.Create(ctx, createdUser.ID); err != nil {
		return nil, fmt.Errorf("failed to create mentor record: %w", err)
	}

	if err := uc.mentorRepo.AssignTeam(ctx, createdUser.ID, team.ID); err != nil {
		return nil, fmt.Errorf("failed to assign mentor to team: %w", err)
	}

	// Generate setup token
	rawToken, err := generateSetupToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate setup token: %w", err)
	}

	tokenHash := hashToken(rawToken)
	expiresAt := uc.now().Add(setupTokenExpiry)

	if _, err := uc.setupTokenRepo.Create(ctx, createdUser.ID, tokenHash, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to store setup token: %w", err)
	}

	setupURL := fmt.Sprintf("%s/setup?token=%s", strings.TrimRight(uc.frontendBaseURL, "/"), rawToken)

	return &CreateMentorResult{
		User:     *createdUser,
		SetupURL: setupURL,
	}, nil
}
