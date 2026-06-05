package domain

import "errors"

var (
	ErrInvalidWinner  = errors.New("invalid winner")
	ErrRatingNotFound = errors.New("rating not found")
)
