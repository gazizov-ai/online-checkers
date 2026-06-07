package repository

import (
	"context"
	"time"

	"github.com/gazizov-ai/online-checkers/services/matchmaking/internal/domain"
	"github.com/google/uuid"
)

type ReservationResult struct {
	Status     domain.SearchStatus
	GameID     *uuid.UUID
	OpponentID *uuid.UUID
}

type QueueRepository interface {
	EnqueueOrReservePair(ctx context.Context, userID uuid.UUID) (ReservationResult, error)
	CompleteMatch(ctx context.Context, userID uuid.UUID, opponentID uuid.UUID, gameID uuid.UUID) error
	ReleaseReservation(ctx context.Context, userID uuid.UUID, opponentID uuid.UUID) error
	CleanupStaleReservations(ctx context.Context, olderThan time.Duration) error
	ConsumeMatchedByUserID(ctx context.Context, userID uuid.UUID) (*domain.QueueEntry, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.QueueEntry, error)
	Cancel(ctx context.Context, userID uuid.UUID) error
}
