package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProfileFromRow(t *testing.T) {
	displayName := "Alice"
	row := profileRow{
		UserID:      uuid.New(),
		Username:    "alice",
		DisplayName: &displayName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	got := profileFromRow(row)
	if got.UserID != row.UserID || got.Username != row.Username || got.DisplayName != row.DisplayName {
		t.Fatalf("profileFromRow() = %+v", got)
	}
}
