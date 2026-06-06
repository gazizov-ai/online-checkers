package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRegisteredEvent struct {
	EventID      uuid.UUID
	UserID       uuid.UUID
	Username     string
	RegisteredAt time.Time
}
