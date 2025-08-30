import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import type { User } from '../services';

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;

  // Actions
  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  setLoading: (loading: boolean) => void;
  logout: () => void;
  login: (user: User, token: string) => void;
  initialize: () => void;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      (set, get) => ({
        user: null,
        token: null,
        isLoading: true,
        isAuthenticated: false,

        setUser: user =>
          set({
            user,
            isAuthenticated: !!user && !!get().token,
          }),
        setToken: token =>
          set({
            token,
            isAuthenticated: !!token && !!get().user,
          }),
        setLoading: isLoading => set({ isLoading }),
        logout: () =>
          set({
            user: null,
            token: null,
            isAuthenticated: false,
          }),
        login: (user, token) =>
          set({
            user,
            token,
            isAuthenticated: true,
          }),
        initialize: () => {
          // This will be called when the app starts to load auth state from localStorage
          const storedToken =
            typeof window !== 'undefined'
              ? localStorage.getItem('auth_token')
              : null;
          const storedUser =
            typeof window !== 'undefined'
              ? localStorage.getItem('auth_user')
              : null;

          if (storedToken && storedUser) {
            try {
              const parsedUser = JSON.parse(storedUser);
              set({
                user: parsedUser,
                token: storedToken,
                isAuthenticated: true,
                isLoading: false,
              });
            } catch (error) {
              console.error('Auth state invalid:', error);
              get().logout();
              set({ isLoading: false });
            }
          } else {
            set({ isLoading: false });
          }
        },
      }),
      {
        name: 'auth-storage',
        partialize: state => ({
          user: state.user,
          token: state.token,
        }),
      }
    ),
    { name: 'auth-store' }
  )
);
