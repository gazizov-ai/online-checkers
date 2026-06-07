import { ChevronRight, IdCard, RefreshCw } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { getMyProfile, profileName } from "../api/profile";
import { getRating } from "../api/rating";
import type { Profile, Rating, User } from "../types/domain";
import { friendlyApiError } from "../ui/labels";
import { ProfileAvatar } from "./ProfileAvatar";

type ProfilePanelProps = {
  onOpenProfile: (userId: string) => void;
  user: User;
};

export function ProfilePanel({ onOpenProfile, user }: ProfilePanelProps) {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [rating, setRating] = useState<Rating | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadProfile = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const [loadedProfile, loadedRating] = await Promise.all([
        getMyProfile(),
        getRating(user.id),
      ]);
      setProfile(loadedProfile);
      setRating(loadedRating);
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
        <button
          className="panel-title-link heading-with-icon"
          onClick={() => onOpenProfile(user.id)}
          type="button"
        >
          <IdCard aria-hidden size={19} />
          <h2>Профиль</h2>
          <ChevronRight aria-hidden size={17} />
        </button>
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

      {profile ? (
        <button
          className="profile-panel__identity"
          onClick={() => onOpenProfile(user.id)}
          type="button"
        >
          <ProfileAvatar profile={profile} />
          <span>
            <strong>{profileName(profile)}</strong>
            <small>@{profile.username}</small>
          </span>
          {profile.country_code ? <b>{profile.country_code}</b> : null}
        </button>
      ) : null}

      {profile?.bio ? <p className="profile-panel__bio">{profile.bio}</p> : null}
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
