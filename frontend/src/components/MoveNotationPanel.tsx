import { ListOrdered } from "lucide-react";
import { useEffect, useMemo, useRef } from "react";
import type { MoveHistoryItem } from "../types/domain";

type MoveNotationPanelProps = {
  history: MoveHistoryItem[];
  whitePlayerId?: string;
  blackPlayerId?: string;
  height: number;
};

type MoveRow = {
  number: number;
  white?: string;
  black?: string;
};

export function MoveNotationPanel({
  history,
  whitePlayerId,
  blackPlayerId,
  height,
}: MoveNotationPanelProps) {
  const listRef = useRef<HTMLDivElement | null>(null);
  const rows = useMemo(() => {
    const moveRows = new Map<number, MoveRow>();

    for (const item of history) {
      const rowNumber = Math.floor((item.turn_number - 1) / 2) + 1;
      const row = moveRows.get(rowNumber) ?? { number: rowNumber };

      if (item.player_id === whitePlayerId) {
        row.white = item.notation;
      } else if (item.player_id === blackPlayerId) {
        row.black = item.notation;
      }

      moveRows.set(rowNumber, row);
    }

    return Array.from(moveRows.values()).sort((a, b) => a.number - b.number);
  }, [blackPlayerId, history, whitePlayerId]);

  useEffect(() => {
    const list = listRef.current;
    if (list) {
      list.scrollTop = list.scrollHeight;
    }
  }, [history]);

  return (
    <aside className="notation-panel" style={{ height }}>
      <div className="heading-with-icon">
        <ListOrdered aria-hidden size={19} />
        <h2>Ходы</h2>
      </div>

      <div className="notation-list" ref={listRef}>
        {rows.length > 0 ? (
          rows.map((row) => (
            <div className="notation-row" key={row.number}>
              <span className="notation-number">{row.number}.</span>
              <span>{row.white ?? ""}</span>
              <span>{row.black ?? ""}</span>
            </div>
          ))
        ) : (
          <span className="notation-empty">Ходы появятся здесь после начала партии.</span>
        )}
      </div>
    </aside>
  );
}
