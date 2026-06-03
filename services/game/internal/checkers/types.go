package checkers

const BoardSize = 8

type Color string

const (
	White Color = "white"
	Black Color = "black"
)

func (c Color) Opponent() Color {
	switch c {
	case White:
		return Black
	case Black:
		return White
	default:
		return ""
	}
}

func (c Color) IsValid() bool {
	return c == White || c == Black
}

type Status string

const (
	StatusActive   Status = "active"
	StatusFinished Status = "finished"
)

func (s Status) IsValid() bool {
	return s == StatusActive || s == StatusFinished
}

type Position struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

func (p Position) IsValid() bool {
	return p.Row >= 0 && p.Row < BoardSize && p.Col >= 0 && p.Col < BoardSize
}

func (p Position) IsDarkSquare() bool {
	return p.IsValid() && (p.Row+p.Col)%2 == 1
}

type Piece struct {
	Color Color `json:"color"`
	King  bool  `json:"king"`
}

type Board struct {
	Cells [BoardSize][BoardSize]*Piece `json:"cells"`
}

func NewInitialBoard() Board {
	var board Board

	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			pos := Position{Row: row, Col: col}

			if !pos.IsDarkSquare() {
				continue
			}

			switch {
			case row <= 2:
				board.Cells[row][col] = &Piece{
					Color: Black,
					King:  false,
				}
			case row >= 5:
				board.Cells[row][col] = &Piece{
					Color: White,
					King:  false,
				}
			}
		}
	}
	return board
}

func (b Board) PieceAt(pos Position) *Piece {
	if !pos.IsValid() {
		return nil
	}

	return b.Cells[pos.Row][pos.Col]
}

func (b *Board) SetPiece(pos Position, piece *Piece) {
	if !pos.IsValid() {
		return
	}

	b.Cells[pos.Row][pos.Col] = piece
}

func (b *Board) RemovePiece(pos Position) {
	if !pos.IsValid() {
		return
	}

	b.Cells[pos.Row][pos.Col] = nil
}

func (b Board) CountPieces(color Color) int {
	count := 0

	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			piece := b.Cells[row][col]
			if piece != nil && piece.Color == color {
				count++
			}
		}
	}

	return count
}

func (b Board) Clone() Board {
	var clone Board

	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			piece := b.Cells[row][col]
			if piece == nil {
				continue
			}

			pieceCopy := *piece
			clone.Cells[row][col] = &pieceCopy
		}
	}

	return clone
}

type Move struct {
	From Position `json:"from"`
	To   Position `json:"to"`
}

func (m Move) rowDelta() int {
	return m.To.Row - m.From.Row
}

func (m Move) colDelta() int {
	return m.To.Col - m.From.Col
}

func (m Move) absRowDelta() int {
	delta := m.rowDelta()
	if delta < 0 {
		return -delta
	}
	return delta
}

func (m Move) absColDelta() int {
	delta := m.colDelta()
	if delta < 0 {
		return -delta
	}

	return delta
}

func (m Move) isDiagonal() bool {
	return m.absRowDelta() == m.absColDelta()
}

type GameSnapshot struct {
	Board       Board     `json:"board"`
	Turn        Color     `json:"turn"`
	Status      Status    `json:"status"`
	Winner      *Color    `json:"winner,omitempty"`
	ForcedPiece *Position `json:"forced_piece,omitempty"`
}

type MoveResult struct {
	Move Move `json:"move"`

	Captured         bool      `json:"captured"`
	CapturedPosition *Position `json:"captured_position,omitempty"`

	Promoted bool `json:"promoted"`

	MustContinueCapture bool `json:"must_continue_capture"`

	GameFinished bool   `json:"game_finished"`
	Winner       *Color `json:"winner,omitempty"`
}
