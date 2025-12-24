import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";

import { logger } from "../lib/logger";
import type { User } from "../services";
import { AuthService } from "../services/auth/authService";

interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;

  // Actions
  setUser: (user: User | null) => void;
  setLoading: (loading: boolean) => void;
  logout: () => void;
  login: (user: User) => void;
  initialize: () => void;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      (set, get) => ({
        user: null,
        isLoading: true,
        isAuthenticated: false,

        setUser: (user) =>
          set({
            user,
            isAuthenticated: !!user,
          }),
        setLoading: (isLoading) => set({ isLoading }),
        logout: () => {
          // Clear auth state
          set({
            user: null,
            isAuthenticated: false,
          });
          // Clear localStorage
          AuthService.clearStorage();
        },
        login: (user) =>
          set({
            user,
            isAuthenticated: true,
          }),
        initialize: () => {
          // Load cached user data from localStorage for faster UI rendering
          // Actual authentication is validated via httpOnly cookie when making API calls
          const storedUser = typeof window !== "undefined" ? localStorage.getItem("auth_user") : null;

          if (storedUser) {
            try {
              const parsedUser = JSON.parse(storedUser);
              set({
                user: parsedUser,
                isAuthenticated: true,
                isLoading: false,
              });

              // Set up the token refresh callback to update the store
              AuthService.setTokenRefreshCallback((authData) => {
                set({
                  user: authData.user,
                  isAuthenticated: true,
                });
              });

              // Initialize token refresh from stored refresh token
              AuthService.initializeFromStorage();
            } catch (error) {
              logger.error("Auth state invalid", error);
              get().logout();
              set({ isLoading: false });
            }
          } else {
            set({ isLoading: false });
          }
        },
      }),
      {
        name: "auth-storage",
        partialize: (state) => ({
          // Only persist minimal user data for UI purposes
          user: state.user
            ? {
                id: state.user.id,
                name: state.user.name,
                email: state.user.email,
                role: state.user.role,
              }
            : null,
        }),
      }
    ),
    { name: "auth-store" }
  )
);
