package service

import "errors"

var (
	ErrInvalidDisplayName       = errors.New("invalid display name")
	ErrInvalidCountryCode       = errors.New("invalid country code")
	ErrInvalidAvatarURL         = errors.New("invalid avatar url")
	ErrInvalidBio               = errors.New("invalid bio")
	ErrTooManyProfilesRequested = errors.New("too many profiles requested")
)
