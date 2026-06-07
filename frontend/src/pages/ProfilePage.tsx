import {
  ArrowLeft,
  CalendarDays,
  Check,
  CircleAlert,
  Clock3,
  Edit3,
  ExternalLink,
  Gamepad2,
  Globe2,
  Save,
  Shield,
  Swords,
  UserRound,
  X,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useState, type FormEvent } from "react";
import { getUserGameHistory } from "../api/game";
import {
  getMyProfile,
  getProfile,
  getProfiles,
  profileName,
  updateMyProfile,
} from "../api/profile";
import { getRating } from "../api/rating";
import { ProfileAvatar } from "../components/ProfileAvatar";
import { StatusPill } from "../components/StatusPill";
import type {
  Profile,
  Rating,
  UpdateProfileInput,
  User,
  UserGameHistoryItem,
} from "../types/domain";
import { friendlyApiError } from "../ui/labels";

type ProfilePageProps = {
  currentUser: User;
  onBack: () => void;
  onOpenGame: (gameId: string) => void;
  onOpenProfile: (userId: string) => void;
  userId: string;
};

type Outcome = "active" | "win" | "loss" | "draw";

const dateFormatter = new Intl.DateTimeFormat("ru-RU", {
  day: "numeric",
  month: "short",
  year: "numeric",
});

function formatDate(value?: string): string {
  return value ? dateFormatter.format(new Date(value)) : "-";
}

function outcomeFor(game: UserGameHistoryItem): Outcome {
  if (game.status === "active") {
    return "active";
  }
  if (game.result === "draw") {
    return "draw";
  }

  const userWon =
    (game.user_color === "white" && game.result === "white_win") ||
    (game.user_color === "black" && game.result === "black_win");
  return userWon ? "win" : "loss";
}

function outcomeLabel(outcome: Outcome): string {
  switch (outcome) {
    case "active":
      return "Идет игра";
    case "win":
      return "Победа";
    case "loss":
      return "Поражение";
    case "draw":
      return "Ничья";
  }
}

function finishReasonLabel(game: UserGameHistoryItem): string {
  switch (game.finish_reason) {
    case "resignation":
      return "Сдача";
    case "draw_agreement":
      return "По соглашению";
    case "checkers_rules":
      return "По правилам";
    default:
      return game.status === "active" ? "Партия продолжается" : "Завершена";
  }
}

