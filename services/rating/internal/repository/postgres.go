package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gazizov-ai/online-checkers/services/rating/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const ratingDelta = 25

type PostgresRatingRepository struct {
	db *sqlx.DB
}

func NewPostgresRatingRepository(db *sqlx.DB) *PostgresRatingRepository {
	return &PostgresRatingRepository{
		db: db,
	}
}

var _ RatingRepository = (*PostgresRatingRepository)(nil)

func (r *PostgresRatingRepository) GetRating(
	ctx context.Context,
	userID uuid.UUID,
) (domain.Rating, error) {
	var rating domain.Rating

	query := `
		SELECT user_id, rating, games_played, wins, losses, updated_at
		FROM ratings
		WHERE user_id = $1
	`

	err := r.db.GetContext(ctx, &rating, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Rating{}, domain.ErrRatingNotFound
		}

		return domain.Rating{}, fmt.Errorf("get rating: %w", err)
	}

	return rating, nil
}

func (r *PostgresRatingRepository) GetLeaderboard(
	ctx context.Context,
	limit int,
) ([]domain.Rating, error) {
	if limit <= 0 {
		limit = 10
	}

	var ratings []domain.Rating

	query := `
		SELECT user_id, rating, games_played, wins, losses, updated_at
		FROM ratings
		ORDER BY rating DESC, wins DESC, games_played ASC, updated_at DESC
		LIMIT $1
	`

	err := r.db.SelectContext(ctx, &ratings, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard: %w", err)
	}

	return ratings, nil
}

func (r *PostgresRatingRepository) ProcessGameFinished(
	ctx context.Context,
	event domain.GameFinishedEvent,
) error {
	if event.WhitePlayerID == event.BlackPlayerID {
		return domain.ErrInvalidWinner
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	insertIfNotProcessedYetQuery := `
		INSERT INTO processed_events (event_id, processed_at)
		VALUES ($1, now())
		ON CONFLICT (event_id) DO NOTHING
	`

	result, err := tx.ExecContext(ctx, insertIfNotProcessedYetQuery, event.EventID)
	if err != nil {
		return fmt.Errorf("insert processed event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get processed event rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil
	}

	createNewIfNotExistsQuery := `
		INSERT INTO ratings (user_id, rating, games_played, wins, losses, updated_at)
		VALUES
			($1, $3, 0, 0, 0, now()),
			($2, $3, 0, 0, 0, now())
		ON CONFLICT (user_id) DO NOTHING
	`

	_, err = tx.ExecContext(
		ctx,
		createNewIfNotExistsQuery,
		event.WhitePlayerID,
		event.BlackPlayerID,
		domain.DefaultRating,
	)
	if err != nil {
		return fmt.Errorf("ensure ratings: %w", err)
	}

	if event.IsDraw() {
		updateDrawQuery := `
			UPDATE ratings
			SET
				games_played = games_played + 1,
				updated_at = now()
			WHERE user_id IN ($1, $2)
		`

		if _, err := tx.ExecContext(ctx, updateDrawQuery, event.WhitePlayerID, event.BlackPlayerID); err != nil {
			return fmt.Errorf("update draw ratings: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit tx: %w", err)
		}

		return nil
	}

	if event.WinnerID == nil {
		return domain.ErrInvalidWinner
	}

	loserID, err := event.LoserID()
	if err != nil {
		return err
	}

	addWinQuery := `
		UPDATE ratings
		SET
			rating = rating + $2,
			games_played = games_played + 1,
			wins = wins + 1,
			updated_at = now()
		WHERE user_id = $1
	`

	_, err = tx.ExecContext(ctx, addWinQuery, *event.WinnerID, ratingDelta)
	if err != nil {
		return fmt.Errorf("update winner rating: %w", err)
	}

	addLossQuery := `
		UPDATE ratings
		SET 
			rating = GREATEST(0, rating - $2),
			games_played = games_played + 1,
			losses = losses + 1,
			updated_at = now()
		WHERE user_id = $1
	`

	_, err = tx.ExecContext(ctx, addLossQuery, loserID, ratingDelta)
	if err != nil {
		return fmt.Errorf("update loser rating: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
