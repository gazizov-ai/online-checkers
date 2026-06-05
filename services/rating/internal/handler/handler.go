package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	"github.com/gazizov-ai/online-checkers/services/rating/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RatingService interface {
	GetRating(ctx context.Context, userID uuid.UUID) (domain.Rating, error)
	GetLeaderboard(ctx context.Context, limit int) ([]domain.Rating, error)
}

type Handler struct {
	ratingService RatingService
}

func NewHandler(ratingService RatingService) *Handler {
	return &Handler{
		ratingService: ratingService,
	}
}

type RatingResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	Rating      int       `json:"rating"`
	GamesPlayed int       `json:"games_played"`
	Wins        int       `json:"wins"`
	Losses      int       `json:"losses"`
}

type LeaderboardResponse struct {
	Items []RatingResponse `json:"items"`
}

func toRatingResponse(rating domain.Rating) RatingResponse {
	return RatingResponse{
		UserID:      rating.UserID,
		Rating:      rating.Rating,
		GamesPlayed: rating.GamesPlayed,
		Wins:        rating.Wins,
		Losses:      rating.Losses,
	}
}

func (h *Handler) GetRating(w http.ResponseWriter, r *http.Request) {
	rawUserID := chi.URLParam(r, "user_id")

	userID, err := uuid.Parse(rawUserID)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid user_id")
		return
	}

	rating, err := h.ratingService.GetRating(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrRatingNotFound) {
			_ = httpx.WriteError(w, http.StatusNotFound, "not_found", "rating not found")
			return
		}

		_ = httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, toRatingResponse(rating))
}

func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit := 10

	rawLimit := r.URL.Query().Get("limit")
	if rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit <= 0 {
			_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid limit")
			return
		}

		if parsedLimit > 100 {
			parsedLimit = 100
		}

		limit = parsedLimit
	}

	ratings, err := h.ratingService.GetLeaderboard(r.Context(), limit)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}

	items := make([]RatingResponse, 0, len(ratings))
	for _, rating := range ratings {
		items = append(items, toRatingResponse(rating))
	}

	_ = httpx.WriteJSON(w, http.StatusOK, LeaderboardResponse{
		Items: items,
	})
}
