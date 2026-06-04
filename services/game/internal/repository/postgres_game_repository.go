package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gazizov-ai/online-checkers/services/game/internal/checkers"
	"github.com/gazizov-ai/online-checkers/services/game/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostgresGameRepository struct {
	db *sqlx.DB
}

func NewPostgresGameRepository(db *sqlx.DB) *PostgresGameRepository {
	return &PostgresGameRepository{db: db}
}

type gameRow struct {
	ID            uuid.UUID  `db:"id"`
	WhitePlayerID uuid.UUID  `db:"white_player_id"`
	BlackPlayerID uuid.UUID  `db:"black_player_id"`
	Status        string     `db:"status"`
	WinnerID      *uuid.UUID `db:"winner_id"`
	BoardState    []byte     `db:"board_state"`
	CurrentTurn   string     `db:"current_turn"`
	CreatedAt     time.Time  `db:"created_at"`
	FinishedAt    *time.Time `db:"finished_at"`
}

func (r *PostgresGameRepository) CreateGame(ctx context.Context, game domain.Game) error {
	boardState, err := json.Marshal(game.Snapshot)
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO games (
			id,
			white_player_id,
			black_player_id,
			status,
			winner_id,
			board_state,
			current_turn
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		game.ID,
		game.WhitePlayerID,
		game.BlackPlayerID,
		game.Status,
		game.WinnerID,
		boardState,
		game.CurrentTurn,
	)

	return err
}

func (r *PostgresGameRepository) GetGame(ctx context.Context, id uuid.UUID) (domain.Game, error) {
	const query = `
		SELECT 
			id,
			white_player_id,
			black_player_id,
			status,
			winner_id,
			board_state,
			current_turn,
			created_at,
			finished_at
		FROM games
		WHERE id = $1
	`

	var row gameRow
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		return domain.Game{}, err
	}

	var snapshot checkers.GameSnapshot
	if err := json.Unmarshal(row.BoardState, &snapshot); err != nil {
		return domain.Game{}, err
	}

	return domain.Game{
		ID:            row.ID,
		WhitePlayerID: row.WhitePlayerID,
		BlackPlayerID: row.BlackPlayerID,
		Status:        domain.GameStatus(row.Status),
		WinnerID:      row.WinnerID,
		Snapshot:      snapshot,
		CurrentTurn:   checkers.Color(row.CurrentTurn),
		CreatedAt:     row.CreatedAt,
		FinishedAt:    row.FinishedAt,
	}, nil
}

func (r *PostgresGameRepository) SaveGameState(ctx context.Context, game domain.Game) error {
	boardState, err := json.Marshal(game.Snapshot)
	if err != nil {
		return err
	}

	const query = `
		UPDATE games
		SET
			status = $2,
			winner_id = $3,
			board_state = $4,
			current_turn = $5,
			finished_at = $6
		WHERE id = $1
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		game.ID,
		game.Status,
		game.WinnerID,
		boardState,
		game.CurrentTurn,
		game.FinishedAt,
	)

	return err
}

func (r *PostgresGameRepository) NextMoveNumber(ctx context.Context, gameID uuid.UUID) (int, error) {
	const query = `
		SELECT COALESCE(MAX(move_number), 0) + 1
		FROM moves
		WHERE game_id = $1
	`

	var next int
	if err := r.db.GetContext(ctx, &next, query, gameID); err != nil {
		return 0, err
	}

	return next, nil
}

func (r *PostgresGameRepository) CreateMove(ctx context.Context, move domain.Move) error {
	const query = `
		INSERT INTO moves (
			id,
			game_id,
			player_id,
			move_number,
			from_row,
			from_col,
			to_row,
			to_col
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		move.ID,
		move.GameID,
		move.PlayerID,
		move.MoveNumber,
		move.FromRow,
		move.FromCol,
		move.ToRow,
		move.ToCol,
	)

	return err
}

var _ GameRepository = (*PostgresGameRepository)(nil)
