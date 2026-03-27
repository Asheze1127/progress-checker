package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Asheze1127/progress-checker/backend/application/service/jwt"
	"github.com/Asheze1127/progress-checker/backend/entities"
	"github.com/Asheze1127/progress-checker/backend/util"
)

var ErrInvalidStaffCredentials = errors.New("invalid email or password")

type StaffLoginResult struct {
	Token string
	Staff entities.Staff
}

type StaffLoginUseCase struct {
	staffRepo entities.StaffRepository
	jwt       *jwt.JWTService
	hasher    *util.PasswordHasher
}

func NewStaffLoginUseCase(
	staffRepo entities.StaffRepository,
	jwt *jwt.JWTService,
	hasher *util.PasswordHasher,
) *StaffLoginUseCase {
	return &StaffLoginUseCase{staffRepo: staffRepo, jwt: jwt, hasher: hasher}
}

func (uc *StaffLoginUseCase) Execute(ctx context.Context, email, password string) (*StaffLoginResult, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	staffWithPw, err := uc.staffRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidStaffCredentials
	}

	if err := uc.hasher.Verify(staffWithPw.PasswordHash, password); err != nil {
		return nil, ErrInvalidStaffCredentials
	}

	token, err := uc.jwt.GenerateStaffToken(&staffWithPw.Staff)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &StaffLoginResult{
		Token: token,
		Staff: staffWithPw.Staff,
	}, nil
}
