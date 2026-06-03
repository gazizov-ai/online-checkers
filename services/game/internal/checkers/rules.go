package checkers

func (g *Game) LegalMoves() []Move {
	if g == nil {
		return nil
	}

	if g.Status == StatusFinished {
		return nil
	}

	if g.ForcedPiece != nil {
		return findCapturesFrom(g, *g.ForcedPiece)
	}

	captures := findAllCaptures(g, g.Turn)
	if len(captures) > 0 {
		return captures
	}

	return findAllSimpleMoves(g, g.Turn)
}

func (g *Game) LegalMovesFrom(pos Position) []Move {
	if g == nil {
		return nil
	}

	if g.Status == StatusFinished {
		return nil
	}

	if !pos.IsValid() {
		return nil
	}

	if g.ForcedPiece != nil {
		if *g.ForcedPiece != pos {
			return nil
		}

		return findCapturesFrom(g, pos)
	}

	piece := g.Board.PieceAt(pos)
	if piece == nil {
		return nil
	}

	if piece.Color != g.Turn {
		return nil
	}

	allCaptures := findAllCaptures(g, g.Turn)
	if len(allCaptures) > 0 {
		return findCapturesFrom(g, pos)
	}

	return findSimpleMovesFrom(g, pos)
}

func findAllSimpleMoves(g *Game, color Color) []Move {
	var moves []Move

	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			pos := Position{Row: row, Col: col}
			piece := g.Board.PieceAt(pos)

			if piece == nil {
				continue
			}

			if piece.Color != color {
				continue
			}

			moves = append(moves, findSimpleMovesFrom(g, pos)...)
		}
	}

	return moves
}

func findSimpleMovesFrom(g *Game, pos Position) []Move {
	piece := g.Board.PieceAt(pos)
	if piece == nil {
		return nil
	}

	if piece.King {
		return findKingSimpleMoves(g, pos)
	}

	return findManSimpleMoves(g, pos)
}

func findManSimpleMoves(g *Game, pos Position) []Move {
	piece := g.Board.PieceAt(pos)
	if piece == nil {
		return nil
	}

	var rowDirection int

	switch piece.Color {
	case White:
		rowDirection = -1
	case Black:
		rowDirection = 1
	default:
		return nil
	}

	candidates := []Position{
		{Row: pos.Row + rowDirection, Col: pos.Col - 1},
		{Row: pos.Row + rowDirection, Col: pos.Col + 1},
	}

	moves := make([]Move, 0, len(candidates))

	for _, to := range candidates {
		if !to.IsValid() {
			continue
		}

		if !to.IsDarkSquare() {
			continue
		}

		if g.Board.PieceAt(to) != nil {
			continue
		}

		moves = append(moves, Move{
			From: pos,
			To:   to,
		})
	}

	return moves
}

func findKingSimpleMoves(g *Game, pos Position) []Move {
	piece := g.Board.PieceAt(pos)
	if piece == nil {
		return nil
	}

	directions := []struct {
		row int
		col int
	}{
		{row: -1, col: -1},
		{row: -1, col: 1},
		{row: 1, col: -1},
		{row: 1, col: 1},
	}

	var moves []Move

	for _, direction := range directions {
		current := Position{
			Row: pos.Row + direction.row,
			Col: pos.Col + direction.col,
		}

		for current.IsValid() {
			if !current.IsDarkSquare() {
				break
			}

			if g.Board.PieceAt(current) != nil {
				break
			}

			moves = append(moves, Move{
				From: pos,
				To:   current,
			})

			current = Position{
				Row: current.Row + direction.row,
				Col: current.Col + direction.col,
			}
		}
	}

	return moves
}

func findAllCaptures(g *Game, color Color) []Move {
	var moves []Move

	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			pos := Position{Row: row, Col: col}
			piece := g.Board.PieceAt(pos)

			if piece == nil {
				continue
			}

			if piece.Color != color {
				continue
			}

			moves = append(moves, findCapturesFrom(g, pos)...)
		}
	}

	return moves
}

func findCapturesFrom(g *Game, pos Position) []Move {
	piece := g.Board.PieceAt(pos)
	if piece == nil {
		return nil
	}

	if piece.King {
		return findKingCaptures(g, pos)
	}

	return findManCaptures(g, pos)
}

