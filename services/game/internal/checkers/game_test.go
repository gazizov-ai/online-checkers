package checkers

import "testing"

func TestNewGameInitialState(t *testing.T) {
	game := NewGame()

	if game == nil {
		t.Fatal("expected game, got nil")
	}

	if game.Status != StatusActive {
		t.Fatalf("expected status %q, got %q", StatusActive, game.Status)
	}

	if game.Turn != White {
		t.Fatalf("expected first turn %q, got %q", White, game.Turn)
	}

	if game.Winner != nil {
		t.Fatalf("expected no winner in the beginning, got %v", *game.Winner)
	}

	if game.ForcedPiece != nil {
		t.Fatalf("expected no forced piece, got %+v", *game.ForcedPiece)
	}
}

func TestInitialBoard(t *testing.T) {
	t.Run("has correct piece count", func(t *testing.T) {
		board := NewInitialBoard()

		if got := board.CountPieces(White); got != 12 {
			t.Fatalf("expected 12 white pieces, got %d", got)
		}

		if got := board.CountPieces(Black); got != 12 {
			t.Fatalf("expected 12 black pieces, got %d", got)
		}
	})

	t.Run("places pieces on dark squares in starting rows", func(t *testing.T) {
		board := NewInitialBoard()

		for row := 0; row < BoardSize; row++ {
			for col := 0; col < BoardSize; col++ {
				pos := Position{Row: row, Col: col}
				piece := board.PieceAt(pos)
				if piece == nil {
					continue
				}

				if !pos.IsDarkSquare() {
					t.Fatalf("expected piece at %+v to be on dark square", pos)
				}

				switch piece.Color {
				case White:
					if row < 5 || row > 7 {
						t.Fatalf("unexpected white piece at %+v", pos)
					}
				case Black:
					if row < 0 || row > 2 {
						t.Fatalf("unexpected black piece at %+v", pos)
					}
				default:
					t.Fatalf("unexpected piece color %q at %+v", piece.Color, pos)
				}
			}
		}
	})
}

func TestBoardAndPositionOperations(t *testing.T) {
	t.Run("validates positions", func(t *testing.T) {
		valid := Position{Row: 0, Col: 0}
		if !valid.IsValid() {
			t.Fatalf("expected %+v to be valid", valid)
		}

		invalidPositions := []Position{
			{Row: -1, Col: 0},
			{Row: 0, Col: -1},
			{Row: BoardSize, Col: 0},
			{Row: 0, Col: BoardSize},
		}

		for _, pos := range invalidPositions {
			if pos.IsValid() {
				t.Fatalf("expected %+v to be invalid", pos)
			}
		}
	})

	t.Run("piece at invalid position returns nil", func(t *testing.T) {
		board := NewInitialBoard()

		piece := board.PieceAt(Position{Row: -1, Col: 0})
		if piece != nil {
			t.Fatalf("expected nil for invalid position, got %+v", *piece)
		}
	})

	t.Run("set piece ignores invalid position", func(t *testing.T) {
		board := Board{}

		board.SetPiece(Position{Row: -1, Col: 0}, &Piece{Color: White})

		if got := board.CountPieces(White); got != 0 {
			t.Fatalf("expected no pieces to be set, got %d", got)
		}
	})

	t.Run("remove piece ignores invalid position", func(t *testing.T) {
		board := NewInitialBoard()

		beforeWhite := board.CountPieces(White)
		beforeBlack := board.CountPieces(Black)

		board.RemovePiece(Position{Row: -1, Col: 0})

		if got := board.CountPieces(White); got != beforeWhite {
			t.Fatalf("expected white count %d, got %d", beforeWhite, got)
		}

		if got := board.CountPieces(Black); got != beforeBlack {
			t.Fatalf("expected black count %d, got %d", beforeBlack, got)
		}
	})

	t.Run("clone is deep copy", func(t *testing.T) {
		board := NewInitialBoard()
		pos := Position{Row: 5, Col: 0}

		originalPiece := board.PieceAt(pos)
		if originalPiece == nil {
			t.Fatalf("expected piece at %+v", pos)
		}

		clone := board.Clone()

		clonePiece := clone.PieceAt(pos)
		if clonePiece == nil {
			t.Fatalf("expected cloned piece at %+v", pos)
		}

		if clonePiece == originalPiece {
			t.Fatal("expected cloned piece to have different pointer")
		}

		clonePiece.King = true

		if originalPiece.King {
			t.Fatal("expected original piece not to be changed after modifying clone")
		}
	})

	t.Run("invalid color has no opponent", func(t *testing.T) {
		if got := Color("red").Opponent(); got != "" {
			t.Fatalf("expected empty opponent for invalid color, got %q", got)
		}
	})
}

