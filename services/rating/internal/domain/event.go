package domain

import (
	"time"

	"github.com/google/uuid"
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

type GameFinishedEvent struct {
	EventID       uuid.UUID
	GameID        uuid.UUID
	WhitePlayerID uuid.UUID
	BlackPlayerID uuid.UUID

	WinnerID *uuid.UUID

	Result GameResult
	Reason FinishReason

	FinishedAt time.Time
}

func (e GameFinishedEvent) IsDraw() bool {
	return e.Result == GameResultDraw
}

func (e GameFinishedEvent) LoserID() (uuid.UUID, error) {
	if e.WinnerID == nil {
		return uuid.Nil, ErrInvalidWinner
	}

	if *e.WinnerID == e.WhitePlayerID {
		return e.BlackPlayerID, nil
	}

	if *e.WinnerID == e.BlackPlayerID {
		return e.WhitePlayerID, nil
	}

	return uuid.Nil, ErrInvalidWinner
}
