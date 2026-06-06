import type { Color, GameSnapshot, Piece, Position } from "../types/domain";

export type LegalMove = {
  from: Position;
  to: Position;
  capture: boolean;
};

const BOARD_SIZE = 8;

const DIAGONALS = [
  { row: -1, col: -1 },
  { row: -1, col: 1 },
  { row: 1, col: -1 },
  { row: 1, col: 1 },
] as const;

export function samePosition(a?: Position | null, b?: Position | null): boolean {
  return Boolean(a && b && a.row === b.row && a.col === b.col);
}

export function positionKey(position: Position): string {
  return `${position.row}:${position.col}`;
}

function isValidPosition(position: Position): boolean {
  return (
    position.row >= 0 &&
    position.row < BOARD_SIZE &&
    position.col >= 0 &&
    position.col < BOARD_SIZE
  );
}

function isDarkSquare(position: Position): boolean {
  return isValidPosition(position) && (position.row + position.col) % 2 === 1;
}

function pieceAt(snapshot: GameSnapshot, position: Position): Piece | null {
  if (!isValidPosition(position)) {
    return null;
  }

  return snapshot.board.cells[position.row]?.[position.col] ?? null;
}

function positionsEqual(a: Position, b: Position): boolean {
  return a.row === b.row && a.col === b.col;
}

function makeMove(from: Position, to: Position, capture: boolean): LegalMove {
  return { from, to, capture };
}

export function getLegalMoves(snapshot: GameSnapshot, color: Color): LegalMove[] {
  if (snapshot.status === "finished") {
    return [];
  }

  if (snapshot.forced_piece) {
    return getCapturesFrom(snapshot, snapshot.forced_piece);
  }

  const captures = getAllCaptures(snapshot, color);
  if (captures.length > 0) {
    return captures;
  }

  return getAllSimpleMoves(snapshot, color);
}

export function getLegalMovesFrom(snapshot: GameSnapshot, position: Position): LegalMove[] {
  if (snapshot.status === "finished" || !isValidPosition(position)) {
    return [];
  }

  if (snapshot.forced_piece) {
    if (!positionsEqual(snapshot.forced_piece, position)) {
      return [];
    }

    return getCapturesFrom(snapshot, position);
  }

  const piece = pieceAt(snapshot, position);
  if (!piece || piece.color !== snapshot.turn) {
    return [];
  }

  const captures = getAllCaptures(snapshot, snapshot.turn);
  if (captures.length > 0) {
    return getCapturesFrom(snapshot, position);
  }

  return getSimpleMovesFrom(snapshot, position);
}

export function getSelectablePositions(snapshot: GameSnapshot, color: Color): Set<string> {
  const positions = new Set<string>();

  for (const move of getLegalMoves(snapshot, color)) {
    positions.add(positionKey(move.from));
  }

  return positions;
}

function getAllSimpleMoves(snapshot: GameSnapshot, color: Color): LegalMove[] {
  const moves: LegalMove[] = [];

  forEachBoardPosition((position) => {
    const piece = pieceAt(snapshot, position);
    if (piece?.color === color) {
      moves.push(...getSimpleMovesFrom(snapshot, position));
    }
  });

  return moves;
}

function getSimpleMovesFrom(snapshot: GameSnapshot, position: Position): LegalMove[] {
  const piece = pieceAt(snapshot, position);
  if (!piece) {
    return [];
  }

  return piece.king
    ? getKingSimpleMoves(snapshot, position)
    : getManSimpleMoves(snapshot, position, piece.color);
}

function getManSimpleMoves(
  snapshot: GameSnapshot,
  position: Position,
  color: Color,
): LegalMove[] {
  const rowDirection = color === "white" ? -1 : 1;
  const candidates = [
    { row: position.row + rowDirection, col: position.col - 1 },
    { row: position.row + rowDirection, col: position.col + 1 },
  ];

  return candidates
    .filter((to) => isDarkSquare(to) && !pieceAt(snapshot, to))
    .map((to) => makeMove(position, to, false));
}

function getKingSimpleMoves(snapshot: GameSnapshot, position: Position): LegalMove[] {
  const moves: LegalMove[] = [];

  for (const direction of DIAGONALS) {
    let current = {
      row: position.row + direction.row,
      col: position.col + direction.col,
    };

    while (isValidPosition(current)) {
      if (!isDarkSquare(current) || pieceAt(snapshot, current)) {
        break;
      }

      moves.push(makeMove(position, current, false));
      current = {
        row: current.row + direction.row,
        col: current.col + direction.col,
      };
    }
  }

  return moves;
}

function getAllCaptures(snapshot: GameSnapshot, color: Color): LegalMove[] {
  const moves: LegalMove[] = [];

  forEachBoardPosition((position) => {
    const piece = pieceAt(snapshot, position);
    if (piece?.color === color) {
      moves.push(...getCapturesFrom(snapshot, position));
    }
  });

  return moves;
}

function getCapturesFrom(snapshot: GameSnapshot, position: Position): LegalMove[] {
  const piece = pieceAt(snapshot, position);
  if (!piece) {
    return [];
  }

  return piece.king
    ? getKingCaptures(snapshot, position, piece.color)
    : getManCaptures(snapshot, position, piece.color);
}

function getManCaptures(snapshot: GameSnapshot, position: Position, color: Color): LegalMove[] {
  const moves: LegalMove[] = [];

  for (const direction of DIAGONALS) {
    const middle = {
      row: position.row + direction.row,
      col: position.col + direction.col,
    };
    const to = {
      row: position.row + 2 * direction.row,
      col: position.col + 2 * direction.col,
    };

    if (!isDarkSquare(to) || pieceAt(snapshot, to)) {
      continue;
    }

    const middlePiece = pieceAt(snapshot, middle);
    if (middlePiece && middlePiece.color !== color) {
      moves.push(makeMove(position, to, true));
    }
  }

  return moves;
}

function getKingCaptures(snapshot: GameSnapshot, position: Position, color: Color): LegalMove[] {
  const moves: LegalMove[] = [];

  for (const direction of DIAGONALS) {
    let current = {
      row: position.row + direction.row,
      col: position.col + direction.col,
    };
    let foundOpponent = false;

    while (isValidPosition(current)) {
      const currentPiece = pieceAt(snapshot, current);

      if (!foundOpponent) {
        if (!currentPiece) {
          current = {
            row: current.row + direction.row,
            col: current.col + direction.col,
          };
          continue;
        }

        if (currentPiece.color === color) {
          break;
        }

        foundOpponent = true;
        current = {
          row: current.row + direction.row,
          col: current.col + direction.col,
        };
        continue;
      }

      if (currentPiece) {
        break;
      }

      moves.push(makeMove(position, current, true));
      current = {
        row: current.row + direction.row,
        col: current.col + direction.col,
      };
    }
  }

  return moves;
}

function forEachBoardPosition(callback: (position: Position) => void): void {
  for (let row = 0; row < BOARD_SIZE; row += 1) {
    for (let col = 0; col < BOARD_SIZE; col += 1) {
      callback({ row, col });
    }
  }
}
