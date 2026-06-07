import { Crown } from "lucide-react";
import type { Color, GameSnapshot, Position } from "../types/domain";

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

function positionKey(position: Position): string {
  return `${position.row}:${position.col}`;
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
            const fileLabel = String.fromCharCode("a".charCodeAt(0) + col);
            const rankLabel = String(8 - row);
            const coordinateTone = isDark ? "coordinate--on-dark" : "coordinate--on-light";

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
                {displayRow === 7 ? (
                  <span
                    aria-hidden
                    className={`coordinate coordinate--file ${coordinateTone}`}
                  >
                    {fileLabel}
                  </span>
                ) : null}
                {displayCol === 7 ? (
                  <span
                    aria-hidden
                    className={`coordinate coordinate--rank ${coordinateTone}`}
                  >
                    {rankLabel}
                  </span>
                ) : null}
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
