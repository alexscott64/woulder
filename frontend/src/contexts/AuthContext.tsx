/* eslint-disable react-refresh/only-export-components */
import { createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { authApi, clearAuthTokens, getStoredRefreshToken } from '../services/auth';
import { MoneyCurrentUser } from '../types/money';

interface AuthContextValue {
  user: MoneyCurrentUser | null;
  isAuthenticated: boolean;
  isBootstrapping: boolean;
  isLoginSubmitting: boolean;
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
  canWrite: boolean;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<MoneyCurrentUser | null>(null);
  const [isBootstrapping, setIsBootstrapping] = useState(true);
  const [isLoginSubmitting, setIsLoginSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refreshSession = useCallback(async () => {
    if (!getStoredRefreshToken()) {
      setUser(null);
      setIsBootstrapping(false);
      return;
    }

    setIsBootstrapping(true);
    setError(null);
    try {
      console.debug('[money-auth] refreshing persisted session');
      const refreshed = await authApi.refresh();
      setUser(refreshed.user);
    } catch {
      clearAuthTokens();
      setUser(null);
      setError('Session expired. Sign in again.');
    } finally {
      setIsBootstrapping(false);
    }
  }, []);

  useEffect(() => {
    // Auth bootstrap intentionally synchronizes persisted refresh-token state into React state.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void refreshSession();
  }, [refreshSession]);

  const login = useCallback(async (email: string, password: string) => {
    setIsLoginSubmitting(true);
    setError(null);
    try {
      console.debug('[money-auth] submitting login request');
      const response = await authApi.login(email, password);
      setUser(response.user);
    } catch {
      setError('Invalid email or password.');
      setUser(null);
      throw new Error('Invalid email or password.');
    } finally {
      setIsLoginSubmitting(false);
    }
  }, []);

  const logout = useCallback(async () => {
    setIsLoginSubmitting(true);
    try {
      await authApi.logout();
      setUser(null);
    } finally {
      setIsLoginSubmitting(false);
    }
  }, []);

  const value = useMemo<AuthContextValue>(() => ({
    user,
    isAuthenticated: Boolean(user),
    isBootstrapping,
    isLoginSubmitting,
    error,
    login,
    logout,
    refreshSession,
    canWrite: user?.role === 'admin' || user?.role === 'developer',
  }), [error, isBootstrapping, isLoginSubmitting, login, logout, refreshSession, user]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used inside AuthProvider');
  }
  return context;
}
