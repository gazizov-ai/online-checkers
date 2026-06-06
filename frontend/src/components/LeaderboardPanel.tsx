import { RefreshCw, Trophy } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { getLeaderboard, getRating } from "../api/rating";
import type { Rating, User } from "../types/domain";
import { friendlyApiError } from "../ui/labels";
import { StatusPill } from "./StatusPill";

export function LeaderboardPanel({ user }: { user: User }) {
  const [items, setItems] = useState<Rating[]>([]);
  const [myRating, setMyRating] = useState<Rating | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadRatings = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const [leaderboard, rating] = await Promise.all([
        getLeaderboard(10),
        getRating(user.id),
      ]);
      setItems(leaderboard);
      setMyRating(rating);
    } catch (err) {
      setError(friendlyApiError(err, "Не удалось загрузить рейтинг."));
    } finally {
      setLoading(false);
    }
  }, [user.id]);

  useEffect(() => {
    void loadRatings();
  }, [loadRatings]);

  return (
    <aside className="side-panel">
      <div className="panel-heading">
        <div className="heading-with-icon">
          <Trophy aria-hidden size={19} />
          <h2>Рейтинг</h2>
        </div>
        <button
          aria-label="Обновить рейтинг"
          className="icon-button"
          disabled={loading}
          onClick={() => void loadRatings()}
          title="Обновить рейтинг"
          type="button"
        >
          <RefreshCw aria-hidden size={18} />
        </button>
      </div>

      <div className="my-rating">
        <span>Мой счет</span>
        <strong>{myRating?.rating ?? 1000}</strong>
        <StatusPill tone={myRating ? "good" : "neutral"}>
          {myRating ? `${myRating.wins}-${myRating.losses}` : "0-0"}
        </StatusPill>
      </div>

      {error ? <p className="inline-error">{error}</p> : null}

      <ol className="leaderboard-list">
        {items.map((item, index) => (
          <li className={item.user_id === user.id ? "is-current-user" : ""} key={item.user_id}>
            <span className="rank">{index + 1}</span>
            <span className="player-id">{item.user_id.slice(0, 8)}</span>
            <strong>{item.rating}</strong>
          </li>
        ))}
      </ol>
    </aside>
  );
}
