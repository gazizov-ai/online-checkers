import { ListOrdered } from "lucide-react";

export function MoveNotationPanel() {
  return (
    <aside className="notation-panel">
      <div className="heading-with-icon">
        <ListOrdered aria-hidden size={19} />
        <h2>Ходы</h2>
      </div>

      <div className="notation-placeholder">
        <span>Нотация появится здесь после подключения данных с бэка.</span>
      </div>
    </aside>
  );
}
