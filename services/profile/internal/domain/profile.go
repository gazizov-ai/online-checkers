package domain

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	UserID      uuid.UUID
	Username    string
	DisplayName *string
	CountryCode *string
	AvatarURL   *string
	Bio         *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UpdateProfileInput struct {
	DisplayName *string
	CountryCode *string
	AvatarURL   *string
	Bio         *string
}
