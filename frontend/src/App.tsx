import { LogOut } from "lucide-react";
import { useState } from "react";
import { useAuth } from "./auth/AuthContext";
import { LeaderboardPanel } from "./components/LeaderboardPanel";
import { ProfilePanel } from "./components/ProfilePanel";
import { AuthPage } from "./pages/AuthPage";
import { GamePage } from "./pages/GamePage";
import { LobbyPage } from "./pages/LobbyPage";
import { ProfilePage } from "./pages/ProfilePage";

export function App() {
  const { user, loading, logout } = useAuth();
  const [activeGameId, setActiveGameId] = useState<string | null>(null);
  const [activeProfileId, setActiveProfileId] = useState<string | null>(null);
  const [profileReturnGameId, setProfileReturnGameId] = useState<string | null>(null);
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
            setActiveProfileId(null);
            setProfileReturnGameId(null);
            setAutoStartSearch(false);
            logout();
          }}
          title="Выйти"
          type="button"
        >
          <LogOut aria-hidden size={18} />
        </button>
      </header>

      {activeProfileId ? (
        <div className="workspace">
          <ProfilePage
            currentUser={user}
            onBack={() => {
              setActiveProfileId(null);
              if (profileReturnGameId) {
                setActiveGameId(profileReturnGameId);
                setProfileReturnGameId(null);
              }
            }}
            onOpenGame={(gameId) => {
              setActiveProfileId(null);
              setProfileReturnGameId(null);
              setActiveGameId(gameId);
            }}
            onOpenProfile={setActiveProfileId}
            userId={activeProfileId}
          />
        </div>
      ) : activeGameId ? (
        <div className="workspace">
          <section className="primary-workspace">
            <GamePage
              gameId={activeGameId}
              onFindNewGame={() => {
                setAutoStartSearch(true);
                setActiveGameId(null);
              }}
              onLeave={() => setActiveGameId(null)}
              onOpenProfile={(userId) => {
                setProfileReturnGameId(activeGameId);
                setActiveGameId(null);
                setActiveProfileId(userId);
              }}
              user={user}
            />
          </section>
        </div>
      ) : (
        <div className="dashboard-stack">
          <div className="dashboard-main">
            <section className="primary-workspace">
              <LobbyPage
                autoStartSearch={autoStartSearch}
                onAutoStartConsumed={() => setAutoStartSearch(false)}
                onGameFound={(gameId) => {
                  setAutoStartSearch(false);
                  setActiveGameId(gameId);
                }}
              />
            </section>
            <ProfilePanel onOpenProfile={setActiveProfileId} user={user} />
          </div>
          <LeaderboardPanel onOpenProfile={setActiveProfileId} user={user} />
        </div>
      )}
    </main>
  );
}
