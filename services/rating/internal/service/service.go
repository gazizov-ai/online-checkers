package service

import (
	"context"

	"github.com/gazizov-ai/online-checkers/services/rating/internal/domain"
	"github.com/gazizov-ai/online-checkers/services/rating/internal/repository"
	"github.com/google/uuid"
)

type RatingService struct {
	repo repository.RatingRepository
}

func NewRatingService(repo repository.RatingRepository) *RatingService {
	return &RatingService{
		repo: repo,
	}
}

func (s *RatingService) ProcessGameFinished(
	ctx context.Context,
	event domain.GameFinishedEvent,
) error {
	return s.repo.ProcessGameFinished(ctx, event)
}

func (s *RatingService) GetRating(
	ctx context.Context,
	userID uuid.UUID,
) (domain.Rating, error) {
	return s.repo.GetRating(ctx, userID)
}

func (s *RatingService) GetLeaderboard(
	ctx context.Context,
	limit int,
) ([]domain.Rating, error) {
	return s.repo.GetLeaderboard(ctx, limit)
}