func TestSnapshotsAndRestoration(t *testing.T) {
	t.Run("snapshot is deep copy", func(t *testing.T) {
		game := NewGame()
		pos := Position{Row: 5, Col: 0}

		snapshot := game.Snapshot()

		snapshotPiece := snapshot.Board.PieceAt(pos)
		if snapshotPiece == nil {
			t.Fatalf("expected piece at %+v in snapshot", pos)
		}

		snapshotPiece.King = true

		originalPiece := game.Board.PieceAt(pos)
		if originalPiece == nil {
			t.Fatalf("expected piece at %+v in original game", pos)
		}

		if originalPiece.King {
			t.Fatal("expected original game not to be changed after modifying snapshot")
		}
	})

	t.Run("snapshot copies winner and forced piece", func(t *testing.T) {
		game := NewGame()

		winner := White
		forcedPiece := Position{Row: 5, Col: 0}

		game.Status = StatusFinished
		game.Winner = &winner
		game.ForcedPiece = &forcedPiece

		snapshot := game.Snapshot()

		if snapshot.Winner == nil {
			t.Fatal("expected winner in snapshot")
		}

		if *snapshot.Winner != winner {
			t.Fatalf("expected winner %q, got %q", winner, *snapshot.Winner)
		}

		if snapshot.ForcedPiece == nil {
			t.Fatal("expected forced piece in snapshot")
		}

		if *snapshot.ForcedPiece != forcedPiece {
			t.Fatalf("expected forced piece %+v, got %+v", forcedPiece, *snapshot.ForcedPiece)
		}

		*snapshot.Winner = Black
		snapshot.ForcedPiece.Row = 0

		if *game.Winner != White {
			t.Fatal("expected snapshot winner to be copied")
		}

		if game.ForcedPiece.Row != forcedPiece.Row {
			t.Fatal("expected snapshot forced piece to be copied")
		}
	})

	t.Run("nil game snapshot is empty", func(t *testing.T) {
		var game *Game

		snapshot := game.Snapshot()

		if snapshot.Board.CountPieces(White) != 0 {
			t.Fatal("expected empty snapshot board")
		}

		if snapshot.Turn != "" {
			t.Fatalf("expected empty turn, got %q", snapshot.Turn)
		}

		if snapshot.Status != "" {
			t.Fatalf("expected empty status, got %q", snapshot.Status)
		}
	})

	t.Run("restores valid active snapshot", func(t *testing.T) {
		game := NewGame()

		restored, err := NewGameFromSnapshot(game.Snapshot())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if restored == nil {
			t.Fatal("expected restored game, got nil")
		}

		if restored.Turn != game.Turn {
			t.Fatalf("expected turn %q, got %q", game.Turn, restored.Turn)
		}

		if restored.Status != game.Status {
			t.Fatalf("expected status %q, got %q", game.Status, restored.Status)
		}
	})

	t.Run("restores valid finished snapshot", func(t *testing.T) {
		snapshot := NewGame().Snapshot()

		winner := White
		snapshot.Status = StatusFinished
		snapshot.Winner = &winner

		game, err := NewGameFromSnapshot(snapshot)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if game.Status != StatusFinished {
			t.Fatalf("expected status %q, got %q", StatusFinished, game.Status)
		}

		if game.Winner == nil {
			t.Fatal("expected winner")
		}

		if *game.Winner != White {
			t.Fatalf("expected winner %q, got %q", White, *game.Winner)
		}

		*snapshot.Winner = Black

		if *game.Winner != White {
			t.Fatal("expected restored winner to be copied")
		}
	})

	t.Run("restores valid forced piece", func(t *testing.T) {
		snapshot := NewGame().Snapshot()

		forcedPiece := Position{Row: 5, Col: 0}
		expectedForcedPiece := forcedPiece
		snapshot.ForcedPiece = &forcedPiece

		game, err := NewGameFromSnapshot(snapshot)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if game.ForcedPiece == nil {
			t.Fatal("expected forced piece")
		}

		if *game.ForcedPiece != expectedForcedPiece {
			t.Fatalf("expected forced piece %+v, got %+v", expectedForcedPiece, *game.ForcedPiece)
		}

		snapshot.ForcedPiece.Row = 0

		if *game.ForcedPiece != expectedForcedPiece {
			t.Fatal("expected restored forced piece to be copied")
		}
	})

	t.Run("rejects invalid snapshots", func(t *testing.T) {
		tests := []struct {
			name   string
			mutate func(*GameSnapshot)
		}{
			{
				name: "invalid turn",
				mutate: func(snapshot *GameSnapshot) {
					snapshot.Turn = Color("red")
				},
			},
			{
				name: "invalid status",
				mutate: func(snapshot *GameSnapshot) {
					snapshot.Status = Status("paused")
				},
			},
			{
				name: "active game with winner",
				mutate: func(snapshot *GameSnapshot) {
					winner := White
					snapshot.Winner = &winner
				},
			},
			{
				name: "finished game without winner",
				mutate: func(snapshot *GameSnapshot) {
					snapshot.Status = StatusFinished
					snapshot.Winner = nil
				},
			},
			{
				name: "invalid forced piece",
				mutate: func(snapshot *GameSnapshot) {
					forcedPiece := Position{Row: -1, Col: 0}
					snapshot.ForcedPiece = &forcedPiece
				},
			},
			{
				name: "forced piece without piece",
				mutate: func(snapshot *GameSnapshot) {
					forcedPiece := Position{Row: 4, Col: 1}
					snapshot.ForcedPiece = &forcedPiece
				},
			},
			{
				name: "forced piece of wrong color",
				mutate: func(snapshot *GameSnapshot) {
					forcedPiece := Position{Row: 2, Col: 1}
					snapshot.ForcedPiece = &forcedPiece
					snapshot.Turn = White
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				snapshot := NewGame().Snapshot()
				tt.mutate(&snapshot)

				_, err := NewGameFromSnapshot(snapshot)
				if err != ErrInvalidState {
					t.Fatalf("expected ErrInvalidState, got %v", err)
				}
			})
		}
	})
}

