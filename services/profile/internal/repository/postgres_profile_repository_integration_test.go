//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gazizov-ai/online-checkers/internal/testutil"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/domain"
	"github.com/google/uuid"
)

func TestPostgresProfileRepository(t *testing.T) {
	db := testutil.NewPostgresDB(t, "migrations/profile")
	repo := NewPostgresProfileRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	event := domain.UserRegisteredEvent{
		EventID:      uuid.New(),
		UserID:       uuid.New(),
		Username:     "alice",
		RegisteredAt: now,
	}
	if err := repo.CreateProfileFromUserRegistered(ctx, event); err != nil {
		t.Fatalf("CreateProfileFromUserRegistered() error = %v", err)
	}
	if err := repo.CreateProfileFromUserRegistered(ctx, event); err != nil {
		t.Fatalf("idempotent CreateProfileFromUserRegistered() error = %v", err)
	}

	profile, err := repo.GetProfileByUserID(ctx, event.UserID)
	if err != nil {
		t.Fatalf("GetProfileByUserID() error = %v", err)
	}
	if profile.Username != event.Username {
		t.Fatalf("profile = %+v", profile)
	}

	displayName := "Alice A."
	countryCode := "US"
	updated, err := repo.UpdateProfile(ctx, event.UserID, domain.UpdateProfileInput{
		DisplayName: &displayName,
		CountryCode: &countryCode,
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if updated.DisplayName == nil || *updated.DisplayName != displayName {
		t.Fatalf("updated profile = %+v", updated)
	}

	profiles, err := repo.GetProfilesByUserIDs(ctx, []uuid.UUID{event.UserID, uuid.New()})
	if err != nil {
		t.Fatalf("GetProfilesByUserIDs() error = %v", err)
	}
	if len(profiles) != 1 || profiles[0].UserID != event.UserID {
		t.Fatalf("profiles = %+v", profiles)
	}

	empty, err := repo.GetProfilesByUserIDs(ctx, nil)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty profiles = %+v, error = %v", empty, err)
	}

	if _, err := repo.GetProfileByUserID(ctx, uuid.New()); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("missing profile error = %v", err)
	}
	if _, err := repo.UpdateProfile(ctx, uuid.New(), domain.UpdateProfileInput{}); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("update missing profile error = %v", err)
	}
}