export function ProfilePage({
  currentUser,
  onBack,
  onOpenGame,
  onOpenProfile,
  userId,
}: ProfilePageProps) {
  const isOwnProfile = currentUser.id === userId;
  const [profile, setProfile] = useState<Profile | null>(null);
  const [rating, setRating] = useState<Rating | null>(null);
  const [games, setGames] = useState<UserGameHistoryItem[]>([]);
  const [profiles, setProfiles] = useState<Record<string, Profile>>({});
  const [nextCursor, setNextCursor] = useState<string | undefined>();
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadOpponentProfiles = useCallback(
    async (items: UserGameHistoryItem[]) => {
      const ids = items.map((game) =>
        game.white_player_id === userId ? game.black_player_id : game.white_player_id,
      );
      try {
        const loadedProfiles = await getProfiles(ids);
        setProfiles((current) => ({
          ...current,
          ...Object.fromEntries(loadedProfiles.map((item) => [item.user_id, item])),
        }));
      } catch {
        // Game history remains useful even if profile enrichment is temporarily unavailable.
      }
    },
    [userId],
  );

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    setGames([]);
    setProfiles({});

    void Promise.all([
      isOwnProfile ? getMyProfile() : getProfile(userId),
      getRating(userId),
      getUserGameHistory(userId),
    ])
      .then(async ([loadedProfile, loadedRating, history]) => {
        if (cancelled) {
          return;
        }
        setProfile(loadedProfile);
        setRating(loadedRating);
        setGames(history.items);
        setNextCursor(history.next_cursor);
        await loadOpponentProfiles(history.items);
      })
      .catch((err) => {
        if (!cancelled) {
          setError(friendlyApiError(err, "Не удалось загрузить профиль."));
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [isOwnProfile, loadOpponentProfiles, userId]);

  const stats = useMemo(
    () => [
      ["Рейтинг", rating?.rating ?? 1000],
      ["Игр", rating?.games_played ?? 0],
      ["Побед", rating?.wins ?? 0],
      ["Поражений", rating?.losses ?? 0],
    ],
    [rating],
  );

  async function loadMoreGames() {
    if (!nextCursor || loadingMore) {
      return;
    }

    setLoadingMore(true);
    setError(null);
    try {
      const history = await getUserGameHistory(userId, nextCursor);
      setGames((current) => [...current, ...history.items]);
      setNextCursor(history.next_cursor);
      await loadOpponentProfiles(history.items);
    } catch (err) {
      setError(friendlyApiError(err, "Не удалось продолжить загрузку истории."));
    } finally {
      setLoadingMore(false);
    }
  }

  return (
    <section className="profile-page">
      <div className="profile-page__topbar">
        <button className="ghost-button" onClick={onBack} type="button">
          <ArrowLeft aria-hidden size={18} />
          Назад
        </button>
      </div>

      {loading ? (
        <div className="profile-page__loading">
          <span className="loader" />
        </div>
      ) : profile ? (
        <>
          <section className="profile-overview">
            <div className="profile-overview__identity">
              <ProfileAvatar profile={profile} size="large" />
              <div>
                <div className="profile-overview__name">
                  <h2>{profileName(profile)}</h2>
                  {isOwnProfile ? <StatusPill tone="good">Это вы</StatusPill> : null}
                </div>
                <p>@{profile.username}</p>
              </div>
            </div>

            <div className="profile-overview__facts">
              <span>
                <Globe2 aria-hidden size={17} />
                {profile.country_code ?? "Страна не указана"}
              </span>
              <span>
                <CalendarDays aria-hidden size={17} />
                В игре с {formatDate(profile.created_at)}
              </span>
            </div>

            <p className={profile.bio ? "profile-overview__bio" : "profile-overview__bio is-empty"}>
              {profile.bio ?? "Пользователь пока ничего о себе не рассказал."}
            </p>
          </section>

          <dl className="profile-page__stats">
            {stats.map(([label, value]) => (
              <div key={label}>
                <dt>{label}</dt>
                <dd>{value}</dd>
              </div>
            ))}
          </dl>

          <div className="profile-page__columns">
            <GameHistory
              games={games}
              loadingMore={loadingMore}
              nextCursor={nextCursor}
              onLoadMore={() => void loadMoreGames()}
              onOpenGame={onOpenGame}
              onOpenProfile={onOpenProfile}
              profiles={profiles}
              viewerId={currentUser.id}
              userId={userId}
            />

            <aside className="profile-details">
              {isOwnProfile ? (
                <EditProfileForm onUpdated={setProfile} profile={profile} />
              ) : (
                <>
                  <div className="panel-heading">
                    <div className="heading-with-icon">
                      <UserRound aria-hidden size={19} />
                      <h2>О профиле</h2>
                    </div>
                  </div>
                  <dl className="profile-detail-list">
                    <div>
                      <dt>Имя пользователя</dt>
                      <dd>@{profile.username}</dd>
                    </div>
                    <div>
                      <dt>Отображаемое имя</dt>
                      <dd>{profile.display_name ?? "Не указано"}</dd>
                    </div>
                    <div>
                      <dt>Страна</dt>
                      <dd>{profile.country_code ?? "Не указана"}</dd>
                    </div>
                    <div>
                      <dt>Профиль обновлен</dt>
                      <dd>{formatDate(profile.updated_at)}</dd>
                    </div>
                  </dl>
                </>
              )}
            </aside>
          </div>
        </>
      ) : null}

      {error ? (
        <p className="inline-error inline-error--with-icon">
          <CircleAlert aria-hidden size={17} />
          {error}
        </p>
      ) : null}
    </section>
  );
}

type GameHistoryProps = {
  games: UserGameHistoryItem[];
  loadingMore: boolean;
  nextCursor?: string;
  onLoadMore: () => void;
  onOpenGame: (gameId: string) => void;
  onOpenProfile: (userId: string) => void;
  profiles: Record<string, Profile>;
  userId: string;
  viewerId: string;
};

function GameHistory({
  games,
  loadingMore,
  nextCursor,
  onLoadMore,
  onOpenGame,
  onOpenProfile,
  profiles,
  userId,
  viewerId,
}: GameHistoryProps) {
  return (
    <section className="game-history">
      <div className="panel-heading">
        <div className="heading-with-icon">
          <Gamepad2 aria-hidden size={19} />
          <h2>История игр</h2>
        </div>
        <span className="section-count">{games.length}</span>
      </div>

      {games.length === 0 ? (
        <div className="history-empty">
          <Swords aria-hidden size={24} />
          <span>Сыгранных партий пока нет.</span>
        </div>
      ) : (
        <div className="history-list">
          {games.map((game) => {
            const opponentId =
              game.white_player_id === userId ? game.black_player_id : game.white_player_id;
            const opponent = profiles[opponentId];
            const outcome = outcomeFor(game);
            const canOpenGame =
              game.white_player_id === viewerId || game.black_player_id === viewerId;

            return (
              <article className="history-item" key={game.game_id}>
                <span className={`history-result history-result--${outcome}`}>
                  {outcome === "win" ? (
                    <Check aria-hidden size={18} />
                  ) : outcome === "loss" ? (
                    <X aria-hidden size={18} />
                  ) : outcome === "active" ? (
                    <Clock3 aria-hidden size={18} />
                  ) : (
                    <Shield aria-hidden size={18} />
                  )}
                </span>

                <div className="history-item__main">
                  <span className="history-item__label">{outcomeLabel(outcome)}</span>
                  <button
                    className="text-link"
                    onClick={() => onOpenProfile(opponentId)}
                    type="button"
                  >
                    {opponent ? profileName(opponent) : "Игрок"}
                  </button>
                  <small>
                    {game.user_color === "white" ? "Белые" : "Черные"} ·{" "}
                    {finishReasonLabel(game)}
                  </small>
                </div>

                <div className="history-item__meta">
                  <time dateTime={game.created_at}>{formatDate(game.created_at)}</time>
                  {canOpenGame ? (
                    <button
                      aria-label="Открыть партию"
                      className="icon-button history-open-button"
                      onClick={() => onOpenGame(game.game_id)}
                      title="Открыть партию"
                      type="button"
                    >
                      <ExternalLink aria-hidden size={17} />
                    </button>
                  ) : null}
                </div>
              </article>
            );
          })}
        </div>
      )}

      {nextCursor ? (
        <button
          className="ghost-button history-more"
          disabled={loadingMore}
          onClick={onLoadMore}
          type="button"
        >
          {loadingMore ? "Загружаем..." : "Показать еще"}
        </button>
      ) : null}
    </section>
  );
}

type EditProfileFormProps = {
  onUpdated: (profile: Profile) => void;
  profile: Profile;
};

function EditProfileForm({ onUpdated, profile }: EditProfileFormProps) {
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [form, setForm] = useState<UpdateProfileInput>({
    display_name: profile.display_name ?? "",
    country_code: profile.country_code ?? "",
    avatar_url: profile.avatar_url ?? "",
    bio: profile.bio ?? "",
  });

  useEffect(() => {
    setForm({
      display_name: profile.display_name ?? "",
      country_code: profile.country_code ?? "",
      avatar_url: profile.avatar_url ?? "",
      bio: profile.bio ?? "",
    });
  }, [profile]);

  function updateField(field: keyof UpdateProfileInput, value: string) {
    setForm((current) => ({ ...current, [field]: value }));
    setSaved(false);
  }

  function cancelEditing() {
    setForm({
      display_name: profile.display_name ?? "",
      country_code: profile.country_code ?? "",
      avatar_url: profile.avatar_url ?? "",
      bio: profile.bio ?? "",
    });
    setError(null);
    setEditing(false);
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    setSaved(false);
    setError(null);

    try {
      const updated = await updateMyProfile(form);
      onUpdated(updated);
      setEditing(false);
      setSaved(true);
    } catch (err) {
      setError(friendlyApiError(err, "Не удалось сохранить профиль."));
    } finally {
      setSaving(false);
    }
  }

  if (!editing) {
    return (
      <>
        <div className="panel-heading">
          <div className="heading-with-icon">
            <UserRound aria-hidden size={19} />
            <h2>Мои данные</h2>
          </div>
          <button
            aria-label="Редактировать профиль"
            className="icon-button"
            onClick={() => {
              setEditing(true);
              setSaved(false);
            }}
            title="Редактировать профиль"
            type="button"
          >
            <Edit3 aria-hidden size={17} />
          </button>
        </div>
        <dl className="profile-detail-list">
          <div>
            <dt>Имя пользователя</dt>
            <dd>@{profile.username}</dd>
          </div>
          <div>
            <dt>Отображаемое имя</dt>
            <dd>{profile.display_name ?? "Не указано"}</dd>
          </div>
          <div>
            <dt>Страна</dt>
            <dd>{profile.country_code ?? "Не указана"}</dd>
          </div>
          <div>
            <dt>Профиль обновлен</dt>
            <dd>{formatDate(profile.updated_at)}</dd>
          </div>
        </dl>
        {saved ? <p className="form-success">Профиль сохранен.</p> : null}
      </>
    );
  }

  return (
    <form className="profile-edit-form" onSubmit={(event) => void submit(event)}>
      <div className="panel-heading">
        <div className="heading-with-icon">
          <Edit3 aria-hidden size={19} />
          <h2>Редактирование</h2>
        </div>
      </div>

      <label>
        Отображаемое имя
        <input
          maxLength={40}
          onChange={(event) => updateField("display_name", event.target.value)}
          placeholder={profile.username}
          value={form.display_name}
        />
      </label>
      <label>
        Страна
        <input
          maxLength={2}
          onChange={(event) => updateField("country_code", event.target.value.toUpperCase())}
          placeholder="RU"
          value={form.country_code}
        />
      </label>
      <label>
        Ссылка на аватар
        <input
          onChange={(event) => updateField("avatar_url", event.target.value)}
          placeholder="https://..."
          type="url"
          value={form.avatar_url}
        />
      </label>
      <label>
        О себе
        <textarea
          maxLength={300}
          onChange={(event) => updateField("bio", event.target.value)}
          rows={5}
          value={form.bio}
        />
        <small>{form.bio.length}/300</small>
      </label>

      {error ? <p className="form-error">{error}</p> : null}

      <div className="profile-edit-form__actions">
        <button className="primary-button" disabled={saving} type="submit">
          <Save aria-hidden size={17} />
          {saving ? "Сохраняем..." : "Сохранить"}
        </button>
        <button
          className="ghost-button"
          disabled={saving}
          onClick={cancelEditing}
          type="button"
        >
          <X aria-hidden size={17} />
          Отмена
        </button>
      </div>
    </form>
  );
}