func TestGameCopiesAndAccessors(t *testing.T) {
	t.Run("clone returns deep copy", func(t *testing.T) {
		game := NewGame()

		clone := game.Clone()
		if clone == nil {
			t.Fatal("expected clone, got nil")
		}

		pos := Position{Row: 5, Col: 0}

		clonePiece := clone.Board.PieceAt(pos)
		if clonePiece == nil {
			t.Fatalf("expected piece at %+v in clone", pos)
		}

		clonePiece.King = true

		originalPiece := game.Board.PieceAt(pos)
		if originalPiece == nil {
			t.Fatalf("expected piece at %+v in original", pos)
		}

		if originalPiece.King {
			t.Fatal("expected original not to change after modifying clone")
		}
	})

	t.Run("nil clone returns nil", func(t *testing.T) {
		var game *Game

		if clone := game.Clone(); clone != nil {
			t.Fatalf("expected nil clone, got %+v", clone)
		}
	})

	t.Run("is finished reports status", func(t *testing.T) {
		game := NewGame()

		if game.IsFinished() {
			t.Fatal("expected new game not to be finished")
		}

		winner := White
		game.Status = StatusFinished
		game.Winner = &winner

		if !game.IsFinished() {
			t.Fatal("expected finished game to be finished")
		}
	})

	t.Run("nil game is not finished", func(t *testing.T) {
		var game *Game

		if game.IsFinished() {
			t.Fatal("expected nil game not to be finished")
		}
	})

	t.Run("winner color returns copy", func(t *testing.T) {
		game := NewGame()

		if winner := game.WinnerColor(); winner != nil {
			t.Fatalf("expected no winner, got %q", *winner)
		}

		winner := White
		game.Status = StatusFinished
		game.Winner = &winner

		got := game.WinnerColor()
		if got == nil {
			t.Fatal("expected winner")
		}

		if *got != White {
			t.Fatalf("expected winner %q, got %q", White, *got)
		}

		*got = Black

		if *game.Winner != White {
			t.Fatal("expected WinnerColor to return copy, not original pointer")
		}
	})

	t.Run("nil game winner is nil", func(t *testing.T) {
		var game *Game

		if winner := game.WinnerColor(); winner != nil {
			t.Fatalf("expected nil winner, got %q", *winner)
		}
	})
}

func TestLegalMoves(t *testing.T) {
	t.Run("nil game returns nil", func(t *testing.T) {
		var game *Game

		if moves := game.LegalMoves(); moves != nil {
			t.Fatalf("expected nil moves, got %+v", moves)
		}
	})

	t.Run("initial position has seven white moves", func(t *testing.T) {
		game := NewGame()

		moves := game.LegalMoves()

		if len(moves) != 7 {
			t.Fatalf("expected 7 legal moves for white at initial position, got %d: %+v", len(moves), moves)
		}
	})

	t.Run("from initial white piece", func(t *testing.T) {
		game := NewGame()
		from := Position{Row: 5, Col: 0}

		assertMoves(t, game.LegalMovesFrom(from), Move{
			From: from,
			To:   Position{Row: 4, Col: 1},
		})
	})

	t.Run("from unavailable positions", func(t *testing.T) {
		tests := []struct {
			name string
			pos  Position
		}{
			{name: "blocked white piece", pos: Position{Row: 6, Col: 1}},
			{name: "opponent piece", pos: Position{Row: 2, Col: 1}},
			{name: "empty square", pos: Position{Row: 4, Col: 1}},
			{name: "invalid position", pos: Position{Row: -1, Col: 0}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				game := NewGame()

				assertNoMoves(t, game.LegalMovesFrom(tt.pos))
			})
		}
	})

	t.Run("finished game returns no moves", func(t *testing.T) {
		game := NewGame()
		winner := White

		game.Status = StatusFinished
		game.Winner = &winner

		assertNoMoves(t, game.LegalMoves())
		assertNoMoves(t, game.LegalMovesFrom(Position{Row: 5, Col: 0}))
	})

	t.Run("nil game returns no moves from position", func(t *testing.T) {
		var game *Game

		if moves := game.LegalMovesFrom(Position{Row: 5, Col: 0}); moves != nil {
			t.Fatalf("expected nil moves, got %+v", moves)
		}
	})

	t.Run("capture is mandatory", func(t *testing.T) {
		game := emptyGame(White)

		capturingWhitePos := Position{Row: 4, Col: 1}
		blackPos := Position{Row: 3, Col: 2}
		captureLanding := Position{Row: 2, Col: 3}
		nonCapturingWhitePos := Position{Row: 5, Col: 4}

		setPiece(&game.Board, capturingWhitePos, White, false)
		setPiece(&game.Board, blackPos, Black, false)
		setPiece(&game.Board, nonCapturingWhitePos, White, false)

		expectedCapture := Move{
			From: capturingWhitePos,
			To:   captureLanding,
		}

		assertMoves(t, game.LegalMoves(), expectedCapture)
		assertMoves(t, game.LegalMovesFrom(capturingWhitePos), expectedCapture)
		assertNoMoves(t, game.LegalMovesFrom(nonCapturingWhitePos))
	})

	t.Run("forced piece restricts moves", func(t *testing.T) {
		game := emptyGame(White)

		forced := Position{Row: 3, Col: 2}
		other := Position{Row: 5, Col: 0}
		black := Position{Row: 2, Col: 3}
		landing := Position{Row: 1, Col: 4}

		setPiece(&game.Board, forced, White, false)
		setPiece(&game.Board, other, White, false)
		setPiece(&game.Board, black, Black, false)
		game.ForcedPiece = &forced

		assertNoMoves(t, game.LegalMovesFrom(other))
		assertMoves(t, game.LegalMovesFrom(forced), Move{
			From: forced,
			To:   landing,
		})
	})
}

