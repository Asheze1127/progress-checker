package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/Asheze1127/progress-checker/backend/util"
)

const minPasswordLength = 8

var (
	ErrSetupTokenInvalid = errors.New("invalid setup token")
	ErrSetupTokenExpired = errors.New("setup token has expired")
	ErrSetupTokenUsed    = errors.New("setup token has already been used")
	ErrPasswordTooShort  = errors.New("password must be at least 8 characters")
)

type SetupPasswordUseCase struct {
	setupTokenRepo entities.SetupTokenRepository
	userRepo       entities.UserRepository
	hasher         *util.PasswordHasher
	now            func() time.Time
}

func NewSetupPasswordUseCase(
	setupTokenRepo entities.SetupTokenRepository,
	userRepo entities.UserRepository,
	hasher *util.PasswordHasher,
) *SetupPasswordUseCase {
	return &SetupPasswordUseCase{
		setupTokenRepo: setupTokenRepo,
		userRepo:       userRepo,
		hasher:         hasher,
		now:            time.Now,
	}
}

func (uc *SetupPasswordUseCase) Execute(ctx context.Context, rawToken, password string) error {
	if rawToken == "" {
		return ErrSetupTokenInvalid
	}
	if len(password) < minPasswordLength {
		return ErrPasswordTooShort
	}

	tokenHash := hashToken(rawToken)

	token, err := uc.setupTokenRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return ErrSetupTokenInvalid
	}

	if token.IsUsed() {
		return ErrSetupTokenUsed
	}

	if token.IsExpired(uc.now()) {
		return ErrSetupTokenExpired
	}

	hashed, err := uc.hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := uc.setupTokenRepo.MarkUsed(ctx, token.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	if err := uc.userRepo.UpdatePasswordHash(ctx, token.UserID, hashed); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
