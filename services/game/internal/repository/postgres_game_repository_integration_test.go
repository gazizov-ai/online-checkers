//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/gazizov-ai/online-checkers/internal/testutil"
	"github.com/gazizov-ai/online-checkers/services/game/internal/checkers"
	"github.com/gazizov-ai/online-checkers/services/game/internal/domain"
	"github.com/google/uuid"
)

func TestPostgresGameRepositoryGameLifecycle(t *testing.T) {
	db := testutil.NewPostgresDB(t, "migrations/game")
	repo := NewPostgresGameRepository(db)
	ctx := context.Background()
	white := uuid.New()
	black := uuid.New()
	game := domain.Game{
		ID:            uuid.New(),
		WhitePlayerID: white,
		BlackPlayerID: black,
		Status:        domain.GameStatusActive,
		Snapshot: checkers.GameSnapshot{
			Board:  checkers.NewInitialBoard(),
			Turn:   checkers.White,
			Status: checkers.StatusActive,
		},
		CurrentTurn: checkers.White,
	}

	if err := repo.CreateGame(ctx, game); err != nil {
		t.Fatalf("CreateGame() error = %v", err)
	}
	stored, err := repo.GetGame(ctx, game.ID)
	if err != nil {
		t.Fatalf("GetGame() error = %v", err)
	}
	if stored.ID != game.ID || stored.Snapshot.Board.CountPieces(checkers.White) != 12 {
		t.Fatalf("stored game = %+v", stored)
	}

	finishedAt := time.Now().UTC().Truncate(time.Microsecond)
	result := domain.GameResultWhiteWin
	reason := domain.FinishReasonResignation
	game.Status = domain.GameStatusFinished
	game.WinnerID = &white
	game.Result = &result
	game.FinishReason = &reason
	game.FinishedAt = &finishedAt
	game.Snapshot.Status = checkers.StatusFinished
	game.Snapshot.Winner = colorPtr(checkers.White)
	if err := repo.SaveGameState(ctx, game); err != nil {
		t.Fatalf("SaveGameState() error = %v", err)
	}

	stored, _ = repo.GetGame(ctx, game.ID)
	if stored.Result == nil || *stored.Result != result || stored.WinnerID == nil || *stored.WinnerID != white {
		t.Fatalf("finished game = %+v", stored)
	}

	finished, err := repo.ListGamesByUser(ctx, white, false, 10, 0)
	if err != nil || len(finished) != 1 || finished[0].UserColor != "white" {
		t.Fatalf("finished games = %+v, error = %v", finished, err)
	}
}

func TestPostgresGameRepositoryMoves(t *testing.T) {
	db := testutil.NewPostgresDB(t, "migrations/game")
	repo := NewPostgresGameRepository(db)
	ctx := context.Background()
	game := domain.Game{
		ID:            uuid.New(),
		WhitePlayerID: uuid.New(),
		BlackPlayerID: uuid.New(),
		Status:        domain.GameStatusActive,
		Snapshot: checkers.GameSnapshot{
			Board:  checkers.NewInitialBoard(),
			Turn:   checkers.White,
			Status: checkers.StatusActive,
		},
		CurrentTurn: checkers.White,
	}
	if err := repo.CreateGame(ctx, game); err != nil {
		t.Fatalf("CreateGame() error = %v", err)
	}

	if next, err := repo.NextMoveNumber(ctx, game.ID); err != nil || next != 1 {
		t.Fatalf("initial NextMoveNumber() = %d, error = %v", next, err)
	}
	if last, err := repo.LastMove(ctx, game.ID); err != nil || last != nil {
		t.Fatalf("initial LastMove() = %+v, error = %v", last, err)
	}

	moves := []domain.Move{
		{ID: uuid.New(), GameID: game.ID, PlayerID: game.WhitePlayerID, MoveNumber: 1, TurnNumber: 1, SequenceNumber: 1, FromRow: 5, FromCol: 0, ToRow: 4, ToCol: 1},
		{ID: uuid.New(), GameID: game.ID, PlayerID: game.WhitePlayerID, MoveNumber: 2, TurnNumber: 1, SequenceNumber: 2, FromRow: 4, FromCol: 1, ToRow: 2, ToCol: 3, IsCapture: true},
	}
	for _, move := range moves {
		if err := repo.CreateMove(ctx, move); err != nil {
			t.Fatalf("CreateMove() error = %v", err)
		}
	}
	if err := repo.UpdateTurnNotation(ctx, game.ID, 1, "a3:c5"); err != nil {
		t.Fatalf("UpdateTurnNotation() error = %v", err)
	}

	if next, _ := repo.NextMoveNumber(ctx, game.ID); next != 3 {
		t.Fatalf("NextMoveNumber() = %d", next)
	}
	last, _ := repo.LastMove(ctx, game.ID)
	if last == nil || last.ID != moves[1].ID {
		t.Fatalf("LastMove() = %+v", last)
	}
	turnMoves, _ := repo.ListMovesByTurn(ctx, game.ID, 1)
	if len(turnMoves) != 2 || turnMoves[0].SequenceNumber != 1 || turnMoves[1].Notation != "a3:c5" {
		t.Fatalf("ListMovesByTurn() = %+v", turnMoves)
	}
	history, _ := repo.ListMoveHistory(ctx, game.ID)
	if len(history) != 1 || history[0].Notation != "a3:c5" {
		t.Fatalf("ListMoveHistory() = %+v", history)
	}
}

func colorPtr(color checkers.Color) *checkers.Color {
	return &color
}
