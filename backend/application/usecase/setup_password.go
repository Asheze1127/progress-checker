package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"

	db "github.com/Asheze1127/progress-checker/backend/database/postgres/generated"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/Asheze1127/progress-checker/backend/util"
	"github.com/google/uuid"
)

const minPasswordLength = 12
const maxPasswordLength = 72

var (
	ErrSetupTokenInvalid = errors.New("invalid setup token")
	ErrSetupTokenExpired = errors.New("setup token has expired")
	ErrSetupTokenUsed    = errors.New("setup token has already been used")
	ErrPasswordTooShort  = errors.New("password must be at least 12 characters")
	ErrPasswordTooLong   = errors.New("password must be at most 72 characters")
	ErrPasswordTooWeak   = errors.New("password must contain uppercase, lowercase, digit, and special character")
)

type SetupPasswordUseCase struct {
	setupTokenRepo entities.SetupTokenRepository
	userRepo       entities.UserRepository
	hasher         *util.PasswordHasher
	database       *sql.DB
	now            func() time.Time
}

func NewSetupPasswordUseCase(
	setupTokenRepo entities.SetupTokenRepository,
	userRepo entities.UserRepository,
	hasher *util.PasswordHasher,
	database *sql.DB,
) *SetupPasswordUseCase {
	return &SetupPasswordUseCase{
		setupTokenRepo: setupTokenRepo,
		userRepo:       userRepo,
		hasher:         hasher,
		database:       database,
		now:            time.Now,
	}
}

func (uc *SetupPasswordUseCase) Execute(ctx context.Context, rawToken, password string) (err error) {
	defer func() {
		attrs := []slog.Attr{slog.Bool("has_token", rawToken != "")}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
		slog.LogAttrs(ctx, slog.LevelDebug, "SetupPasswordUseCase.Execute", attrs...)
	}()

	if rawToken == "" {
		return ErrSetupTokenInvalid
	}
	if len([]rune(password)) < minPasswordLength {
		return ErrPasswordTooShort
	}
	if len([]rune(password)) > maxPasswordLength {
		return ErrPasswordTooLong
	}
	if !isPasswordComplex(password) {
		return ErrPasswordTooWeak
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

	// Run password update and token mark-used in a single transaction
	tx, err := uc.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := db.New(tx)

	userID, err := uuid.Parse(string(token.UserID))
	if err != nil {
		return fmt.Errorf("failed to parse user ID: %w", err)
	}
	tokenID, err := uuid.Parse(string(token.ID))
	if err != nil {
		return fmt.Errorf("failed to parse token ID: %w", err)
	}

	if err := qtx.UpdateUserPasswordHash(ctx, db.UpdateUserPasswordHashParams{
		PasswordHash: hashed,
		ID:           userID,
	}); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err := qtx.MarkSetupTokenUsed(ctx, tokenID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

const allowedSpecialChars = "!@#$%^&*()_+-=[]{}|;:',.<>?/~`\"\\"

func isPasswordComplex(password string) bool {
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune(allowedSpecialChars, r):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}
