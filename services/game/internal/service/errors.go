package service

import "errors"

var (
	ErrPlayerNotInGame = errors.New("player is not in game")
	ErrNotPlayersTurn  = errors.New("not player's turn")
)
