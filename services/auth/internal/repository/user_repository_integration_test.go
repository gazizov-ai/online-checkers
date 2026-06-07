//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gazizov-ai/online-checkers/internal/testutil"
	"github.com/gazizov-ai/online-checkers/services/auth/internal/domain"
	"github.com/google/uuid"
)

func TestPostgresUserRepository(t *testing.T) {
	db := testutil.NewPostgresDB(t, "migrations/auth")
	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	email := "alice@example.com"
	user := domain.User{
		ID:           uuid.New(),
		Username:     "alice",
		Email:        &email,
		PasswordHash: "hash",
		CreatedAt:    time.Now().UTC().Truncate(time.Microsecond),
	}
	event := domain.OutboxEvent{
		ID:            uuid.New(),
		EventType:     domain.EventTypeUserRegistered,
		AggregateType: domain.OutboxAggregateUser,
		AggregateID:   user.ID,
		Topic:         "user.registered",
		KafkaKey:      user.ID.String(),
		Payload:       []byte("payload"),
		Headers:       map[string]string{"event_type": domain.EventTypeUserRegistered},
		CreatedAt:     user.CreatedAt,
	}

	if err := repo.CreateUserWithOutboxEvent(ctx, user, event); err != nil {
		t.Fatalf("CreateUserWithOutboxEvent() error = %v", err)
	}

	byID, err := repo.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUserByID() error = %v", err)
	}
	byUsername, err := repo.GetUserByUsername(ctx, user.Username)
	if err != nil {
		t.Fatalf("GetUserByUsername() error = %v", err)
	}
	if byID.ID != user.ID || byUsername.Email == nil || *byUsername.Email != email {
		t.Fatalf("unexpected users: byID=%+v byUsername=%+v", byID, byUsername)
	}

	pending, err := repo.ListPendingOutboxEvents(ctx, 10)
	if err != nil {
		t.Fatalf("ListPendingOutboxEvents() error = %v", err)
	}
	if len(pending) != 1 || pending[0].ID != event.ID {
		t.Fatalf("pending events = %+v", pending)
	}

	if err := repo.MarkOutboxEventFailedAttempt(ctx, event.ID, "broker unavailable"); err != nil {
		t.Fatalf("MarkOutboxEventFailedAttempt() error = %v", err)
	}
	pending, _ = repo.ListPendingOutboxEvents(ctx, 10)
	if pending[0].Attempts != 1 || pending[0].LastError == nil {
		t.Fatalf("failed event not updated: %+v", pending[0])
	}

	if err := repo.MarkOutboxEventPublished(ctx, event.ID); err != nil {
		t.Fatalf("MarkOutboxEventPublished() error = %v", err)
	}
	pending, _ = repo.ListPendingOutboxEvents(ctx, 10)
	if len(pending) != 0 {
		t.Fatalf("published event still pending: %+v", pending)
	}

	duplicateUsername := user
	duplicateUsername.ID = uuid.New()
	duplicateUsername.Email = nil
	if err := repo.CreateUser(ctx, duplicateUsername); !errors.Is(err, ErrUsernameAlreadyExists) {
		t.Fatalf("duplicate username error = %v", err)
	}

	duplicateEmail := user
	duplicateEmail.ID = uuid.New()
	duplicateEmail.Username = "other"
	if err := repo.CreateUser(ctx, duplicateEmail); !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("duplicate email error = %v", err)
	}

	if _, err := repo.GetUserByID(ctx, uuid.New()); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("missing user error = %v", err)
	}
}
