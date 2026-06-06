package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/gazizov-ai/online-checkers/services/profile/internal/domain"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/repository"
	"github.com/google/uuid"
)

type ProfileService struct {
	repo repository.ProfileRepository
}

func NewProfileService(repo repository.ProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

func normalizeUpdateProfileInput(input domain.UpdateProfileInput) (domain.UpdateProfileInput, error) {
	if input.DisplayName != nil {
		value := strings.TrimSpace(*input.DisplayName)
		if value == "" {
			input.DisplayName = nil
		} else {
			if utf8.RuneCountInString(value) > 40 {
				return domain.UpdateProfileInput{}, ErrInvalidDisplayName
			}
			input.DisplayName = &value
		}
	}

	if input.CountryCode != nil {
		value := strings.ToUpper(strings.TrimSpace(*input.CountryCode))
		if value == "" {
			input.CountryCode = nil
		} else {
			if len(value) != 2 || value[0] < 'A' || value[0] > 'Z' || value[1] < 'A' || value[1] > 'Z' {
				return domain.UpdateProfileInput{}, ErrInvalidCountryCode
			}
			input.CountryCode = &value
		}
	}

	if input.AvatarURL != nil {
		value := strings.TrimSpace(*input.AvatarURL)
		if value == "" {
			input.AvatarURL = nil
		} else {
			parsed, err := url.ParseRequestURI(value)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				return domain.UpdateProfileInput{}, ErrInvalidAvatarURL
			}
			if parsed.Scheme != "http" && parsed.Scheme != "https" {
				return domain.UpdateProfileInput{}, ErrInvalidAvatarURL
			}
			if len(value) > 500 {
				return domain.UpdateProfileInput{}, ErrInvalidAvatarURL
			}
			input.AvatarURL = &value
		}
	}

	if input.Bio != nil {
		value := strings.TrimSpace(*input.Bio)
		if value == "" {
			input.Bio = nil
		} else {
			if utf8.RuneCountInString(value) > 300 {
				return domain.UpdateProfileInput{}, ErrInvalidBio
			}
			input.Bio = &value
		}
	}

	return input, nil
}

func (s *ProfileService) GetProfile(
	ctx context.Context,
	userID uuid.UUID,
) (domain.Profile, error) {
	return s.repo.GetProfileByUserID(ctx, userID)
}

func (s *ProfileService) UpdateProfile(
	ctx context.Context,
	userID uuid.UUID,
	input domain.UpdateProfileInput,
) (domain.Profile, error) {
	normalized, err := normalizeUpdateProfileInput(input)
	if err != nil {
		return domain.Profile{}, err
	}

	return s.repo.UpdateProfile(ctx, userID, normalized)
}

func (s *ProfileService) GetProfiles(
	ctx context.Context,
	userIDs []uuid.UUID,
) ([]domain.Profile, error) {
	if len(userIDs) > 100 {
		return nil, ErrTooManyProfilesRequested
	}

	uniqueIDs := make([]uuid.UUID, 0, len(userIDs))
	seen := make(map[uuid.UUID]struct{}, len(userIDs))

	for _, id := range userIDs {
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}

	return s.repo.GetProfilesByUserIDs(ctx, uniqueIDs)
}

func (s *ProfileService) HandleUserRegistered(
	ctx context.Context,
	event domain.UserRegisteredEvent,
) error {
	if event.EventID == uuid.Nil {
		return errors.New("event_id is required")
	}

	if event.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}

	if strings.TrimSpace(event.Username) == "" {
		return errors.New("username is required")
	}

	if event.RegisteredAt.IsZero() {
		return errors.New("registered_at is required")
	}

	return s.repo.CreateProfileFromUserRegistered(ctx, event)
}
