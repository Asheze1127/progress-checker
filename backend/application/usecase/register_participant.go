package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/Asheze1127/progress-checker/backend/api/middleware"
	"github.com/Asheze1127/progress-checker/backend/entities"
	slackinfra "github.com/Asheze1127/progress-checker/backend/infrastructure/slack"
)

type RegisterParticipantUseCase struct {
	userRepo        entities.UserRepository
	teamRepo        entities.TeamRepository
	mentorRepo      entities.MentorRepository
	participantRepo entities.ParticipantRepository
	slackClient     SlackUserInfoFetcher
}

// SlackUserInfoFetcher is already defined in create_mentor.go but we need it here too.
// Since both are in the same package, the interface is shared.

func NewRegisterParticipantUseCase(
	userRepo entities.UserRepository,
	teamRepo entities.TeamRepository,
	mentorRepo entities.MentorRepository,
	participantRepo entities.ParticipantRepository,
	slackClient SlackUserInfoFetcher,
) *RegisterParticipantUseCase {
	return &RegisterParticipantUseCase{
		userRepo:        userRepo,
		teamRepo:        teamRepo,
		mentorRepo:      mentorRepo,
		participantRepo: participantRepo,
		slackClient:     slackClient,
	}
}

func (uc *RegisterParticipantUseCase) Execute(ctx context.Context, slackUserID, teamID string) (*entities.User, error) {
	if strings.TrimSpace(slackUserID) == "" {
		return nil, fmt.Errorf("slack_user_id is required")
	}
	if strings.TrimSpace(teamID) == "" {
		return nil, fmt.Errorf("team_id is required")
	}

	// Get the authenticated mentor from context
	mentorUser := middleware.UserFromContext(ctx)
	if mentorUser == nil {
		return nil, fmt.Errorf("not authorized: authentication required")
	}

	// Verify the mentor is assigned to this team
	mentorTeamIDs, err := uc.mentorRepo.GetTeamIDs(ctx, mentorUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentor teams: %w", err)
	}

	authorized := false
	for _, tid := range mentorTeamIDs {
		if string(tid) == teamID {
			authorized = true
			break
		}
	}
	if !authorized {
		return nil, fmt.Errorf("not authorized for this team")
	}

	// Verify team exists
	if _, err := uc.teamRepo.GetByID(ctx, entities.TeamID(teamID)); err != nil {
		return nil, fmt.Errorf("team not found")
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
		email = slackUserID + "@slack.local"
	}

	// Create participant user (no password needed)
	user := &entities.User{
		SlackUserID: entities.SlackUserID(slackUserID),
		Name:        name,
		Email:       email,
		Role:        entities.UserRoleParticipant,
	}

	createdUser, err := uc.userRepo.Create(ctx, user, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create participant: %w", err)
	}

	if err := uc.participantRepo.Create(ctx, createdUser.ID, entities.TeamID(teamID)); err != nil {
		return nil, fmt.Errorf("failed to create participant record: %w", err)
	}

	return createdUser, nil
}