func TestApplyMoveSimpleAndInvalidMoves(t *testing.T) {
	t.Run("simple move updates board", func(t *testing.T) {
		game := NewGame()

		move := Move{
			From: Position{Row: 5, Col: 0},
			To:   Position{Row: 4, Col: 1},
		}

		result, err := game.ApplyMove(move)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.Move != move {
			t.Fatalf("expected result move %+v, got %+v", move, result.Move)
		}

		if result.Captured {
			t.Fatal("expected simple move not to capture")
		}

		if result.Promoted {
			t.Fatal("expected simple move not to promote")
		}

		if result.MustContinueCapture {
			t.Fatal("expected simple move not to require continuing capture")
		}

		if result.GameFinished {
			t.Fatal("expected game not to finish")
		}

		if result.Winner != nil {
			t.Fatalf("expected no winner, got %v", *result.Winner)
		}

		assertEmptySquare(t, game.Board, move.From)
		assertPiece(t, game.Board, move.To, White, false)

		if game.Turn != Black {
			t.Fatalf("expected turn to switch to %q, got %q", Black, game.Turn)
		}
	})

	t.Run("black moves after white move", func(t *testing.T) {
		game := NewGame()

		whiteMove := Move{
			From: Position{Row: 5, Col: 0},
			To:   Position{Row: 4, Col: 1},
		}

		_, err := game.ApplyMove(whiteMove)
		if err != nil {
			t.Fatalf("expected white move to be valid, got %v", err)
		}

		if game.Turn != Black {
			t.Fatalf("expected turn to be %q, got %q", Black, game.Turn)
		}

		blackMove := Move{
			From: Position{Row: 2, Col: 1},
			To:   Position{Row: 3, Col: 0},
		}

		result, err := game.ApplyMove(blackMove)
		if err != nil {
			t.Fatalf("expected black move to be valid, got %v", err)
		}

		if result.Move != blackMove {
			t.Fatalf("expected result move %+v, got %+v", blackMove, result.Move)
		}

		assertEmptySquare(t, game.Board, blackMove.From)
		assertPiece(t, game.Board, blackMove.To, Black, false)

		if game.Turn != White {
			t.Fatalf("expected turn to switch back to %q, got %q", White, game.Turn)
		}
	})

	t.Run("rejects invalid move", func(t *testing.T) {
		game := NewGame()

		move := Move{
			From: Position{Row: 5, Col: 0},
			To:   Position{Row: 3, Col: 2},
		}

		_, err := game.ApplyMove(move)
		if err != ErrInvalidMove {
			t.Fatalf("expected ErrInvalidMove, got %v", err)
		}
	})

	t.Run("invalid move does not change board", func(t *testing.T) {
		game := NewGame()
		before := game.Snapshot()

		move := Move{
			From: Position{Row: 5, Col: 0},
			To:   Position{Row: 3, Col: 2},
		}

		_, err := game.ApplyMove(move)
		if err != ErrInvalidMove {
			t.Fatalf("expected ErrInvalidMove, got %v", err)
		}

		after := game.Snapshot()

		if before.Turn != after.Turn {
			t.Fatalf("expected turn not to change, before %q after %q", before.Turn, after.Turn)
		}

		if before.Status != after.Status {
			t.Fatalf("expected status not to change, before %q after %q", before.Status, after.Status)
		}

		assertBoardsEqual(t, before.Board, after.Board)
	})

	t.Run("rejects move after game finished", func(t *testing.T) {
		game := NewGame()
		winner := White

		game.Status = StatusFinished
		game.Winner = &winner

		move := Move{
			From: Position{Row: 5, Col: 0},
			To:   Position{Row: 4, Col: 1},
		}

		_, err := game.ApplyMove(move)
		if err != ErrGameFinished {
			t.Fatalf("expected ErrGameFinished, got %v", err)
		}
	})
}

