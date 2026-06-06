package checkers

type Game struct {
	Board Board `json:"board"`

	Turn Color `json:"turn"`

	Status Status `json:"status"`

	Winner *Color `json:"winner,omitempty"`

	ForcedPiece *Position `json:"forced_piece,omitempty"`
}

func NewGame() *Game {
	return &Game{
		Board:       NewInitialBoard(),
		Turn:        White,
		Status:      StatusActive,
		Winner:      nil,
		ForcedPiece: nil,
	}
}

func NewGameFromSnapshot(snapshot GameSnapshot) (*Game, error) {
	if !snapshot.Turn.IsValid() {
		return nil, ErrInvalidState
	}

	if !snapshot.Status.IsValid() {
		return nil, ErrInvalidState
	}

	switch snapshot.Status {
	case StatusActive:
		if snapshot.Winner != nil {
			return nil, ErrInvalidState
		}
	case StatusFinished:
		if snapshot.Winner != nil && !snapshot.Winner.IsValid() {
			return nil, ErrInvalidState
		}
	default:
		return nil, ErrInvalidState
	}

	if snapshot.ForcedPiece != nil {
		if !snapshot.ForcedPiece.IsValid() {
			return nil, ErrInvalidState
		}

		piece := snapshot.Board.PieceAt(*snapshot.ForcedPiece)
		if piece == nil {
			return nil, ErrInvalidState
		}

		if piece.Color != snapshot.Turn {
			return nil, ErrInvalidState
		}
	}

	board := snapshot.Board.Clone()

	var winner *Color
	if snapshot.Winner != nil {
		winnerValue := *snapshot.Winner
		winner = &winnerValue
	}

	var forcedPiece *Position
	if snapshot.ForcedPiece != nil {
		forcedPieceValue := *snapshot.ForcedPiece
		forcedPiece = &forcedPieceValue
	}

	return &Game{
		Board:       board,
		Turn:        snapshot.Turn,
		Status:      snapshot.Status,
		Winner:      winner,
		ForcedPiece: forcedPiece,
	}, nil
}

func (g *Game) Snapshot() GameSnapshot {
	if g == nil {
		return GameSnapshot{}
	}

	var winner *Color
	if g.Winner != nil {
		winnerValue := *g.Winner
		winner = &winnerValue
	}

	var forcedPiece *Position
	if g.ForcedPiece != nil {
		forcedPieceValue := *g.ForcedPiece
		forcedPiece = &forcedPieceValue
	}

	return GameSnapshot{
		Board:       g.Board.Clone(),
		Turn:        g.Turn,
		Status:      g.Status,
		Winner:      winner,
		ForcedPiece: forcedPiece,
	}
}

func (g *Game) Clone() *Game {
	if g == nil {
		return nil
	}

	snapshot := g.Snapshot()

	clone, err := NewGameFromSnapshot(snapshot)
	if err != nil {
		return nil
	}

	return clone
}

func (g *Game) IsFinished() bool {
	if g == nil {
		return false
	}

	return g.Status == StatusFinished
}

func (g *Game) WinnerColor() *Color {
	if g == nil || g.Winner == nil {
		return nil
	}

	winner := *g.Winner
	return &winner
}

func (g *Game) ApplyMove(move Move) (MoveResult, error) {
	result := MoveResult{
		Move: move,
	}

	if g == nil {
		return result, ErrInvalidState
	}

	if g.Status == StatusFinished {
		return result, ErrGameFinished
	}

	legalMoves := g.LegalMoves()
	if !containsMove(legalMoves, move) {
		return result, ErrInvalidMove
	}

	piece := g.Board.PieceAt(move.From)
	if piece == nil {
		return result, ErrPieceNotFound
	}

	captured := isCaptureMove(g, move)

	var capturedPosition *Position
	if captured {
		var err error
		capturedPosition, err = capturedPositionForMove(g, move)
		if err != nil {
			return result, err
		}
	}

	movedPiece := *piece

	if shouldPromote(movedPiece, move.To) {
		movedPiece.King = true
		result.Promoted = true
	}

	g.Board.RemovePiece(move.From)
	g.Board.SetPiece(move.To, &movedPiece)

	if captured {
		g.Board.RemovePiece(*capturedPosition)

		result.Captured = true
		result.CapturedPosition = capturedPosition
	}

	if captured && hasCaptureFrom(g, move.To) {
		forcedPiece := move.To
		g.ForcedPiece = &forcedPiece

		result.MustContinueCapture = true

		return result, nil
	}

	g.ForcedPiece = nil
	g.Turn = g.Turn.Opponent()

	finishIfNeeded(g)

	if g.Status == StatusFinished {
		result.GameFinished = true
		result.Winner = g.Winner
	}

	return result, nil
}
