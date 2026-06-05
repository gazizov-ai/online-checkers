package domain

import (
	"time"

	"github.com/google/uuid"
)

const DefaultRating = 1000

type Rating struct {
	UserID      uuid.UUID `db:"user_id"`
	Rating      int       `db:"rating"`
	GamesPlayed int       `db:"games_played"`
	Wins        int       `db:"wins"`
	Losses      int       `db:"losses"`
	UpdatedAt   time.Time `db:"updated_at"`
}
