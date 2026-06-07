package repository

import (
	"testing"
	"time"

	"github.com/gazizov-ai/online-checkers/services/game/internal/domain"
	"github.com/google/uuid"
)

func TestRepositoryRowMappings(t *testing.T) {
	result := string(domain.GameResultWhiteWin)
	reason := string(domain.FinishReasonResignation)
	if got := gameResultFromString(&result); got == nil || *got != domain.GameResultWhiteWin {
		t.Fatalf("gameResultFromString() = %v", got)
	}
	if got := finishReasonFromString(&reason); got == nil || *got != domain.FinishReasonResignation {
		t.Fatalf("finishReasonFromString() = %v", got)
	}
	if gameResultFromString(nil) != nil || finishReasonFromString(nil) != nil {
		t.Fatal("nil database values must remain nil")
	}

	row := moveRow{ID: uuid.New(), GameID: uuid.New(), MoveNumber: 2, CreatedAt: time.Now()}
	got := moveRowToDomain(row)
	if got.ID != row.ID || got.GameID != row.GameID || got.MoveNumber != row.MoveNumber {
		t.Fatalf("moveRowToDomain() = %+v", got)
	}
}
