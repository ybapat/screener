"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  type ReactNode,
} from "react";
import {
  login as apiLogin,
  register as apiRegister,
  getProfile,
  type User,
  type UserRole,
} from "@/lib/api";
import { setTokens, clearTokens, isAuthenticated } from "@/lib/auth";

interface AuthState {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (params: {
    email: string;
    password: string;
    display_name: string;
    role: UserRole;
    age_range?: string;
    country?: string;
  }) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthState | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshUser = useCallback(async () => {
    try {
      const profile = await getProfile();
      setUser(profile);
    } catch {
      setUser(null);
      clearTokens();
    }
  }, []);

  useEffect(() => {
    if (isAuthenticated()) {
      refreshUser().finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, [refreshUser]);

  const login = useCallback(
    async (email: string, password: string) => {
      const result = await apiLogin(email, password);
      setTokens(result.tokens.access_token, result.tokens.refresh_token);
      setUser(result.user);
    },
    []
  );

  const register = useCallback(
    async (params: {
      email: string;
      password: string;
      display_name: string;
      role: UserRole;
      age_range?: string;
      country?: string;
    }) => {
      const result = await apiRegister(params);
      setTokens(result.tokens.access_token, result.tokens.refresh_token);
      setUser(result.user);
    },
    []
  );

  const logout = useCallback(() => {
    clearTokens();
    setUser(null);
    window.location.href = "/";
  }, []);

  return (
    <AuthContext.Provider
      value={{ user, loading, login, register, logout, refreshUser }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthState {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
