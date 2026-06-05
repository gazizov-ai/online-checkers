package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gazizov-ai/online-checkers/services/matchmaking/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostgresQueueRepository struct {
	db *sqlx.DB
}

func NewPostgresQueueRepository(db *sqlx.DB) *PostgresQueueRepository {
	return &PostgresQueueRepository{db: db}
}

var _ QueueRepository = (*PostgresQueueRepository)(nil)

type queueRow struct {
	UserID uuid.UUID  `db:"user_id"`
	Status string     `db:"status"`
	GameID *uuid.UUID `db:"game_id"`
}

func (r queueRow) toDomain() domain.QueueEntry {
	return domain.QueueEntry{
		UserID: r.UserID,
		Status: domain.SearchStatus(r.Status),
		GameID: r.GameID,
	}
}

func (r *PostgresQueueRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.QueueEntry, error) {
	var row queueRow

	query := `
		SELECT user_id, status, game_id
		FROM matchmaking_queue
		WHERE user_id = $1
	`

	err := r.db.GetContext(ctx, &row, query, userID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	entry := row.toDomain()
	return &entry, nil
}

func (r *PostgresQueueRepository) EnqueueOrReservePair(
	ctx context.Context,
	userID uuid.UUID,
) (ReservationResult, error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return ReservationResult{}, err
	}
	defer tx.Rollback()

	var existing queueRow

	checkExistingQuery := `
		SELECT user_id, status, game_id 
		FROM matchmaking_queue
		WHERE user_id = $1
		FOR UPDATE
	`

	err = tx.GetContext(ctx, &existing, checkExistingQuery, userID)

	if err == nil {
		if err := tx.Commit(); err != nil {
			return ReservationResult{}, err
		}

		return ReservationResult{
			Status: domain.SearchStatus(existing.Status),
			GameID: existing.GameID,
		}, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return ReservationResult{}, err
	}

	var opponent queueRow

	opponentSelectQuery := `
		SELECT user_id, status, game_id
		FROM matchmaking_queue
		WHERE status = 'waiting'
			AND user_id <> $1
		ORDER BY created_at ASC
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`

	err = tx.GetContext(ctx, &opponent, opponentSelectQuery, userID)

	if errors.Is(err, sql.ErrNoRows) {
		enqueueQuery := `
			INSERT INTO matchmaking_queue (
				user_id,
				status,
				game_id,
				created_at,
				updated_at
			)
			VALUES($1, 'waiting', NULL, now(), now())
		`

		_, err := tx.ExecContext(ctx, enqueueQuery, userID)
		if err != nil {
			return ReservationResult{}, err
		}

		if err := tx.Commit(); err != nil {
			return ReservationResult{}, err
		}

		return ReservationResult{
			Status: domain.StatusWaiting,
		}, nil
	}

	if err != nil {
		return ReservationResult{}, err
	}

	updateToMatchingQuery := `
		UPDATE matchmaking_queue
		SET status = 'matching',
			updated_at = now()
		WHERE user_id = $1
	`

	_, err = tx.ExecContext(ctx, updateToMatchingQuery, opponent.UserID)
	if err != nil {
		return ReservationResult{}, err
	}

	reservePairQuery := `
		INSERT INTO matchmaking_queue (
			user_id,
			status,
			game_id,
			created_at,
			updated_at
		)
		VALUES ($1, 'matching', NULL, now(), now())
	`
	_, err = tx.ExecContext(ctx, reservePairQuery, userID)
	if err != nil {
		return ReservationResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return ReservationResult{}, err
	}

	opponentID := opponent.UserID

	return ReservationResult{
		Status:     domain.StatusMatching,
		OpponentID: &opponentID,
	}, nil
}

func (r *PostgresQueueRepository) CompleteMatch(
	ctx context.Context,
	userID uuid.UUID,
	opponentID uuid.UUID,
	gameID uuid.UUID,
) error {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	setMatchedQuery := `
		UPDATE matchmaking_queue
		SET status = 'matched',
			game_id = $3,
			updated_at = now()
		WHERE user_id IN ($1, $2)
			AND status = 'matching'
	`

	result, err := tx.ExecContext(ctx, setMatchedQuery, userID, opponentID, gameID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 2 {
		return fmt.Errorf("complete match: expected to update 2 rows, updated %d", rowsAffected)
	}

	return tx.Commit()
}

func (r *PostgresQueueRepository) ReleaseReservation(
	ctx context.Context,
	userID uuid.UUID,
	opponentID uuid.UUID,
) error {
	returnToWaitingQuery := `
		UPDATE matchmaking_queue
		SET status = 'waiting',
			game_id = NULL,
			updated_at = now()
		WHERE user_id IN ($1, $2)
			AND status = 'matching'
	`

	_, err := r.db.ExecContext(ctx, returnToWaitingQuery, userID, opponentID)

	return err
}

func (r *PostgresQueueRepository) Cancel(
	ctx context.Context,
	userID uuid.UUID,
) error {
	cancelQueueingQuery := `
		DELETE FROM matchmaking_queue
		WHERE user_id = $1
			AND status = 'waiting'
	`

	_, err := r.db.ExecContext(ctx, cancelQueueingQuery, userID)

	return err
}

func (r *PostgresQueueRepository) CleanupStaleReservations(
	ctx context.Context,
	olderThan time.Duration,
) error {
	query := `
		UPDATE matchmaking_queue
		SET status = 'waiting',
			game_id = NULL,
			updated_at = now()
		WHERE status = 'matching'
			AND updated_at < now() - ($1 * interval '1 second')
	`

	_, err := r.db.ExecContext(ctx, query, int(olderThan.Seconds()))
	return err
}
