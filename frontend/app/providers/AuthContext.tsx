import React, { createContext, useContext, useEffect } from 'react';
import { useAuthStore } from '../stores/auth-store';
import { useLogin, useRegister } from '../hooks/mutations/use-auth-mutations';
import { AuthService } from '../services';
import type { User, LoginRequest, RegisterRequest } from '../services';

interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  register: (userData: RegisterRequest) => Promise<void>;
  logout: () => void;
  updateUser: (userData: Partial<User>) => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: React.ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  // Use Zustand store for state management
  const { user, token, isLoading, setLoading, initialize } = useAuthStore();

  // Only use mutations on client side to avoid SSR issues
  const loginMutation = typeof window !== 'undefined' ? useLogin() : null;
  const registerMutation = typeof window !== 'undefined' ? useRegister() : null;

  // Initialize auth state on mount
  useEffect(() => {
    initialize();
  }, [initialize]);

  const login = async (credentials: LoginRequest): Promise<void> => {
    if (!loginMutation) {
      throw new Error('Login not available during server-side rendering');
    }
    return new Promise((resolve, reject) => {
      loginMutation.mutate(credentials, {
        onSuccess: () => resolve(),
        onError: error => reject(error),
      });
    });
  };

  const register = async (userData: RegisterRequest): Promise<void> => {
    if (!registerMutation) {
      throw new Error('Register not available during server-side rendering');
    }
    return new Promise((resolve, reject) => {
      registerMutation.mutate(userData, {
        onSuccess: () => resolve(),
        onError: error => reject(error),
      });
    });
  };

  const logout = () => {
    // Use Zustand store logout method
    const { logout: storeLogout } = useAuthStore.getState();
    storeLogout();
  };

  const updateUser = async (userData: Partial<User>): Promise<void> => {
    if (!user) {
      throw new Error('Not authenticated');
    }

    // This would need a mutation for updating user profile
    // For now, we'll throw an error to indicate it's not implemented
    throw new Error('User update not implemented in new system');
  };

  const refreshUser = async (): Promise<void> => {
    // This would need a query for refreshing current user
    // For now, we'll throw an error to indicate it's not implemented
    throw new Error('User refresh not implemented in new system');
  };

  const value: AuthContextType = {
    user,
    token,
    isLoading,
    isAuthenticated: !!user && !!token,
    login,
    register,
    logout,
    updateUser,
    refreshUser,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);

  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }

  return context;
}
