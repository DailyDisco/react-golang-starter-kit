import type { LoginRequest, RegisterRequest, User } from "../../services";
import { AuthService } from "../../services/auth/authService";
import { useAuthStore } from "../../stores/auth-store";
import { useLogin, useRegister } from "../mutations/use-auth-mutations";
import { useUpdateUser as useUpdateUserMutation } from "../mutations/use-user-mutations";
import { useClearFeatureFlags } from "../queries/use-feature-flags";

interface AuthHookType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  register: (userData: RegisterRequest) => Promise<void>;
  logout: () => void;
  updateUser: (userData: Partial<User>) => Promise<void>;
  refreshUser: () => Promise<void>;
}

// Custom hook that provides auth functionality using Zustand store directly
// Authentication is handled via httpOnly cookies - no tokens stored in JS
export function useAuth(): AuthHookType {
  // Use Zustand store for state management
  const { user, isLoading, logout: storeLogout, setUser } = useAuthStore();

  // Always call hooks unconditionally (React hooks rule)
  // The mutations will handle SSR gracefully internally
  const loginMutation = useLogin();
  const registerMutation = useRegister();
  const updateUserMutation = useUpdateUserMutation();
  const clearFeatureFlags = useClearFeatureFlags();

  // Check if we're on client side for operations that need window
  const isClient = typeof window !== "undefined";

  const login = async (credentials: LoginRequest): Promise<void> => {
    if (!isClient) {
      throw new Error("Login not available during server-side rendering");
    }
    return new Promise((resolve, reject) => {
      loginMutation.mutate(credentials, {
        onSuccess: () => resolve(),
        onError: (error) => reject(error),
      });
    });
  };

  const register = async (userData: RegisterRequest): Promise<void> => {
    if (!isClient) {
      throw new Error("Register not available during server-side rendering");
    }
    return new Promise((resolve, reject) => {
      registerMutation.mutate(userData, {
        onSuccess: () => resolve(),
        onError: (error) => reject(error),
      });
    });
  };

  const logout = () => {
    // Clear feature flags cache before logging out
    clearFeatureFlags();
    // Use Zustand store logout method
    storeLogout();
  };

  const updateUser = async (userData: Partial<User>): Promise<void> => {
    if (!user) {
      throw new Error("Not authenticated");
    }
    if (!isClient) {
      throw new Error("Update not available during server-side rendering");
    }

    const updatedUser: User = { ...user, ...userData };
    return new Promise((resolve, reject) => {
      updateUserMutation.mutate(updatedUser, {
        onSuccess: () => resolve(),
        onError: (error) => reject(error),
      });
    });
  };

  const refreshUser = async (): Promise<void> => {
    if (!isClient) {
      throw new Error("User refresh not available during server-side rendering");
    }

    try {
      // Fetch the current user from the API
      const freshUser = await AuthService.getCurrentUser();
      // Update the Zustand store with the refreshed user data
      setUser(freshUser);
      // Also update localStorage for persistence
      localStorage.setItem("auth_user", JSON.stringify(freshUser));
    } catch (error) {
      // If refresh fails (e.g., session expired), log out
      storeLogout();
      throw error;
    }
  };

  return {
    user,
    isLoading,
    isAuthenticated: !!user,
    login,
    register,
    logout,
    updateUser,
    refreshUser,
  };
}