func TestManCapturesAndPromotion(t *testing.T) {
	t.Run("capture removes opponent piece", func(t *testing.T) {
		game := emptyGame(White)

		whitePos := Position{Row: 4, Col: 1}
		blackPos := Position{Row: 3, Col: 2}
		to := Position{Row: 2, Col: 3}

		setPiece(&game.Board, whitePos, White, false)
		setPiece(&game.Board, blackPos, Black, false)

		move := Move{From: whitePos, To: to}

		result, err := game.ApplyMove(move)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assertCaptured(t, result, blackPos)
		assertEmptySquare(t, game.Board, whitePos)
		assertEmptySquare(t, game.Board, blackPos)
		assertPiece(t, game.Board, to, White, false)

		if game.Turn != Black {
			t.Fatalf("expected turn to switch to black, got %q", game.Turn)
		}
	})

	t.Run("simple move is rejected when capture is available", func(t *testing.T) {
		game := emptyGame(White)

		capturingWhitePos := Position{Row: 4, Col: 1}
		blackPos := Position{Row: 3, Col: 2}
		captureLanding := Position{Row: 2, Col: 3}
		simpleWhitePos := Position{Row: 5, Col: 4}
		simpleMoveTo := Position{Row: 4, Col: 5}

		setPiece(&game.Board, capturingWhitePos, White, false)
		setPiece(&game.Board, blackPos, Black, false)
		setPiece(&game.Board, simpleWhitePos, White, false)

		assertMoves(t, game.LegalMoves(), Move{
			From: capturingWhitePos,
			To:   captureLanding,
		})

		_, err := game.ApplyMove(Move{From: simpleWhitePos, To: simpleMoveTo})
		if err != ErrInvalidMove {
			t.Fatalf("expected ErrInvalidMove for simple move when capture is available, got %v", err)
		}
	})

	t.Run("man can capture backward", func(t *testing.T) {
		game := emptyGame(White)

		whitePos := Position{Row: 3, Col: 2}
		blackPos := Position{Row: 4, Col: 3}
		landing := Position{Row: 5, Col: 4}

		setPiece(&game.Board, whitePos, White, false)
		setPiece(&game.Board, blackPos, Black, false)

		assertMoves(t, game.LegalMovesFrom(whitePos), Move{
			From: whitePos,
			To:   landing,
		})
	})

	t.Run("man cannot capture if landing square is occupied", func(t *testing.T) {
		game := emptyGame(White)

		whitePos := Position{Row: 4, Col: 1}
		blackPos := Position{Row: 3, Col: 2}
		occupiedLanding := Position{Row: 2, Col: 3}

		setPiece(&game.Board, whitePos, White, false)
		setPiece(&game.Board, blackPos, Black, false)
		setPiece(&game.Board, occupiedLanding, White, false)

		moves := game.LegalMovesFrom(whitePos)

		invalidCapture := Move{
			From: whitePos,
			To:   occupiedLanding,
		}

		if containsMove(moves, invalidCapture) {
			t.Fatalf("expected capture %+v not to be legal when landing square is occupied, got %+v", invalidCapture, moves)
		}

		assertMoves(t, moves, Move{
			From: whitePos,
			To:   Position{Row: 3, Col: 0},
		})
	})

	t.Run("multiple capture keeps turn and sets forced piece", func(t *testing.T) {
		game, start, afterFirstCapture, afterSecondCapture := newManCaptureChain()

		result, err := game.ApplyMove(Move{From: start, To: afterFirstCapture})
		if err != nil {
			t.Fatalf("expected first capture to be valid, got %v", err)
		}

		if !result.Captured {
			t.Fatal("expected first move to be capture")
		}

		if !result.MustContinueCapture {
			t.Fatal("expected must continue capture")
		}

		if game.Turn != White {
			t.Fatalf("expected turn to remain %q, got %q", White, game.Turn)
		}

		assertForcedPiece(t, game, afterFirstCapture)
		assertMoves(t, game.LegalMoves(), Move{
			From: afterFirstCapture,
			To:   afterSecondCapture,
		})
	})

	t.Run("multiple capture rejects other piece move", func(t *testing.T) {
		game, start, afterFirstCapture, _ := newManCaptureChain()
		otherWhite := Position{Row: 5, Col: 6}
		setPiece(&game.Board, otherWhite, White, false)

		_, err := game.ApplyMove(Move{From: start, To: afterFirstCapture})
		if err != nil {
			t.Fatalf("expected first capture to be valid, got %v", err)
		}

		_, err = game.ApplyMove(Move{
			From: otherWhite,
			To:   Position{Row: 4, Col: 7},
		})
		if err != ErrInvalidMove {
			t.Fatalf("expected ErrInvalidMove when moving other piece during capture chain, got %v", err)
		}
	})

	t.Run("multiple capture clears forced piece after final capture", func(t *testing.T) {
		game, start, afterFirstCapture, afterSecondCapture := newManCaptureChain()
		firstBlack := Position{Row: 4, Col: 1}
		secondBlack := Position{Row: 2, Col: 3}

		_, err := game.ApplyMove(Move{From: start, To: afterFirstCapture})
		if err != nil {
			t.Fatalf("expected first capture to be valid, got %v", err)
		}

		result, err := game.ApplyMove(Move{From: afterFirstCapture, To: afterSecondCapture})
		if err != nil {
			t.Fatalf("expected second capture to be valid, got %v", err)
		}

		if !result.Captured {
			t.Fatal("expected second move to be capture")
		}

		if result.MustContinueCapture {
			t.Fatal("expected capture chain to be finished")
		}

		if game.ForcedPiece != nil {
			t.Fatalf("expected forced piece to be cleared, got %+v", *game.ForcedPiece)
		}

		if game.Turn != Black {
			t.Fatalf("expected turn to switch to %q, got %q", Black, game.Turn)
		}

		assertEmptySquare(t, game.Board, firstBlack)
		assertEmptySquare(t, game.Board, secondBlack)
		assertPiece(t, game.Board, afterSecondCapture, White, false)
	})

	t.Run("promotes white man to king", func(t *testing.T) {
		game := emptyGame(White)

		from := Position{Row: 1, Col: 2}
		to := Position{Row: 0, Col: 1}

		setPiece(&game.Board, from, White, false)

		result, err := game.ApplyMove(Move{From: from, To: to})
		if err != nil {
			t.Fatalf("expected promotion move to be valid, got %v", err)
		}

		if !result.Promoted {
			t.Fatal("expected result.Promoted to be true")
		}

		assertPiece(t, game.Board, to, White, true)

		if game.Turn != Black {
			t.Fatalf("expected turn to switch to %q, got %q", Black, game.Turn)
		}
	})

	t.Run("promotes black man to king", func(t *testing.T) {
		game := emptyGame(Black)

		from := Position{Row: 6, Col: 1}
		to := Position{Row: 7, Col: 0}

		setPiece(&game.Board, from, Black, false)

		result, err := game.ApplyMove(Move{From: from, To: to})
		if err != nil {
			t.Fatalf("expected promotion move to be valid, got %v", err)
		}

		if !result.Promoted {
			t.Fatal("expected result.Promoted to be true")
		}

		assertPiece(t, game.Board, to, Black, true)

		if game.Turn != White {
			t.Fatalf("expected turn to switch to %q, got %q", White, game.Turn)
		}
	})

	t.Run("promotes white man after capture", func(t *testing.T) {
		game := emptyGame(White)

		from := Position{Row: 2, Col: 1}
		blackPos := Position{Row: 1, Col: 2}
		to := Position{Row: 0, Col: 3}

		setPiece(&game.Board, from, White, false)
		setPiece(&game.Board, blackPos, Black, false)

		result, err := game.ApplyMove(Move{From: from, To: to})
		if err != nil {
			t.Fatalf("expected capture promotion to be valid, got %v", err)
		}

		assertCaptured(t, result, blackPos)

		if !result.Promoted {
			t.Fatal("expected move to promote")
		}

		assertEmptySquare(t, game.Board, blackPos)
		assertPiece(t, game.Board, to, White, true)
	})

	t.Run("man promotes during capture and must continue as king", func(t *testing.T) {
		game, start, promotionSquare, _ := newPromotionCaptureChain()

		result, err := game.ApplyMove(Move{From: start, To: promotionSquare})
		if err != nil {
			t.Fatalf("expected promotion capture to be valid, got %v", err)
		}

		if !result.Captured {
			t.Fatal("expected first move to capture")
		}

		if !result.Promoted {
			t.Fatal("expected man to promote during capture")
		}

		if !result.MustContinueCapture {
			t.Fatal("expected promoted king to continue capture")
		}

		if game.Turn != White {
			t.Fatalf("expected turn to remain %q, got %q", White, game.Turn)
		}

		assertForcedPiece(t, game, promotionSquare)
		assertPiece(t, game.Board, promotionSquare, White, true)
		assertMoves(t, game.LegalMoves(),
			Move{From: promotionSquare, To: Position{Row: 3, Col: 6}},
			Move{From: promotionSquare, To: Position{Row: 4, Col: 7}},
		)
	})

	t.Run("promoted king capture chain finishes after final capture", func(t *testing.T) {
		game, start, promotionSquare, finalLanding := newPromotionCaptureChain()
		firstBlack := Position{Row: 1, Col: 2}
		secondBlack := Position{Row: 2, Col: 5}

		_, err := game.ApplyMove(Move{From: start, To: promotionSquare})
		if err != nil {
			t.Fatalf("expected first capture to be valid, got %v", err)
		}

		result, err := game.ApplyMove(Move{From: promotionSquare, To: finalLanding})
		if err != nil {
			t.Fatalf("expected promoted king continuation to be valid, got %v", err)
		}

		if !result.Captured {
			t.Fatal("expected second move to capture")
		}

		if result.Promoted {
			t.Fatal("expected second move not to promote again")
		}

		if result.MustContinueCapture {
			t.Fatal("expected capture chain to be finished")
		}

		if game.ForcedPiece != nil {
			t.Fatalf("expected forced piece to be cleared, got %+v", *game.ForcedPiece)
		}

		if game.Turn != Black {
			t.Fatalf("expected turn to switch to %q, got %q", Black, game.Turn)
		}

		assertEmptySquare(t, game.Board, firstBlack)
		assertEmptySquare(t, game.Board, secondBlack)
		assertPiece(t, game.Board, finalLanding, White, true)
	})
}

