import { IdCard, RefreshCw } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { getRating } from "../api/rating";
import type { Rating, User } from "../types/domain";
import { friendlyApiError } from "../ui/labels";

export function ProfilePanel({ user }: { user: User }) {
  const [rating, setRating] = useState<Rating | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadProfile = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      setRating(await getRating(user.id));
    } catch (err) {
      setError(friendlyApiError(err, "Не удалось загрузить профиль."));
    } finally {
      setLoading(false);
    }
  }, [user.id]);

  useEffect(() => {
    void loadProfile();
  }, [loadProfile]);

  return (
    <aside className="profile-panel">
      <div className="panel-heading">
        <div className="heading-with-icon">
          <IdCard aria-hidden size={19} />
          <h2>Профиль</h2>
        </div>
        <button
          aria-label="Обновить профиль"
          className="icon-button"
          disabled={loading}
          onClick={() => void loadProfile()}
          title="Обновить профиль"
          type="button"
        >
          <RefreshCw aria-hidden size={18} />
        </button>
      </div>

      {error ? <p className="inline-error">{error}</p> : null}

      <dl className="profile-stats">
        <div>
          <dt>Рейтинг</dt>
          <dd>{rating?.rating ?? 1000}</dd>
        </div>
        <div>
          <dt>Игр</dt>
          <dd>{rating?.games_played ?? 0}</dd>
        </div>
        <div>
          <dt>Побед</dt>
          <dd>{rating?.wins ?? 0}</dd>
        </div>
        <div>
          <dt>Поражений</dt>
          <dd>{rating?.losses ?? 0}</dd>
        </div>
      </dl>
    </aside>
  );
}
