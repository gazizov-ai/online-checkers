package checkers

import "errors"

var (
	ErrGameFinished        = errors.New("game is finished")
	ErrInvalidMove         = errors.New("invalid move")
	ErrPieceNotFound       = errors.New("piece not found")
	ErrWrongPieceColor     = errors.New("wrong piece color")
	ErrDestinationOccupied = errors.New("destination occupied")
	ErrCaptureRequired     = errors.New("capture is required")
	ErrMustContinueCapture = errors.New("must continue capture")
	ErrInvalidPosition     = errors.New("invalid position")
	ErrInvalidState        = errors.New("invalid game state")
)
