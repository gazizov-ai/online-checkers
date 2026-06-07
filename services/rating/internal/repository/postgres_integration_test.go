//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gazizov-ai/online-checkers/internal/testutil"
	"github.com/gazizov-ai/online-checkers/services/rating/internal/domain"
	"github.com/google/uuid"
)

func TestPostgresRatingRepositoryProcessGameFinished(t *testing.T) {
	db := testutil.NewPostgresDB(t, "migrations/rating")
	repo := NewPostgresRatingRepository(db)
	ctx := context.Background()
	white := uuid.New()
	black := uuid.New()
	winner := white
	event := domain.GameFinishedEvent{
		EventID:       uuid.New(),
		GameID:        uuid.New(),
		WhitePlayerID: white,
		BlackPlayerID: black,
		WinnerID:      &winner,
		Result:        domain.GameResultWhiteWin,
		Reason:        domain.FinishReasonCheckersRules,
		FinishedAt:    time.Now(),
	}

	if err := repo.ProcessGameFinished(ctx, event); err != nil {
		t.Fatalf("ProcessGameFinished() error = %v", err)
	}
	if err := repo.ProcessGameFinished(ctx, event); err != nil {
		t.Fatalf("idempotent ProcessGameFinished() error = %v", err)
	}

	whiteRating, _ := repo.GetRating(ctx, white)
	blackRating, _ := repo.GetRating(ctx, black)
	if whiteRating.Rating != domain.DefaultRating+ratingDelta || whiteRating.Wins != 1 || whiteRating.GamesPlayed != 1 {
		t.Fatalf("white rating = %+v", whiteRating)
	}
	if blackRating.Rating != domain.DefaultRating-ratingDelta || blackRating.Losses != 1 || blackRating.GamesPlayed != 1 {
		t.Fatalf("black rating = %+v", blackRating)
	}

	leaderboard, err := repo.GetLeaderboard(ctx, 0)
	if err != nil {
		t.Fatalf("GetLeaderboard() error = %v", err)
	}
	if len(leaderboard) != 2 || leaderboard[0].UserID != white {
		t.Fatalf("leaderboard = %+v", leaderboard)
	}

	draw := event
	draw.EventID = uuid.New()
	draw.WinnerID = nil
	draw.Result = domain.GameResultDraw
	if err := repo.ProcessGameFinished(ctx, draw); err != nil {
		t.Fatalf("process draw: %v", err)
	}
	whiteRating, _ = repo.GetRating(ctx, white)
	if whiteRating.GamesPlayed != 2 || whiteRating.Rating != domain.DefaultRating+ratingDelta {
		t.Fatalf("white rating after draw = %+v", whiteRating)
	}
}

func TestPostgresRatingRepositoryValidation(t *testing.T) {
	db := testutil.NewPostgresDB(t, "migrations/rating")
	repo := NewPostgresRatingRepository(db)
	ctx := context.Background()
	player := uuid.New()

	if err := repo.ProcessGameFinished(ctx, domain.GameFinishedEvent{
		EventID:       uuid.New(),
		WhitePlayerID: player,
		BlackPlayerID: player,
		Result:        domain.GameResultDraw,
	}); !errors.Is(err, domain.ErrInvalidWinner) {
		t.Fatalf("same-player error = %v", err)
	}

	if _, err := repo.GetRating(ctx, uuid.New()); !errors.Is(err, domain.ErrRatingNotFound) {
		t.Fatalf("missing rating error = %v", err)
	}
}
