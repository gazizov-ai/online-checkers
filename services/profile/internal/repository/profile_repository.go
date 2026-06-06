package repository

import (
	"context"
	"errors"

	"github.com/gazizov-ai/online-checkers/services/profile/internal/domain"
	"github.com/google/uuid"
)

var ErrProfileNotFound = errors.New("profile not found")

type ProfileRepository interface {
	GetProfileByUserID(ctx context.Context, userID uuid.UUID) (domain.Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, input domain.UpdateProfileInput) (domain.Profile, error)
	GetProfilesByUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]domain.Profile, error)

	CreateProfileFromUserRegistered(ctx context.Context, event domain.UserRegisteredEvent) error
}
