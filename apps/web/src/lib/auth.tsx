import React from "react";

import * as api from "./api";

export type AuthState = {
  token: string | null;
  user: api.User | null;
  loading: boolean;
  error: string | null;
};

type AuthContextValue = AuthState & {
  login: (identifier: string, password: string) => Promise<void>;
  register: (email: string, username: string, password: string) => Promise<void>;
  logout: () => void;
};

const AuthContext = React.createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = React.useState<string | null>(
    () => localStorage.getItem("jabber_token")
  );
  const [user, setUser] = React.useState<api.User | null>(null);
  const [loading, setLoading] = React.useState<boolean>(!!token);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    let cancelled = false;

    async function load() {
      if (!token) {
        setLoading(false);
        return;
      }

      setLoading(true);
      try {
        const me = await api.fetchMe(token);
        if (!cancelled) {
          setUser(me);
          setError(null);
        }
      } catch (err) {
        if (!cancelled) {
          setToken(null);
          setUser(null);
          setError((err as Error).message);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, [token]);

  async function handleLogin(identifier: string, password: string) {
    setLoading(true);
    try {
      const result = await api.login(identifier, password);
      setToken(result.token);
      setUser(result.user);
      localStorage.setItem("jabber_token", result.token);
      setError(null);
    } catch (err) {
      setError((err as Error).message);
      throw err;
    } finally {
      setLoading(false);
    }
  }

  async function handleRegister(email: string, username: string, password: string) {
    setLoading(true);
    try {
      const result = await api.register(email, username, password);
      setToken(result.token);
      setUser(result.user);
      localStorage.setItem("jabber_token", result.token);
      setError(null);
    } catch (err) {
      setError((err as Error).message);
      throw err;
    } finally {
      setLoading(false);
    }
  }

  function handleLogout() {
    localStorage.removeItem("jabber_token");
    setToken(null);
    setUser(null);
  }

  const value: AuthContextValue = {
    token,
    user,
    loading,
    error,
    login: handleLogin,
    register: handleRegister,
    logout: handleLogout
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = React.useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}
