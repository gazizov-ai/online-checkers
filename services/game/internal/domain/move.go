package domain

import (
	"time"

	"github.com/google/uuid"
)

type Move struct {
	ID         uuid.UUID
	GameID     uuid.UUID
	PlayerID   uuid.UUID
	MoveNumber int
	FromRow    int
	FromCol    int
	ToRow      int
	ToCol      int
	CreatedAt  time.Time
}
