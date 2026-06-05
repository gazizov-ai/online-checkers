package service

import (
	"context"
	"time"

	"github.com/gazizov-ai/online-checkers/services/game/internal/checkers"
	"github.com/gazizov-ai/online-checkers/services/game/internal/domain"
	"github.com/gazizov-ai/online-checkers/services/game/internal/repository"
	"github.com/google/uuid"
)

type GameService struct {
	repo                  repository.GameRepository
	gameFinishedPublisher GameFinishedPublisher
}

func NewGameService(repo repository.GameRepository, gameFinishedPublisher GameFinishedPublisher) *GameService {
	return &GameService{
		repo:                  repo,
		gameFinishedPublisher: gameFinishedPublisher,
	}
}

type GameFinishedPublisher interface {
	PublishGameFinished(ctx context.Context, event GameFinishedEvent) error
}

type GameFinishedEvent struct {
	EventID       uuid.UUID
	GameID        uuid.UUID
	WhitePlayerID uuid.UUID
	BlackPlayerID uuid.UUID
	WinnerID      uuid.UUID
	FinishedAt    time.Time
}

type CreateGameInput struct {
	WhitePlayerID uuid.UUID
	BlackPlayerID uuid.UUID
}

type ApplyMoveInput struct {
	GameID   uuid.UUID
	PlayerID uuid.UUID
	From     checkers.Position
	To       checkers.Position
}

type ApplyMoveOutput struct {
	Game   domain.Game
	Result checkers.MoveResult
}

func playerColor(game domain.Game, playerID uuid.UUID) (checkers.Color, bool) {
	switch playerID {
	case game.WhitePlayerID:
		return checkers.White, true
	case game.BlackPlayerID:
		return checkers.Black, true
	default:
		return "", false
	}
}

func (s *GameService) CreateGame(ctx context.Context, input CreateGameInput) (uuid.UUID, error) {
	gameEngine := checkers.NewGame()
	snapshot := gameEngine.Snapshot()

	game := domain.Game{
		ID:            uuid.New(),
		WhitePlayerID: input.WhitePlayerID,
		BlackPlayerID: input.BlackPlayerID,
		Status:        domain.GameStatusActive,
		WinnerID:      nil,
		Snapshot:      snapshot,
		CurrentTurn:   snapshot.Turn,
	}

	if err := s.repo.CreateGame(ctx, game); err != nil {
		return uuid.Nil, err
	}

	return game.ID, nil
}

func (s *GameService) GetGame(ctx context.Context, id uuid.UUID) (domain.Game, error) {
	return s.repo.GetGame(ctx, id)
}

func (s *GameService) ApplyMove(ctx context.Context, input ApplyMoveInput) (ApplyMoveOutput, error) {
	game, err := s.repo.GetGame(ctx, input.GameID)
	if err != nil {
		return ApplyMoveOutput{}, err
	}

	color, ok := playerColor(game, input.PlayerID)
	if !ok {
		return ApplyMoveOutput{}, ErrPlayerNotInGame
	}

	if color != game.CurrentTurn {
		return ApplyMoveOutput{}, ErrNotPlayersTurn
	}

	engine, err := checkers.NewGameFromSnapshot(game.Snapshot)
	if err != nil {
		return ApplyMoveOutput{}, err
	}

	move := checkers.Move{
		From: input.From,
		To:   input.To,
	}

	result, err := engine.ApplyMove(move)
	if err != nil {
		return ApplyMoveOutput{}, err
	}

	nextMoveNumber, err := s.repo.NextMoveNumber(ctx, game.ID)
	if err != nil {
		return ApplyMoveOutput{}, err
	}

	moveRecord := domain.Move{
		ID:         uuid.New(),
		GameID:     game.ID,
		PlayerID:   input.PlayerID,
		MoveNumber: nextMoveNumber,
		FromRow:    input.From.Row,
		FromCol:    input.From.Col,
		ToRow:      input.To.Row,
		ToCol:      input.To.Col,
	}

	if err := s.repo.CreateMove(ctx, moveRecord); err != nil {
		return ApplyMoveOutput{}, err
	}

	snapshot := engine.Snapshot()

	game.Snapshot = snapshot
	game.CurrentTurn = snapshot.Turn
	game.Status = domain.GameStatus(snapshot.Status)

	if snapshot.Winner != nil {
		if *snapshot.Winner == checkers.White {
			winnerID := game.WhitePlayerID
			game.WinnerID = &winnerID
		} else {
			winnerID := game.BlackPlayerID
			game.WinnerID = &winnerID
		}
	}

	if snapshot.Status == checkers.StatusFinished {
		now := time.Now().UTC()
		game.FinishedAt = &now
	}

	if err := s.repo.SaveGameState(ctx, game); err != nil {
		return ApplyMoveOutput{}, err
	}

	if result.GameFinished && game.WinnerID != nil && game.FinishedAt != nil && s.gameFinishedPublisher != nil {
		event := GameFinishedEvent{
			EventID:       uuid.New(),
			GameID:        game.ID,
			WhitePlayerID: game.WhitePlayerID,
			BlackPlayerID: game.BlackPlayerID,
			WinnerID:      *game.WinnerID,
			FinishedAt:    *game.FinishedAt,
		}

		if err := s.gameFinishedPublisher.PublishGameFinished(ctx, event); err != nil {
			return ApplyMoveOutput{}, err
		}
	}

	return ApplyMoveOutput{
		Game:   game,
		Result: result,
	}, nil
}
