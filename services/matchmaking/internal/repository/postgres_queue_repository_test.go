package repository

import (
	"testing"

	"github.com/gazizov-ai/online-checkers/services/matchmaking/internal/domain"
	"github.com/google/uuid"
)

func TestQueueRowToDomain(t *testing.T) {
	gameID := uuid.New()
	row := queueRow{UserID: uuid.New(), Status: string(domain.StatusMatched), GameID: &gameID}

	got := row.toDomain()
	if got.UserID != row.UserID || got.Status != domain.StatusMatched || got.GameID != row.GameID {
		t.Fatalf("toDomain() = %+v", got)
	}
}
