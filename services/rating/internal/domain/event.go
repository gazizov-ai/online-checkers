package domain

import (
	"time"

	"github.com/google/uuid"
)

type GameFinishedEvent struct {
	EventID       uuid.UUID
	GameID        uuid.UUID
	WhitePlayerID uuid.UUID
	BlackPlayerID uuid.UUID
	WinnerID      uuid.UUID
	FinishedAt    time.Time
}

func (e GameFinishedEvent) LoserID() (uuid.UUID, error) {
	if e.WinnerID == e.WhitePlayerID {
		return e.BlackPlayerID, nil
	}

	if e.WinnerID == e.BlackPlayerID {
		return e.WhitePlayerID, nil
	}

	return uuid.Nil, ErrInvalidWinner
}
