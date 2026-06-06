package websocket

import (
	"encoding/json"

	"github.com/gazizov-ai/online-checkers/services/game/internal/checkers"
)

const (
	MessageTypeMove         = "move"
	MessageTypeResign       = "resign"
	MessageTypeDrawOffer    = "draw_offer"
	MessageTypeDrawResponse = "draw_response"
	MessageTypeGameState    = "game_state"
	MessageTypeError        = "error"
	MessageTypeGameFinished = "game_finished"
	MessageTypeDrawOffered  = "draw_offered"
	MessageTypeDrawDeclined = "draw_declined"
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
	GameID      string                   `json:"game_id"`
	BoardState  checkers.GameSnapshot    `json:"board_state"`
	LegalMoves  []checkers.Move          `json:"legal_moves"`
	MoveHistory []MoveHistoryItemPayload `json:"move_history"`

	Status      string  `json:"status"`
	CurrentTurn string  `json:"current_turn"`
	WinnerID    *string `json:"winner_id,omitempty"`

	Result       *string `json:"result,omitempty"`
	FinishReason *string `json:"finish_reason,omitempty"`
	DrawOfferBy  *string `json:"draw_offer_by,omitempty"`
}

type MoveHistoryItemPayload struct {
	TurnNumber int    `json:"turn_number"`
	PlayerID   string `json:"player_id"`
	Notation   string `json:"notation"`
}

type GameFinishedPayload struct {
	GameID   string  `json:"game_id"`
	WinnerID *string `json:"winner_id,omitempty"`

	Result       *string `json:"result,omitempty"`
	FinishReason *string `json:"finish_reason,omitempty"`
}

type DrawResponsePayload struct {
	Accepted bool `json:"accepted"`
}

type DrawOfferedPayload struct {
	GameID    string `json:"game_id"`
	OfferedBy string `json:"offered_by"`
}

type DrawDeclinedPayload struct {
	GameID     string `json:"game_id"`
	DeclinedBy string `json:"declined_by"`
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

func NewDrawOfferedMessage(payload DrawOfferedPayload) OutgoingMessage {
	return OutgoingMessage{
		Type:    MessageTypeDrawOffered,
		Payload: payload,
	}
}

func NewDrawDeclinedMessage(payload DrawDeclinedPayload) OutgoingMessage {
	return OutgoingMessage{
		Type:    MessageTypeDrawDeclined,
		Payload: payload,
	}
}