func TestKingMovementAndCaptures(t *testing.T) {
	t.Run("can move any distance diagonally", func(t *testing.T) {
		game := emptyGame(White)

		kingPos := Position{Row: 5, Col: 0}
		setPiece(&game.Board, kingPos, White, true)

		assertMoves(t, game.LegalMovesFrom(kingPos),
			Move{From: kingPos, To: Position{Row: 4, Col: 1}},
			Move{From: kingPos, To: Position{Row: 3, Col: 2}},
			Move{From: kingPos, To: Position{Row: 2, Col: 3}},
			Move{From: kingPos, To: Position{Row: 1, Col: 4}},
			Move{From: kingPos, To: Position{Row: 0, Col: 5}},
			Move{From: kingPos, To: Position{Row: 6, Col: 1}},
			Move{From: kingPos, To: Position{Row: 7, Col: 2}},
		)
	})

	t.Run("cannot move non diagonally", func(t *testing.T) {
		game := emptyGame(White)

		from := Position{Row: 5, Col: 0}
		setPiece(&game.Board, from, White, true)

		_, err := game.ApplyMove(Move{
			From: from,
			To:   Position{Row: 5, Col: 4},
		})
		if err != ErrInvalidMove {
			t.Fatalf("expected ErrInvalidMove for non-diagonal king move, got %v", err)
		}
	})

	t.Run("simple move cannot jump over piece", func(t *testing.T) {
		game := emptyGame(White)

		kingPos := Position{Row: 5, Col: 0}
		blockingPos := Position{Row: 3, Col: 2}

		setPiece(&game.Board, kingPos, White, true)
		setPiece(&game.Board, blockingPos, White, false)

		moves := game.LegalMovesFrom(kingPos)

		allowedBeforeBlock := Move{
			From: kingPos,
			To:   Position{Row: 4, Col: 1},
		}

		blockedMove := Move{
			From: kingPos,
			To:   Position{Row: 2, Col: 3},
		}

		if !containsMove(moves, allowedBeforeBlock) {
			t.Fatalf("expected move before blocking piece %+v, got %+v", allowedBeforeBlock, moves)
		}

		if containsMove(moves, blockedMove) {
			t.Fatalf("expected king not to jump over piece with simple move, got %+v", moves)
		}
	})

	t.Run("simple move updates board", func(t *testing.T) {
		game := emptyGame(White)

		from := Position{Row: 5, Col: 0}
		to := Position{Row: 2, Col: 3}

		setPiece(&game.Board, from, White, true)

		result, err := game.ApplyMove(Move{From: from, To: to})
		if err != nil {
			t.Fatalf("expected king move to be valid, got %v", err)
		}

		if result.Captured {
			t.Fatal("expected simple king move not to capture")
		}

		if result.Promoted {
			t.Fatal("expected king move not to promote")
		}

		assertEmptySquare(t, game.Board, from)
		assertPiece(t, game.Board, to, White, true)

		if game.Turn != Black {
			t.Fatalf("expected turn to switch to black, got %q", game.Turn)
		}
	})

	t.Run("can capture at distance", func(t *testing.T) {
		game := emptyGame(White)

		kingPos := Position{Row: 5, Col: 0}
		blackPos := Position{Row: 3, Col: 2}

		setPiece(&game.Board, kingPos, White, true)
		setPiece(&game.Board, blackPos, Black, false)

		assertMoves(t, game.LegalMovesFrom(kingPos),
			Move{From: kingPos, To: Position{Row: 2, Col: 3}},
			Move{From: kingPos, To: Position{Row: 1, Col: 4}},
			Move{From: kingPos, To: Position{Row: 0, Col: 5}},
		)
	})

	t.Run("capture removes opponent piece", func(t *testing.T) {
		game := emptyGame(White)

		kingPos := Position{Row: 5, Col: 0}
		blackPos := Position{Row: 3, Col: 2}
		landing := Position{Row: 1, Col: 4}

		setPiece(&game.Board, kingPos, White, true)
		setPiece(&game.Board, blackPos, Black, false)

		result, err := game.ApplyMove(Move{From: kingPos, To: landing})
		if err != nil {
			t.Fatalf("expected king capture to be valid, got %v", err)
		}

		assertCaptured(t, result, blackPos)
		assertEmptySquare(t, game.Board, kingPos)
		assertEmptySquare(t, game.Board, blackPos)
		assertPiece(t, game.Board, landing, White, true)

		if game.Turn != Black {
			t.Fatalf("expected turn to switch to black, got %q", game.Turn)
		}
	})

	t.Run("cannot capture over own piece", func(t *testing.T) {
		game := emptyGame(White)

		kingPos := Position{Row: 5, Col: 0}
		ownPiecePos := Position{Row: 3, Col: 2}
		blackPos := Position{Row: 2, Col: 3}

		setPiece(&game.Board, kingPos, White, true)
		setPiece(&game.Board, ownPiecePos, White, false)
		setPiece(&game.Board, blackPos, Black, false)

		capture := Move{
			From: kingPos,
			To:   Position{Row: 1, Col: 4},
		}

		if moves := game.LegalMovesFrom(kingPos); containsMove(moves, capture) {
			t.Fatalf("expected king not to capture over own piece, got %+v", moves)
		}
	})

	t.Run("cannot capture two pieces in one move", func(t *testing.T) {
		game := emptyGame(White)

		kingPos := Position{Row: 5, Col: 0}
		firstBlack := Position{Row: 3, Col: 2}
		secondBlack := Position{Row: 1, Col: 4}

		setPiece(&game.Board, kingPos, White, true)
		setPiece(&game.Board, firstBlack, Black, false)
		setPiece(&game.Board, secondBlack, Black, false)

		moves := game.LegalMovesFrom(kingPos)

		validLanding := Move{
			From: kingPos,
			To:   Position{Row: 2, Col: 3},
		}

		invalidLandingAfterSecondPiece := Move{
			From: kingPos,
			To:   Position{Row: 0, Col: 5},
		}

		if !containsMove(moves, validLanding) {
			t.Fatalf("expected landing before second piece %+v, got %+v", validLanding, moves)
		}

		if containsMove(moves, invalidLandingAfterSecondPiece) {
			t.Fatalf("expected king not to capture two pieces in one move, got %+v", moves)
		}
	})

	t.Run("capture is mandatory over simple move", func(t *testing.T) {
		game := emptyGame(White)

		kingPos := Position{Row: 5, Col: 0}
		blackPos := Position{Row: 3, Col: 2}

		setPiece(&game.Board, kingPos, White, true)
		setPiece(&game.Board, blackPos, Black, false)

		_, err := game.ApplyMove(Move{
			From: kingPos,
			To:   Position{Row: 4, Col: 1},
		})
		if err != ErrInvalidMove {
			t.Fatalf("expected ErrInvalidMove for simple king move when capture is available, got %v", err)
		}
	})

	t.Run("multiple capture keeps turn and sets forced piece", func(t *testing.T) {
		game, start, afterFirstCapture, afterSecondCapture := newKingCaptureChain()

		result, err := game.ApplyMove(Move{From: start, To: afterFirstCapture})
		if err != nil {
			t.Fatalf("expected first king capture to be valid, got %v", err)
		}

		if !result.Captured {
			t.Fatal("expected first move to capture")
		}

		if !result.MustContinueCapture {
			t.Fatal("expected king to continue capture")
		}

		if game.Turn != White {
			t.Fatalf("expected turn to remain %q, got %q", White, game.Turn)
		}

		assertForcedPiece(t, game, afterFirstCapture)
		assertMoves(t, game.LegalMoves(), Move{
			From: afterFirstCapture,
			To:   afterSecondCapture,
		})
	})

	t.Run("multiple capture clears forced piece after final capture", func(t *testing.T) {
		game, start, afterFirstCapture, afterSecondCapture := newKingCaptureChain()
		firstBlack := Position{Row: 3, Col: 2}
		secondBlack := Position{Row: 1, Col: 4}

		_, err := game.ApplyMove(Move{From: start, To: afterFirstCapture})
		if err != nil {
			t.Fatalf("expected first king capture to be valid, got %v", err)
		}

		result, err := game.ApplyMove(Move{From: afterFirstCapture, To: afterSecondCapture})
		if err != nil {
			t.Fatalf("expected second king capture to be valid, got %v", err)
		}

		if !result.Captured {
			t.Fatal("expected second move to capture")
		}

		if result.MustContinueCapture {
			t.Fatal("expected capture chain to be finished")
		}

		if game.ForcedPiece != nil {
			t.Fatalf("expected forced piece to be cleared, got %+v", *game.ForcedPiece)
		}

		if game.Turn != Black {
			t.Fatalf("expected turn to switch to %q, got %q", Black, game.Turn)
		}

		assertEmptySquare(t, game.Board, firstBlack)
		assertEmptySquare(t, game.Board, secondBlack)
		assertPiece(t, game.Board, afterSecondCapture, White, true)
	})

	t.Run("multiple capture rejects other piece move", func(t *testing.T) {
		game, start, afterFirstCapture, _ := newKingCaptureChain()
		otherWhite := Position{Row: 5, Col: 6}
		setPiece(&game.Board, otherWhite, White, false)

		_, err := game.ApplyMove(Move{From: start, To: afterFirstCapture})
		if err != nil {
			t.Fatalf("expected first king capture to be valid, got %v", err)
		}

		_, err = game.ApplyMove(Move{
			From: otherWhite,
			To:   Position{Row: 4, Col: 7},
		})
		if err != ErrInvalidMove {
			t.Fatalf("expected ErrInvalidMove when moving other piece during king capture chain, got %v", err)
		}
	})
}

