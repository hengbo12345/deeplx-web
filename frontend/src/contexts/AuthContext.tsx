import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { api } from '../lib/api';

const TOKEN_KEY = 'deeplx_token';

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  token: string | null;
  login: (token: string) => Promise<boolean>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [token, setToken] = useState<string | null>(null);

  useEffect(() => {
    const stored = localStorage.getItem(TOKEN_KEY);
    if (!stored) {
      setIsLoading(false);
      return;
    }
    api.verifyToken(stored).then((res) => {
      if (res.success && res.data?.valid) {
        setToken(stored);
        setIsAuthenticated(true);
      } else {
        localStorage.removeItem(TOKEN_KEY);
      }
      setIsLoading(false);
    }).catch(() => {
      localStorage.removeItem(TOKEN_KEY);
      setIsLoading(false);
    });
  }, []);

  const login = useCallback(async (newToken: string): Promise<boolean> => {
    try {
      const res = await api.verifyToken(newToken);
      if (res.success && res.data?.valid) {
        localStorage.setItem(TOKEN_KEY, newToken);
        setToken(newToken);
        setIsAuthenticated(true);
        return true;
      }
      return false;
    } catch {
      return false;
    }
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    setToken(null);
    setIsAuthenticated(false);
  }, []);

  return (
    <AuthContext.Provider value={{ isAuthenticated, isLoading, token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
