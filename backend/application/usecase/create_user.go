package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Asheze1127/progress-checker/backend/entities"
)

const setupTokenExpiry = 48 * time.Hour

type CreateUserResult struct {
	User     entities.User
	SetupURL string
}

type CreateUserUseCase struct {
	userRepo        entities.UserRepository
	teamRepo        entities.TeamRepository
	setupTokenRepo  entities.SetupTokenRepository
	participantRepo entities.ParticipantRepository
	mentorRepo      entities.MentorRepository
	frontendBaseURL string
	now             func() time.Time
}

func NewCreateUserUseCase(
	userRepo entities.UserRepository,
	teamRepo entities.TeamRepository,
	setupTokenRepo entities.SetupTokenRepository,
	participantRepo entities.ParticipantRepository,
	mentorRepo entities.MentorRepository,
	frontendBaseURL string,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo:        userRepo,
		teamRepo:        teamRepo,
		setupTokenRepo:  setupTokenRepo,
		participantRepo: participantRepo,
		mentorRepo:      mentorRepo,
		frontendBaseURL: frontendBaseURL,
		now:             time.Now,
	}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, slackUserID, name, email string, role entities.UserRole, teamID entities.TeamID) (*CreateUserResult, error) {
	if strings.TrimSpace(slackUserID) == "" {
		return nil, fmt.Errorf("slack_user_id is required")
	}
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(email) == "" {
		return nil, fmt.Errorf("email is required")
	}
	if role != entities.UserRoleParticipant && role != entities.UserRoleMentor {
		return nil, fmt.Errorf("role must be participant or mentor")
	}
	if strings.TrimSpace(string(teamID)) == "" {
		return nil, fmt.Errorf("team_id is required")
	}

	if _, err := uc.teamRepo.GetByID(ctx, teamID); err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	user := &entities.User{
		SlackUserID: entities.SlackUserID(slackUserID),
		Name:        name,
		Email:       email,
		Role:        role,
	}

	createdUser, err := uc.userRepo.Create(ctx, user, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if role == entities.UserRoleParticipant {
		if err := uc.participantRepo.Create(ctx, createdUser.ID, teamID); err != nil {
			return nil, fmt.Errorf("failed to create participant: %w", err)
		}
	} else {
		if err := uc.mentorRepo.Create(ctx, createdUser.ID); err != nil {
			return nil, fmt.Errorf("failed to create mentor: %w", err)
		}
	}

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

	return &CreateUserResult{
		User:     *createdUser,
		SetupURL: setupURL,
	}, nil
}

