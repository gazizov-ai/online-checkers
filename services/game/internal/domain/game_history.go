package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserGameHistoryItem struct {
	GameID        uuid.UUID
	WhitePlayerID uuid.UUID
	BlackPlayerID uuid.UUID

	UserColor string

	Status       string
	Result       *string
	FinishReason *string
	WinnerID     *uuid.UUID

	CreatedAt  time.Time
	FinishedAt *time.Time
}

type UserGameHistoryPage struct {
	Items      []UserGameHistoryItem
	NextCursor *string
}
