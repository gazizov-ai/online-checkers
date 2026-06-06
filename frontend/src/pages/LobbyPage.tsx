import { Search, X } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  cancelMatchmaking,
  getMatchmakingStatus,
  searchMatch,
} from "../api/matchmaking";
import { StatusPill } from "../components/StatusPill";
import type { MatchmakingResponse, User } from "../types/domain";
import { friendlyApiError, matchmakingStatusLabel } from "../ui/labels";

type LobbyPageProps = {
  user: User;
  onGameFound: (gameId: string) => void;
  autoStartSearch?: boolean;
  onAutoStartConsumed?: () => void;
};

export function LobbyPage({
  user,
  onGameFound,
  autoStartSearch = false,
  onAutoStartConsumed,
}: LobbyPageProps) {
  const [matchmaking, setMatchmaking] = useState<MatchmakingResponse | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pollingRef = useRef<number | null>(null);

  const stopPolling = useCallback(() => {
    if (pollingRef.current !== null) {
      window.clearInterval(pollingRef.current);
      pollingRef.current = null;
    }
  }, []);

  const handleMatchmakingResponse = useCallback(
    (response: MatchmakingResponse) => {
      setMatchmaking(response);

      if (response.game_id) {
        stopPolling();
        onGameFound(response.game_id);
      }
    },
    [onGameFound, stopPolling],
  );

  const pollStatus = useCallback(() => {
    stopPolling();
    pollingRef.current = window.setInterval(() => {
      void getMatchmakingStatus()
        .then(handleMatchmakingResponse)
        .catch((err) => setError(friendlyApiError(err, "Не удалось обновить статус поиска.")));
    }, 1600);
  }, [handleMatchmakingResponse, stopPolling]);

  useEffect(() => () => stopPolling(), [stopPolling]);

  const searching = pollingRef.current !== null;

  const startSearch = useCallback(async () => {
    setBusy(true);
    setError(null);

    try {
      const response = await searchMatch();
      handleMatchmakingResponse(response);

      if (!response.game_id) {
        pollStatus();
      }
    } catch (err) {
      setError(friendlyApiError(err, "Не удалось начать поиск игры."));
    } finally {
      setBusy(false);
    }
  }, [handleMatchmakingResponse, pollStatus]);

  useEffect(() => {
    if (!autoStartSearch || busy || searching) {
      return;
    }

    onAutoStartConsumed?.();
    void startSearch();
  }, [autoStartSearch, busy, onAutoStartConsumed, searching, startSearch]);

  async function cancelSearch() {
    setBusy(true);
    setError(null);

    try {
      await cancelMatchmaking();
      setMatchmaking(null);
      stopPolling();
    } catch (err) {
      setError(friendlyApiError(err, "Не удалось отменить поиск."));
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="lobby-panel">
      <div className="panel-heading">
        <div>
          <h2>Лобби</h2>
          <p>{user.username}</p>
        </div>
        {matchmaking ? (
          <StatusPill tone={searching ? "warn" : "neutral"}>
            {matchmakingStatusLabel(matchmaking.status)}
          </StatusPill>
        ) : null}
      </div>

      <div className="action-row">
        <button className="primary-button" disabled={busy || searching} onClick={() => void startSearch()} type="button">
          <Search aria-hidden size={18} />
          Найти игру
        </button>
        <button className="ghost-button" disabled={busy || !searching} onClick={() => void cancelSearch()} type="button">
          <X aria-hidden size={18} />
          Отмена
        </button>
      </div>

      {searching ? (
        <div className="search-progress" role="status">
          <span className="search-spinner" />
          <div>
            <strong>Ищем соперника</strong>
            <p>Партия откроется автоматически.</p>
          </div>
        </div>
      ) : null}

      {error ? <p className="form-error">{error}</p> : null}
    </section>
  );
}