func TestGameFinish(t *testing.T) {
	t.Run("finishes when opponent has no pieces", func(t *testing.T) {
		game := emptyGame(White)

		whitePos := Position{Row: 4, Col: 1}
		blackPos := Position{Row: 3, Col: 2}
		landing := Position{Row: 2, Col: 3}

		setPiece(&game.Board, whitePos, White, false)
		setPiece(&game.Board, blackPos, Black, false)

		result, err := game.ApplyMove(Move{From: whitePos, To: landing})
		if err != nil {
			t.Fatalf("expected capture to be valid, got %v", err)
		}

		assertFinishedWithWinner(t, game, result, White)

		if game.Turn != Black {
			t.Fatalf("expected turn to be switched to losing player %q, got %q", Black, game.Turn)
		}
	})

	t.Run("finishes when opponent has no legal moves", func(t *testing.T) {
		game := emptyGame(White)

		whitePos := Position{Row: 5, Col: 0}
		blackPos := Position{Row: 7, Col: 2}

		setPiece(&game.Board, whitePos, White, false)
		setPiece(&game.Board, blackPos, Black, false)

		result, err := game.ApplyMove(Move{
			From: whitePos,
			To:   Position{Row: 4, Col: 1},
		})
		if err != nil {
			t.Fatalf("expected simple move to be valid, got %v", err)
		}

		assertFinishedWithWinner(t, game, result, White)
	})
}

