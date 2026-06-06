package checkers

import "fmt"

func SquareName(pos Position) string {
	file := byte('a' + pos.Col)
	rank := 8 - pos.Row

	return fmt.Sprintf("%c%d", file, rank)
}

func MoveSegmentNotation(move Move, captured bool) string {
	sep := "-"
	if captured {
		sep = ":"
	}

	return SquareName(move.From) + sep + SquareName(move.To)
}

func MoveChainNotation(moves []Move, captured bool) string {
	if len(moves) == 0 {
		return ""
	}

	sep := "-"
	if captured {
		sep = ":"
	}

	notation := SquareName(moves[0].From)

	for _, move := range moves {
		notation += sep + SquareName(move.To)
	}

	return notation
}
