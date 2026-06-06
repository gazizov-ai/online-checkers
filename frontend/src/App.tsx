import { LogOut } from "lucide-react";
import { useState } from "react";
import { useAuth } from "./auth/AuthContext";
import { LeaderboardPanel } from "./components/LeaderboardPanel";
import { ProfilePanel } from "./components/ProfilePanel";
import { AuthPage } from "./pages/AuthPage";
import { GamePage } from "./pages/GamePage";
import { LobbyPage } from "./pages/LobbyPage";

export function App() {
  const { user, loading, logout } = useAuth();
  const [activeGameId, setActiveGameId] = useState<string | null>(null);
  const [autoStartSearch, setAutoStartSearch] = useState(false);

  if (loading) {
    return (
      <main className="loading-screen">
        <span className="loader" />
      </main>
    );
  }

  if (!user) {
    return <AuthPage />;
  }

  return (
    <main className="app-shell">
      <header className="app-header">
        <div className="brand-mark brand-mark--compact">
          <span className="brand-dot brand-dot--white" />
          <span className="brand-dot brand-dot--black" />
          <div>
            <h1>Online Checkers</h1>
            <p>{user.username}</p>
          </div>
        </div>

        <button
          aria-label="Выйти"
          className="icon-button"
          onClick={() => {
            setActiveGameId(null);
            setAutoStartSearch(false);
            logout();
          }}
          title="Выйти"
          type="button"
        >
          <LogOut aria-hidden size={18} />
        </button>
      </header>

      {activeGameId ? (
        <div className="workspace">
          <section className="primary-workspace">
            <GamePage
              gameId={activeGameId}
              onFindNewGame={() => {
                setAutoStartSearch(true);
                setActiveGameId(null);
              }}
              onLeave={() => setActiveGameId(null)}
              user={user}
            />
          </section>
        </div>
      ) : (
        <div className="dashboard-stack">
          <section className="primary-workspace">
            <LobbyPage
              autoStartSearch={autoStartSearch}
              onAutoStartConsumed={() => setAutoStartSearch(false)}
              onGameFound={(gameId) => {
                setAutoStartSearch(false);
                setActiveGameId(gameId);
              }}
              user={user}
            />
          </section>
          <LeaderboardPanel user={user} />
          <ProfilePanel user={user} />
        </div>
      )}
    </main>
  );
}
