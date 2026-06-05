package repository

import (
	"context"

	"github.com/gazizov-ai/online-checkers/services/rating/internal/domain"
	"github.com/google/uuid"
)

type RatingRepository interface {
	ProcessGameFinished(ctx context.Context, event domain.GameFinishedEvent) error
	GetRating(ctx context.Context, userID uuid.UUID) (domain.Rating, error)
	GetLeaderboard(ctx context.Context, limit int) ([]domain.Rating, error)
}
