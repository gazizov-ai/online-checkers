package repository

import (
	"context"

	"github.com/gazizov-ai/online-checkers/services/game/internal/domain"
	"github.com/google/uuid"
)

type GameRepository interface {
	CreateGame(ctx context.Context, game domain.Game) error
	GetGame(ctx context.Context, id uuid.UUID) (domain.Game, error)

	SaveGameState(ctx context.Context, game domain.Game) error

	CreateMove(ctx context.Context, move domain.Move) error
	NextMoveNumber(ctx context.Context, gameID uuid.UUID) (int, error)
	LastMove(ctx context.Context, gameID uuid.UUID) (*domain.Move, error)
	ListMovesByTurn(ctx context.Context, gameID uuid.UUID, turnNumber int) ([]domain.Move, error)
	UpdateTurnNotation(ctx context.Context, gameID uuid.UUID, turnNumber int, notation string) error

	ListMoveHistory(ctx context.Context, gameID uuid.UUID) ([]domain.MoveHistoryItem, error)
	ListGamesByUser(
		ctx context.Context,
		userID uuid.UUID,
		includeActive bool,
		limit int,
		offset int,
	) ([]domain.UserGameHistoryItem, error)
}
