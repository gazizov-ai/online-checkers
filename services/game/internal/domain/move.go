package domain

import (
	"time"

	"github.com/google/uuid"
)

type Move struct {
	ID       uuid.UUID
	GameID   uuid.UUID
	PlayerID uuid.UUID

	MoveNumber     int
	TurnNumber     int
	SequenceNumber int

	FromRow int
	FromCol int
	ToRow   int
	ToCol   int

	IsCapture bool
	Notation  string

	CreatedAt time.Time
}

type MoveHistoryItem struct {
	TurnNumber int
	PlayerID   uuid.UUID
	Notation   string
	CreatedAt  time.Time
}
