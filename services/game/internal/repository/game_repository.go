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
}
