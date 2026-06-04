package domain

import (
	"time"

	"github.com/gazizov-ai/online-checkers/services/game/internal/checkers"
	"github.com/google/uuid"
)

type GameStatus string

const (
	GameStatusActive   GameStatus = "active"
	GameStatusFinished GameStatus = "finished"
)

type Game struct {
	ID            uuid.UUID
	WhitePlayerID uuid.UUID
	BlackPlayerID uuid.UUID
	Status        GameStatus
	WinnerID      *uuid.UUID
	Snapshot      checkers.GameSnapshot
	CurrentTurn   checkers.Color
	CreatedAt     time.Time
	FinishedAt    *time.Time
}
