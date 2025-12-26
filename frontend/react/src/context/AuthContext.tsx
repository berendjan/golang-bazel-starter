import { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { Session, Identity } from '@ory/client';
import { kratos } from '../lib/kratos';

interface AuthContextType {
  session: Session | null;
  identity: Identity | null;
  loading: boolean;
  error: string | null;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<Session | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refreshSession = async () => {
    try {
      setLoading(true);
      setError(null);
      const { data } = await kratos.toSession();
      setSession(data);
    } catch (err) {
      setSession(null);
      // 401 is expected when not logged in
      if ((err as { response?: { status: number } })?.response?.status !== 401) {
        setError('Failed to check session');
      }
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    try {
      // Create logout flow
      const { data: logoutFlow } = await kratos.createBrowserLogoutFlow();
      // Perform logout
      await kratos.updateLogoutFlow({ token: logoutFlow.logout_token });
      setSession(null);
      // Redirect to home after logout
      window.location.href = '/';
    } catch (err) {
      console.error('Logout failed:', err);
      setError('Failed to logout');
    }
  };

  useEffect(() => {
    refreshSession();
  }, []);

  const value: AuthContextType = {
    session,
    identity: session?.identity ?? null,
    loading,
    error,
    logout,
    refreshSession,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
