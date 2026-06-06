import { Crown } from "lucide-react";
import type { Color, GameSnapshot, Position } from "../types/domain";
import { positionKey } from "../game/legalMoves";

type BoardProps = {
  snapshot: GameSnapshot;
  playerColor?: Color;
  legalTargets?: Set<string>;
  captureTargets?: Set<string>;
  disabled?: boolean;
  onSquareClick: (position: Position) => void;
};

function displayIndex(index: number, playerColor?: Color): number {
  return playerColor === "black" ? 7 - index : index;
}

export function Board({
  snapshot,
  playerColor,
  legalTargets = new Set(),
  captureTargets = new Set(),
  disabled = false,
  onSquareClick,
}: BoardProps) {
  return (
    <div className="board-shell" aria-label="Доска">
      <div className="board">
        {Array.from({ length: 8 }).map((_, displayRow) =>
          Array.from({ length: 8 }).map((__, displayCol) => {
            const row = displayIndex(displayRow, playerColor);
            const col = displayIndex(displayCol, playerColor);
            const position = { row, col };
            const piece = snapshot.board.cells[row]?.[col] ?? null;
            const isDark = (row + col) % 2 === 1;
            const key = positionKey(position);
            const isLegalTarget = legalTargets.has(key);
            const isCaptureTarget = captureTargets.has(key);

            return (
              <button
                aria-label={`row ${row + 1}, column ${col + 1}`}
                className={[
                  "square",
                  isDark ? "square--dark" : "square--light",
                  isLegalTarget ? "square--legal-target" : "",
                  isCaptureTarget ? "square--capture-target" : "",
                ].join(" ")}
                disabled={disabled}
                key={`${row}-${col}`}
                onClick={() => onSquareClick(position)}
                type="button"
              >
                {piece ? (
                  <span className={`piece piece--${piece.color}`}>
                    {piece.king ? <Crown aria-hidden size={18} strokeWidth={2.4} /> : null}
                  </span>
                ) : null}
              </button>
            );
          }),
        )}
      </div>
    </div>
  );
}
