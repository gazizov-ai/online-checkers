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
	Result        domain.GameResult
	Reason        domain.FinishReason
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

type ResignInput struct {
	GameID   uuid.UUID
	PlayerID uuid.UUID
}

type ResignOutput struct {
	Game domain.Game
}

type OfferDrawInput struct {
	GameID   uuid.UUID
	PlayerID uuid.UUID
}

type OfferDrawOutput struct {
	Game domain.Game
}

type RespondDrawInput struct {
	GameID   uuid.UUID
	PlayerID uuid.UUID
	Accepted bool
}

type RespondDrawOutput struct {
	Game domain.Game
}

type GetMoveHistoryInput struct {
	GameID uuid.UUID
}

type GetMoveHistoryOutput struct {
	Items []domain.MoveHistoryItem
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

func winnerIDForColor(game domain.Game, color checkers.Color) (uuid.UUID, error) {
	switch color {
	case checkers.White:
		return game.WhitePlayerID, nil
	case checkers.Black:
		return game.BlackPlayerID, nil
	default:
		return uuid.Nil, ErrPlayerNotInGame
	}
}

func nextTurnPosition(lastMove *domain.Move, snapshot checkers.GameSnapshot) (turnNumber int, sequenceNumber int) {
	if lastMove == nil {
		return 1, 1
	}

	if snapshot.ForcedPiece != nil {
		return lastMove.TurnNumber, lastMove.SequenceNumber + 1
	}

	return lastMove.TurnNumber + 1, 1
}

func checkersMoveFromDomain(move domain.Move) checkers.Move {
	return checkers.Move{
		From: checkers.Position{
			Row: move.FromRow,
			Col: move.FromCol,
		},
		To: checkers.Position{
			Row: move.ToRow,
			Col: move.ToCol,
		},
	}
}

func notationForMoves(moves []domain.Move) string {
	if len(moves) == 0 {
		return ""
	}

	checkersMoves := make([]checkers.Move, 0, len(moves))

	hasCapture := false
	for _, move := range moves {
		checkersMoves = append(checkersMoves, checkersMoveFromDomain(move))

		if move.IsCapture {
			hasCapture = true
		}
	}

	return checkers.MoveChainNotation(checkersMoves, hasCapture)
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

func (s *GameService) publishGameFinishedIfNeeded(ctx context.Context, game domain.Game) error {
	if game.FinishedAt == nil || s.gameFinishedPublisher == nil {
		return nil
	}

	if game.Result == nil || game.FinishReason == nil {
		return ErrInvalidFinishedGame
	}

	if *game.Result != domain.GameResultDraw && game.WinnerID == nil {
		return ErrInvalidFinishedGame
	}

	var winnerID uuid.UUID
	if game.WinnerID != nil {
		winnerID = *game.WinnerID
	}

	event := GameFinishedEvent{
		EventID:       uuid.New(),
		GameID:        game.ID,
		WhitePlayerID: game.WhitePlayerID,
		BlackPlayerID: game.BlackPlayerID,
		WinnerID:      winnerID,
		FinishedAt:    *game.FinishedAt,
		Result:        *game.Result,
		Reason:        *game.FinishReason,
	}

	return s.gameFinishedPublisher.PublishGameFinished(ctx, event)
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

	oldSnapshot := game.Snapshot

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

	lastMove, err := s.repo.LastMove(ctx, game.ID)
	if err != nil {
		return ApplyMoveOutput{}, err
	}

	turnNumber, sequenceNumber := nextTurnPosition(lastMove, oldSnapshot)

	moveRecord := domain.Move{
		ID:             uuid.New(),
		GameID:         game.ID,
		PlayerID:       input.PlayerID,
		MoveNumber:     nextMoveNumber,
		TurnNumber:     turnNumber,
		SequenceNumber: sequenceNumber,
		FromRow:        input.From.Row,
		FromCol:        input.From.Col,
		ToRow:          input.To.Row,
		ToCol:          input.To.Col,
		IsCapture:      result.Captured,
		Notation:       checkers.MoveSegmentNotation(move, result.Captured),
	}

	if err := s.repo.CreateMove(ctx, moveRecord); err != nil {
		return ApplyMoveOutput{}, err
	}

	turnMoves, err := s.repo.ListMovesByTurn(ctx, game.ID, turnNumber)
	if err != nil {
		return ApplyMoveOutput{}, err
	}

	notation := notationForMoves(turnMoves)

	if err := s.repo.UpdateTurnNotation(ctx, game.ID, turnNumber, notation); err != nil {
		return ApplyMoveOutput{}, err
	}

	snapshot := engine.Snapshot()

	game.Snapshot = snapshot
	game.CurrentTurn = snapshot.Turn
	game.Status = domain.GameStatus(snapshot.Status)

	if game.DrawOfferBy != nil {
		game.DrawOfferBy = nil
	}

	if result.GameFinished {
		if result.Winner == nil {
			return ApplyMoveOutput{}, ErrInvalidFinishedGame
		}

		winnerID, err := winnerIDForColor(game, *result.Winner)
		if err != nil {
			return ApplyMoveOutput{}, err
		}

		gameResult, err := gameResultForWinner(game, winnerID)
		if err != nil {
			return ApplyMoveOutput{}, err
		}

		now := time.Now().UTC()

		finishGame(
			&game,
			gameResult,
			domain.FinishReasonCheckersRules,
			&winnerID,
			result.Winner,
			now,
		)
	}

	if err := s.repo.SaveGameState(ctx, game); err != nil {
		return ApplyMoveOutput{}, err
	}

	if result.GameFinished {
		if err := s.publishGameFinishedIfNeeded(ctx, game); err != nil {
			return ApplyMoveOutput{}, err
		}
	}

	return ApplyMoveOutput{
		Game:   game,
		Result: result,
	}, nil
}

func (s *GameService) Resign(ctx context.Context, input ResignInput) (ResignOutput, error) {
	game, err := s.repo.GetGame(ctx, input.GameID)
	if err != nil {
		return ResignOutput{}, err
	}

	color, ok := playerColor(game, input.PlayerID)
	if !ok {
		return ResignOutput{}, ErrPlayerNotInGame
	}

	if game.Status != domain.GameStatusActive {
		return ResignOutput{}, ErrGameNotActive
	}

	var winnerColor checkers.Color
	var winnerID uuid.UUID

	switch color {
	case checkers.White:
		winnerColor = checkers.Black
		winnerID = game.BlackPlayerID
	case checkers.Black:
		winnerColor = checkers.White
		winnerID = game.WhitePlayerID
	default:
		return ResignOutput{}, ErrPlayerNotInGame
	}

	gameResult, err := gameResultForWinner(game, winnerID)
	if err != nil {
		return ResignOutput{}, err
	}

	now := time.Now().UTC()

	finishGame(
		&game,
		gameResult,
		domain.FinishReasonResignation,
		&winnerID,
		&winnerColor,
		now,
	)

	if err := s.repo.SaveGameState(ctx, game); err != nil {
		return ResignOutput{}, err
	}

	if err := s.publishGameFinishedIfNeeded(ctx, game); err != nil {
		return ResignOutput{}, err
	}

	return ResignOutput{
		Game: game,
	}, nil
}

func (s *GameService) OfferDraw(ctx context.Context, input OfferDrawInput) (OfferDrawOutput, error) {
	game, err := s.repo.GetGame(ctx, input.GameID)
	if err != nil {
		return OfferDrawOutput{}, err
	}

	if _, ok := playerColor(game, input.PlayerID); !ok {
		return OfferDrawOutput{}, ErrPlayerNotInGame
	}

	if game.Status != domain.GameStatusActive {
		return OfferDrawOutput{}, ErrGameNotActive
	}

	if game.DrawOfferBy != nil {
		return OfferDrawOutput{}, ErrDrawAlreadyOffered
	}

	offerBy := input.PlayerID
	game.DrawOfferBy = &offerBy

	if err := s.repo.SaveGameState(ctx, game); err != nil {
		return OfferDrawOutput{}, err
	}

	return OfferDrawOutput{Game: game}, nil
}

func (s *GameService) RespondDraw(ctx context.Context, input RespondDrawInput) (RespondDrawOutput, error) {
	game, err := s.repo.GetGame(ctx, input.GameID)
	if err != nil {
		return RespondDrawOutput{}, err
	}

	if _, ok := playerColor(game, input.PlayerID); !ok {
		return RespondDrawOutput{}, ErrPlayerNotInGame
	}

	if game.Status != domain.GameStatusActive {
		return RespondDrawOutput{}, ErrGameNotActive
	}

	if game.DrawOfferBy == nil {
		return RespondDrawOutput{}, ErrNoDrawOffer
	}

	if *game.DrawOfferBy == input.PlayerID {
		return RespondDrawOutput{}, ErrCannotAnswerOwnDrawOffer
	}

	if !input.Accepted {
		game.DrawOfferBy = nil

		if err := s.repo.SaveGameState(ctx, game); err != nil {
			return RespondDrawOutput{}, err
		}

		return RespondDrawOutput{Game: game}, nil
	}

	now := time.Now().UTC()

	finishGame(
		&game,
		domain.GameResultDraw,
		domain.FinishReasonDrawAgreement,
		nil,
		nil,
		now,
	)

	if err := s.repo.SaveGameState(ctx, game); err != nil {
		return RespondDrawOutput{}, err
	}

	if err := s.publishGameFinishedIfNeeded(ctx, game); err != nil {
		return RespondDrawOutput{}, err
	}

	return RespondDrawOutput{Game: game}, nil
}

func finishGame(
	game *domain.Game,
	result domain.GameResult,
	reason domain.FinishReason,
	winnerID *uuid.UUID,
	winnerColor *checkers.Color,
	finishedAt time.Time,
) {
	game.Status = domain.GameStatusFinished
	game.Result = &result
	game.FinishReason = &reason
	game.WinnerID = winnerID
	game.FinishedAt = &finishedAt
	game.DrawOfferBy = nil

	snapshot := game.Snapshot
	snapshot.Status = checkers.StatusFinished
	snapshot.Winner = winnerColor
	snapshot.ForcedPiece = nil

	game.Snapshot = snapshot
}

func gameResultForWinner(game domain.Game, winnerID uuid.UUID) (domain.GameResult, error) {
	switch winnerID {
	case game.WhitePlayerID:
		return domain.GameResultWhiteWin, nil
	case game.BlackPlayerID:
		return domain.GameResultBlackWin, nil
	default:
		return "", ErrPlayerNotInGame
	}
}

func (s *GameService) GetMoveHistory(
	ctx context.Context,
	input GetMoveHistoryInput,
) (GetMoveHistoryOutput, error) {
	items, err := s.repo.ListMoveHistory(ctx, input.GameID)
	if err != nil {
		return GetMoveHistoryOutput{}, err
	}

	return GetMoveHistoryOutput{
		Items: items,
	}, nil
}
