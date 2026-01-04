import React from "react";

import * as api from "./api";
import { useLogin, useMe, useRegister } from "./hooks/auth";

export type AuthState = {
  token: string | null;
  user: api.User | null;
  loading: boolean;
  error: string | null;
  authError: string | null;
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
  const [authError, setAuthError] = React.useState<string | null>(null);
  const meQuery = useMe(token);
  const loginMutation = useLogin();
  const registerMutation = useRegister();

  React.useEffect(() => {
    if (meQuery.data) {
      setUser(meQuery.data);
    } else if (!token) {
      setUser(null);
    }
  }, [meQuery.data, token]);

  React.useEffect(() => {
    if (meQuery.isError && token) {
      localStorage.removeItem("jabber_token");
      setToken(null);
      setUser(null);
    }
  }, [meQuery.isError, token]);

  async function handleLogin(identifier: string, password: string) {
    try {
      const result = await loginMutation.mutateAsync({ identifier, password });
      setToken(result.token);
      setUser(result.user);
      localStorage.setItem("jabber_token", result.token);
      setAuthError(null);
    } catch (err) {
      setAuthError((err as Error).message);
      throw err;
    }
  }

  async function handleRegister(email: string, username: string, password: string) {
    try {
      const result = await registerMutation.mutateAsync({ email, username, password });
      setToken(result.token);
      setUser(result.user);
      localStorage.setItem("jabber_token", result.token);
      setAuthError(null);
    } catch (err) {
      setAuthError((err as Error).message);
      throw err;
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
    loading: meQuery.isLoading || loginMutation.isPending || registerMutation.isPending,
    error: (meQuery.error as Error | null)?.message ?? null,
    authError,
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
