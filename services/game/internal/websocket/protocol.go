package websocket

import (
	"encoding/json"

	"github.com/gazizov-ai/online-checkers/services/game/internal/checkers"
)

const (
	MessageTypeMove         = "move"
	MessageTypeGameState    = "game_state"
	MessageTypeError        = "error"
	MessageTypeGameFinished = "game_finished"
)

type IncomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type MovePayload struct {
	From checkers.Position `json:"from"`
	To   checkers.Position `json:"to"`
}

type OutgoingMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
	Message string `json:"message,omitempty"`
}

type GameStatePayload struct {
	GameID      string                `json:"game_id"`
	BoardState  checkers.GameSnapshot `json:"board_state"`
	Status      string                `json:"status"`
	CurrentTurn string                `json:"current_turn"`
	WinnerID    *string               `json:"winner_id,omitempty"`
}

type GameFinishedPayload struct {
	GameID   string  `json:"game_id"`
	WinnerID *string `json:"winner_id,omitempty"`
}

func NewGameStateMessage(payload GameStatePayload) OutgoingMessage {
	return OutgoingMessage{
		Type:    MessageTypeGameState,
		Payload: payload,
	}
}

func NewErrorMessage(message string) OutgoingMessage {
	return OutgoingMessage{
		Type:    MessageTypeError,
		Message: message,
	}
}

func NewGameFinishedMessage(payload GameFinishedPayload) OutgoingMessage {
	return OutgoingMessage{
		Type:    MessageTypeGameFinished,
		Payload: payload,
	}
}
