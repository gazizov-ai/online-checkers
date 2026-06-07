//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/gazizov-ai/online-checkers/internal/testutil"
	"github.com/gazizov-ai/online-checkers/services/matchmaking/internal/domain"
	"github.com/google/uuid"
)

func TestPostgresQueueRepositoryMatchLifecycle(t *testing.T) {
	db := testutil.NewPostgresDB(t, "migrations/matchmaking")
	repo := NewPostgresQueueRepository(db)
	ctx := context.Background()
	first := uuid.New()
	second := uuid.New()

	result, err := repo.EnqueueOrReservePair(ctx, first)
	if err != nil {
		t.Fatalf("enqueue first: %v", err)
	}
	if result.Status != domain.StatusWaiting || result.OpponentID != nil {
		t.Fatalf("first result = %+v", result)
	}

	repeated, err := repo.EnqueueOrReservePair(ctx, first)
	if err != nil || repeated.Status != domain.StatusWaiting {
		t.Fatalf("repeated enqueue = %+v, error = %v", repeated, err)
	}

	result, err = repo.EnqueueOrReservePair(ctx, second)
	if err != nil {
		t.Fatalf("reserve pair: %v", err)
	}
	if result.Status != domain.StatusMatching || result.OpponentID == nil || *result.OpponentID != first {
		t.Fatalf("pair result = %+v", result)
	}

	gameID := uuid.New()
	if err := repo.CompleteMatch(ctx, second, first, gameID); err != nil {
		t.Fatalf("CompleteMatch() error = %v", err)
	}

	repeatedMatch, err := repo.EnqueueOrReservePair(ctx, second)
	if err != nil {
		t.Fatalf("repeat matched EnqueueOrReservePair() error = %v", err)
	}
	if repeatedMatch.Status != domain.StatusWaiting || repeatedMatch.GameID != nil {
		t.Fatalf("repeated matched result = %+v", repeatedMatch)
	}

	firstEntry, err := repo.GetByUserID(ctx, first)
	if err != nil {
		t.Fatalf("GetByUserID(%s) error = %v", first, err)
	}
	if firstEntry == nil ||
		firstEntry.Status != domain.StatusMatched ||
		firstEntry.GameID == nil ||
		*firstEntry.GameID != gameID {
		t.Fatalf("matched entry = %+v", firstEntry)
	}

	secondEntry, err := repo.GetByUserID(ctx, second)
	if err != nil {
		t.Fatalf("GetByUserID(%s) error = %v", second, err)
	}
	if secondEntry == nil || secondEntry.Status != domain.StatusWaiting || secondEntry.GameID != nil {
		t.Fatalf("new search entry = %+v", secondEntry)
	}

	if err := repo.Cancel(ctx, first); err != nil {
		t.Fatalf("Cancel(matched) error = %v", err)
	}

	entry, err := repo.GetByUserID(ctx, first)
	if err != nil {
		t.Fatalf("GetByUserID(%s) error = %v", first, err)
	}
	if entry != nil {
		t.Fatalf("Cancel should remove matched entry, got %+v", entry)
	}
}

func TestPostgresQueueRepositoryReleaseCancelAndCleanup(t *testing.T) {
	db := testutil.NewPostgresDB(t, "migrations/matchmaking")
	repo := NewPostgresQueueRepository(db)
	ctx := context.Background()
	first := uuid.New()
	second := uuid.New()

	_, _ = repo.EnqueueOrReservePair(ctx, first)
	_, _ = repo.EnqueueOrReservePair(ctx, second)
	if err := repo.ReleaseReservation(ctx, first, second); err != nil {
		t.Fatalf("ReleaseReservation() error = %v", err)
	}
	for _, userID := range []uuid.UUID{first, second} {
		entry, _ := repo.GetByUserID(ctx, userID)
		if entry == nil || entry.Status != domain.StatusWaiting {
			t.Fatalf("released entry = %+v", entry)
		}
	}

	if _, err := db.Exec(`
		UPDATE matchmaking_queue
		SET status = 'matching', updated_at = now() - interval '10 minutes'
		WHERE user_id IN ($1, $2)
	`, first, second); err != nil {
		t.Fatalf("age reservations: %v", err)
	}
	if err := repo.CleanupStaleReservations(ctx, time.Minute); err != nil {
		t.Fatalf("CleanupStaleReservations() error = %v", err)
	}

	if err := repo.Cancel(ctx, first); err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	entry, err := repo.GetByUserID(ctx, first)
	if err != nil || entry != nil {
		t.Fatalf("cancelled entry = %+v, error = %v", entry, err)
	}
}

func TestPostgresQueueRepositoryCompleteRequiresPair(t *testing.T) {
	db := testutil.NewPostgresDB(t, "migrations/matchmaking")
	repo := NewPostgresQueueRepository(db)
	ctx := context.Background()
	first := uuid.New()

	_, _ = repo.EnqueueOrReservePair(ctx, first)
	if err := repo.CompleteMatch(ctx, first, uuid.New(), uuid.New()); err == nil {
		t.Fatal("expected CompleteMatch() to reject an incomplete reservation")
	}
}