func findManCaptures(g *Game, pos Position) []Move {
	piece := g.Board.PieceAt(pos)
	if piece == nil {
		return nil
	}

	directions := []struct {
		row int
		col int
	}{
		{row: -1, col: -1},
		{row: -1, col: 1},
		{row: 1, col: -1},
		{row: 1, col: 1},
	}

	var moves []Move

	for _, direction := range directions {
		middle := Position{
			Row: pos.Row + direction.row,
			Col: pos.Col + direction.col,
		}

		to := Position{
			Row: pos.Row + 2*direction.row,
			Col: pos.Col + 2*direction.col,
		}

		if !to.IsDarkSquare() {
			continue
		}

		middlePiece := g.Board.PieceAt(middle)
		if middlePiece == nil {
			continue
		}

		if middlePiece.Color == piece.Color {
			continue
		}

		if g.Board.PieceAt(to) != nil {
			continue
		}

		moves = append(moves, Move{
			From: pos,
			To:   to,
		})
	}

	return moves
}

func findKingCaptures(g *Game, pos Position) []Move {
	piece := g.Board.PieceAt(pos)
	if piece == nil {
		return nil
	}

	directions := []struct {
		row int
		col int
	}{
		{row: -1, col: -1},
		{row: -1, col: 1},
		{row: 1, col: -1},
		{row: 1, col: 1},
	}

	var moves []Move

	for _, direction := range directions {
		current := Position{
			Row: pos.Row + direction.row,
			Col: pos.Col + direction.col,
		}

		var foundOpponent bool

		for current.IsValid() {
			currentPiece := g.Board.PieceAt(current)

			if !foundOpponent {
				if currentPiece == nil {
					current = Position{
						Row: current.Row + direction.row,
						Col: current.Col + direction.col,
					}
					continue
				}

				if currentPiece.Color == piece.Color {
					break
				}

				foundOpponent = true

				current = Position{
					Row: current.Row + direction.row,
					Col: current.Col + direction.col,
				}

				continue
			}

			if currentPiece != nil {
				break
			}

			moves = append(moves, Move{
				From: pos,
				To:   current,
			})

			current = Position{
				Row: current.Row + direction.row,
				Col: current.Col + direction.col,
			}
		}
	}

	return moves
}

func hasCaptureFrom(g *Game, pos Position) bool {
	return len(findCapturesFrom(g, pos)) > 0
}

func isCaptureMove(g *Game, move Move) bool {
	capturedPosition, err := capturedPositionForMove(g, move)
	return err == nil && capturedPosition != nil
}

func capturedPositionForMove(g *Game, move Move) (*Position, error) {
	if !move.isDiagonal() {
		return nil, ErrInvalidMove
	}

	rowStep := sign(move.rowDelta())
	colStep := sign(move.colDelta())

	current := Position{
		Row: move.From.Row + rowStep,
		Col: move.From.Col + colStep,
	}

	for current != move.To {
		piece := g.Board.PieceAt(current)
		if piece != nil {
			captured := current
			return &captured, nil
		}

		current = Position{
			Row: current.Row + rowStep,
			Col: current.Col + colStep,
		}
	}

	return nil, ErrInvalidMove
}

func sign(value int) int {
	switch {
	case value > 0:
		return 1
	case value < 0:
		return -1
	default:
		return 0
	}
}

func shouldPromote(piece Piece, to Position) bool {
	if piece.King {
		return false
	}

	switch piece.Color {
	case White:
		return to.Row == 0
	case Black:
		return to.Row == BoardSize-1
	default:
		return false
	}
}

func finishIfNeeded(g *Game) {
	if g == nil || g.Status == StatusFinished {
		return
	}

	if g.Board.CountPieces(g.Turn) == 0 {
		winner := g.Turn.Opponent()
		g.Status = StatusFinished
		g.Winner = &winner
		g.ForcedPiece = nil
		return
	}

	if len(g.LegalMoves()) == 0 {
		winner := g.Turn.Opponent()
		g.Status = StatusFinished
		g.Winner = &winner
		g.ForcedPiece = nil
		return
	}
}

func containsMove(moves []Move, target Move) bool {
	for _, move := range moves {
		if move == target {
			return true
		}
	}

	return false
}
