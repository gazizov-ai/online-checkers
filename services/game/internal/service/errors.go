package service

import "errors"

var (
	ErrPlayerNotInGame          = errors.New("player is not in game")
	ErrNotPlayersTurn           = errors.New("not player's turn")
	ErrGameNotActive            = errors.New("game is not active")
	ErrInvalidFinishedGame      = errors.New("invalid finished game")
	ErrNoDrawOffer              = errors.New("no draw offer")
	ErrDrawAlreadyOffered       = errors.New("draw already offered")
	ErrCannotAnswerOwnDrawOffer = errors.New("cannot answer own draw offer")
	ErrInvalidCursor            = errors.New("invalid cursor")
)
