package domain

import "github.com/google/uuid"

type SearchStatus string

const (
	StatusWaiting  SearchStatus = "waiting"
	StatusMatching SearchStatus = "matching"
	StatusMatched  SearchStatus = "matched"
)

type QueueEntry struct {
	UserID uuid.UUID
	Status SearchStatus
	GameID *uuid.UUID
}