func emptyGame(turn Color) *Game {
	game := NewGame()
	game.Board = Board{}
	game.Turn = turn
	game.Status = StatusActive
	game.Winner = nil
	game.ForcedPiece = nil

	return game
}

func setPiece(board *Board, pos Position, color Color, king bool) {
	board.SetPiece(pos, &Piece{
		Color: color,
		King:  king,
	})
}

func newManCaptureChain() (*Game, Position, Position, Position) {
	game := emptyGame(White)

	start := Position{Row: 5, Col: 0}
	firstBlack := Position{Row: 4, Col: 1}
	afterFirstCapture := Position{Row: 3, Col: 2}
	secondBlack := Position{Row: 2, Col: 3}
	afterSecondCapture := Position{Row: 1, Col: 4}

	setPiece(&game.Board, start, White, false)
	setPiece(&game.Board, firstBlack, Black, false)
	setPiece(&game.Board, secondBlack, Black, false)

	return game, start, afterFirstCapture, afterSecondCapture
}

func newKingCaptureChain() (*Game, Position, Position, Position) {
	game := emptyGame(White)

	start := Position{Row: 5, Col: 0}
	firstBlack := Position{Row: 3, Col: 2}
	afterFirstCapture := Position{Row: 2, Col: 3}
	secondBlack := Position{Row: 1, Col: 4}
	afterSecondCapture := Position{Row: 0, Col: 5}

	setPiece(&game.Board, start, White, true)
	setPiece(&game.Board, firstBlack, Black, false)
	setPiece(&game.Board, secondBlack, Black, false)

	return game, start, afterFirstCapture, afterSecondCapture
}

func newPromotionCaptureChain() (*Game, Position, Position, Position) {
	game := emptyGame(White)

	start := Position{Row: 2, Col: 1}
	firstBlack := Position{Row: 1, Col: 2}
	promotionSquare := Position{Row: 0, Col: 3}
	secondBlack := Position{Row: 2, Col: 5}
	finalLanding := Position{Row: 4, Col: 7}

	setPiece(&game.Board, start, White, false)
	setPiece(&game.Board, firstBlack, Black, false)
	setPiece(&game.Board, secondBlack, Black, false)

	return game, start, promotionSquare, finalLanding
}

func assertMoves(t *testing.T, moves []Move, expected ...Move) {
	t.Helper()

	if len(moves) != len(expected) {
		t.Fatalf("expected %d moves, got %d: %+v", len(expected), len(moves), moves)
	}

	for _, move := range expected {
		if !containsMove(moves, move) {
			t.Fatalf("expected move %+v, got %+v", move, moves)
		}
	}
}

func assertNoMoves(t *testing.T, moves []Move) {
	t.Helper()

	if len(moves) != 0 {
		t.Fatalf("expected no moves, got %+v", moves)
	}
}

func assertPiece(t *testing.T, board Board, pos Position, color Color, king bool) {
	t.Helper()

	piece := board.PieceAt(pos)
	if piece == nil {
		t.Fatalf("expected piece at %+v", pos)
	}

	if piece.Color != color {
		t.Fatalf("expected %q piece at %+v, got %q", color, pos, piece.Color)
	}

	if piece.King != king {
		t.Fatalf("expected king=%v at %+v, got %+v", king, pos, *piece)
	}
}

func assertEmptySquare(t *testing.T, board Board, pos Position) {
	t.Helper()

	if piece := board.PieceAt(pos); piece != nil {
		t.Fatalf("expected square %+v to be empty, got %+v", pos, *piece)
	}
}

func assertCaptured(t *testing.T, result MoveResult, captured Position) {
	t.Helper()

	if !result.Captured {
		t.Fatal("expected move to capture")
	}

	if result.CapturedPosition == nil {
		t.Fatal("expected captured position")
	}

	if *result.CapturedPosition != captured {
		t.Fatalf("expected captured position %+v, got %+v", captured, *result.CapturedPosition)
	}
}

func assertForcedPiece(t *testing.T, game *Game, expected Position) {
	t.Helper()

	if game.ForcedPiece == nil {
		t.Fatal("expected forced piece to be set")
	}

	if *game.ForcedPiece != expected {
		t.Fatalf("expected forced piece %+v, got %+v", expected, *game.ForcedPiece)
	}
}

func assertFinishedWithWinner(t *testing.T, game *Game, result MoveResult, winner Color) {
	t.Helper()

	if !result.GameFinished {
		t.Fatal("expected game to be finished")
	}

	if result.Winner == nil {
		t.Fatal("expected winner in result")
	}

	if *result.Winner != winner {
		t.Fatalf("expected winner %q, got %q", winner, *result.Winner)
	}

	if game.Status != StatusFinished {
		t.Fatalf("expected game status %q, got %q", StatusFinished, game.Status)
	}

	if game.Winner == nil {
		t.Fatal("expected game winner")
	}

	if *game.Winner != winner {
		t.Fatalf("expected game winner %q, got %q", winner, *game.Winner)
	}
}

func assertBoardsEqual(t *testing.T, before Board, after Board) {
	t.Helper()

	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			pos := Position{Row: row, Col: col}

			beforePiece := before.PieceAt(pos)
			afterPiece := after.PieceAt(pos)

			if beforePiece == nil && afterPiece == nil {
				continue
			}

			if beforePiece == nil || afterPiece == nil {
				t.Fatalf("expected board not to change at %+v", pos)
			}

			if *beforePiece != *afterPiece {
				t.Fatalf("expected board not to change at %+v, before %+v after %+v", pos, *beforePiece, *afterPiece)
			}
		}
	}
}
