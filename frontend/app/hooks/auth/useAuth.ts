import type { LoginRequest, RegisterRequest, User } from '../../services';
import { useAuthStore } from '../../stores/auth-store';
import { useLogin, useRegister } from '../mutations/use-auth-mutations';

interface AuthHookType {
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

// Custom hook that provides auth functionality using Zustand store directly
export function useAuth(): AuthHookType {
  // Use Zustand store for state management
  const {
    user,
    token,
    isLoading,
    setLoading,
    logout: storeLogout,
  } = useAuthStore();

  // Only use mutations on client side to avoid SSR issues
  const loginMutation = typeof window !== 'undefined' ? useLogin() : null;
  const registerMutation = typeof window !== 'undefined' ? useRegister() : null;

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

  return {
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
}
