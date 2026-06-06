import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import * as authApi from "../api/auth";
import { clearStoredToken, getStoredToken } from "../api/http";
import type { User } from "../types/domain";

type AuthContextValue = {
  user: User | null;
  loading: boolean;
  login: (input: authApi.LoginInput) => Promise<void>;
  register: (input: authApi.RegisterInput) => Promise<void>;
  logout: () => void;
  refreshMe: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshMe = useCallback(async () => {
    if (!getStoredToken()) {
      setUser(null);
      setLoading(false);
      return;
    }

    try {
      const currentUser = await authApi.me();
      setUser(currentUser);
    } catch {
      clearStoredToken();
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refreshMe();
  }, [refreshMe]);

  const login = useCallback(async (input: authApi.LoginInput) => {
    const loggedInUser = await authApi.login(input);
    setUser(loggedInUser);
  }, []);

  const register = useCallback(async (input: authApi.RegisterInput) => {
    await authApi.register(input);
    const loggedInUser = await authApi.login({
      username: input.username,
      password: input.password,
    });
    setUser(loggedInUser);
  }, []);

  const logout = useCallback(() => {
    clearStoredToken();
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({
      user,
      loading,
      login,
      register,
      logout,
      refreshMe,
    }),
    [loading, login, logout, refreshMe, register, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const value = useContext(AuthContext);

  if (!value) {
    throw new Error("useAuth must be used inside AuthProvider");
  }

  return value;
}
