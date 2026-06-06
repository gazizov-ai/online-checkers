import { LogIn, UserPlus } from "lucide-react";
import { useState, type FormEvent } from "react";
import { useAuth } from "../auth/AuthContext";
import { friendlyApiError } from "../ui/labels";

type AuthMode = "login" | "register";

export function AuthPage() {
  const { login, register } = useAuth();
  const [mode, setMode] = useState<AuthMode>("login");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError(null);

    try {
      if (mode === "login") {
        await login({ username, password });
      } else {
        await register({ username, email, password });
      }
    } catch (err) {
      setError(friendlyApiError(err, "Не удалось выполнить вход."));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="auth-layout">
      <section className="auth-panel">
        <div className="brand-mark">
          <span className="brand-dot brand-dot--white" />
          <span className="brand-dot brand-dot--black" />
          <div>
            <h1>Online Checkers</h1>
            <p>Русские шашки</p>
          </div>
        </div>

        <div className="segmented-control" role="tablist">
          <button
            aria-selected={mode === "login"}
            className={mode === "login" ? "is-active" : ""}
            onClick={() => setMode("login")}
            type="button"
          >
            <LogIn aria-hidden size={17} />
            Вход
          </button>
          <button
            aria-selected={mode === "register"}
            className={mode === "register" ? "is-active" : ""}
            onClick={() => setMode("register")}
            type="button"
          >
            <UserPlus aria-hidden size={17} />
            Регистрация
          </button>
        </div>

        <form className="auth-form" onSubmit={(event) => void handleSubmit(event)}>
          <label>
            Имя пользователя
            <input
              autoComplete="username"
              minLength={1}
              onChange={(event) => setUsername(event.target.value)}
              required
              value={username}
            />
          </label>

          {mode === "register" ? (
            <label>
              Почта
              <input
                autoComplete="email"
                onChange={(event) => setEmail(event.target.value)}
                type="email"
                value={email}
              />
            </label>
          ) : null}

          <label>
            Пароль
            <input
              autoComplete={mode === "login" ? "current-password" : "new-password"}
              minLength={8}
              onChange={(event) => setPassword(event.target.value)}
              required
              type="password"
              value={password}
            />
          </label>

          {error ? <p className="form-error">{error}</p> : null}

          <button className="primary-button" disabled={submitting} type="submit">
            {mode === "login" ? <LogIn aria-hidden size={18} /> : <UserPlus aria-hidden size={18} />}
            {mode === "login" ? "Войти" : "Создать аккаунт"}
          </button>
        </form>
      </section>

      <section className="auth-board" aria-hidden>
        {Array.from({ length: 64 }).map((_, index) => {
          const row = Math.floor(index / 8);
          const col = index % 8;
          const dark = (row + col) % 2 === 1;
          const piece =
            dark && row <= 2 ? "black" : dark && row >= 5 ? "white" : null;

          return (
            <span className={dark ? "preview-square is-dark" : "preview-square"} key={index}>
              {piece ? <i className={`preview-piece preview-piece--${piece}`} /> : null}
            </span>
          );
        })}
      </section>
    </main>
  );
}
