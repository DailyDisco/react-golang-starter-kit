import { create } from "zustand";
import { devtools, persist } from "zustand/middleware";

import { broadcastLogout, broadcastSessionExpired, subscribeToAuthEvents } from "../lib/auth-channel";
import { logger } from "../lib/logger";
import type { User } from "../services";
import { AuthService } from "../services/auth/authService";

interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isInitialized: boolean;

  // Actions
  setUser: (user: User | null) => void;
  setLoading: (loading: boolean) => void;
  logout: (broadcast?: boolean) => void;
  login: (user: User) => void;
  initialize: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      (set, get) => ({
        user: null,
        isLoading: true,
        isAuthenticated: false,
        isInitialized: false,

        setUser: (user) =>
          set({
            user,
            isAuthenticated: !!user,
          }),
        setLoading: (isLoading) => set({ isLoading }),
        logout: (broadcast = true) => {
          // Stop session heartbeat
          AuthService.stopSessionHeartbeat();
          // Clear auth state
          set({
            user: null,
            isAuthenticated: false,
          });
          // Clear localStorage
          AuthService.clearStorage();
          // Broadcast logout to other tabs (unless this was triggered by another tab)
          if (broadcast) {
            broadcastLogout();
          }
        },
        login: (user) =>
          set({
            user,
            isAuthenticated: true,
          }),
        initialize: async () => {
          // Prevent double initialization
          if (get().isInitialized) {
            return;
          }

          // Subscribe to auth events from other tabs (logout, session expired)
          subscribeToAuthEvents((event) => {
            if (event.type === "logout" || event.type === "session_expired") {
              // Another tab logged out or session expired - sync this tab
              logger.info("Auth event from another tab", { type: event.type });
              get().logout(false); // Don't broadcast back
              if (typeof window !== "undefined") {
                window.dispatchEvent(new CustomEvent("session-expired"));
              }
            }
          });

          // Load cached user data from localStorage for faster UI rendering
          const storedUser = typeof window !== "undefined" ? localStorage.getItem("auth_user") : null;

          if (!storedUser) {
            set({ isLoading: false, isInitialized: true });
            return;
          }

          try {
            const parsedUser = JSON.parse(storedUser);

            // Show cached user data immediately for better UX
            set({
              user: parsedUser,
              isAuthenticated: true,
            });

            // Set up the token refresh callback to update the store
            AuthService.setTokenRefreshCallback((authData) => {
              set({
                user: authData.user,
                isAuthenticated: true,
              });
            });

            // Actually validate the session with the server
            const isValid = await AuthService.initializeFromStorage();

            if (isValid) {
              // Session is valid - start heartbeat to detect future expiration
              AuthService.startSessionHeartbeat(5 * 60 * 1000, () => {
                // Session expired - logout and notify other tabs
                broadcastSessionExpired();
                get().logout(false); // Already broadcast via broadcastSessionExpired
                if (typeof window !== "undefined") {
                  window.dispatchEvent(new CustomEvent("session-expired"));
                }
              });

              set({ isLoading: false, isInitialized: true });
            } else {
              // Session is invalid - clear auth state
              logger.warn("Session validation failed during initialization");
              get().logout();
              set({ isLoading: false, isInitialized: true });
            }
          } catch (error) {
            logger.error("Auth state invalid", error);
            get().logout();
            set({ isLoading: false, isInitialized: true });
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
