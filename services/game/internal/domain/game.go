package domain

import (
	"time"

	"github.com/gazizov-ai/online-checkers/services/game/internal/checkers"
	"github.com/google/uuid"
)

type GameStatus string

const (
	GameStatusActive   GameStatus = "active"
	GameStatusFinished GameStatus = "finished"
)

type GameResult string

const (
	GameResultWhiteWin GameResult = "white_win"
	GameResultBlackWin GameResult = "black_win"
	GameResultDraw     GameResult = "draw"
)

type FinishReason string

const (
	FinishReasonCheckersRules FinishReason = "checkers_rules"
	FinishReasonResignation   FinishReason = "resignation"
	FinishReasonDrawAgreement FinishReason = "draw_agreement"
)

type Game struct {
	ID            uuid.UUID
	WhitePlayerID uuid.UUID
	BlackPlayerID uuid.UUID
	Status        GameStatus
	WinnerID      *uuid.UUID
	Result        *GameResult
	FinishReason  *FinishReason
	DrawOfferBy   *uuid.UUID
	Snapshot      checkers.GameSnapshot
	CurrentTurn   checkers.Color
	CreatedAt     time.Time
	FinishedAt    *time.Time
}
