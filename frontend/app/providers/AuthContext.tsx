import React, { createContext, useContext, useEffect, useState } from 'react';
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
    const [user, setUser] = useState<User | null>(null);
    const [token, setToken] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    // Load user and token from localStorage on mount
    useEffect(() => {
        const loadAuthState = async () => {
            try {
                const storedToken = localStorage.getItem('auth_token');
                const storedUser = localStorage.getItem('auth_user');

                if (storedToken && storedUser) {
                    const parsedUser = JSON.parse(storedUser);
                    setToken(storedToken);
                    setUser(parsedUser);

                    // Verify token is still valid by fetching current user
                    try {
                        await refreshUser();
                    } catch (error) {
                        // Token is invalid, clear auth state
                        logout();
                    }
                }
            } catch (error) {
                console.error('Failed to load auth state:', error);
                logout();
            } finally {
                setIsLoading(false);
            }
        };

        loadAuthState();
    }, []);

    const login = async (credentials: LoginRequest): Promise<void> => {
        setIsLoading(true);
        try {
            const authData = await AuthService.login(credentials);

            setUser(authData.user);
            setToken(authData.token);

            // Store in localStorage
            AuthService.storeAuthData(authData);
        } catch (error) {
            throw error;
        } finally {
            setIsLoading(false);
        }
    };

    const register = async (userData: RegisterRequest): Promise<void> => {
        setIsLoading(true);
        try {
            const authData = await AuthService.register(userData);

            setUser(authData.user);
            setToken(authData.token);

            // Store in localStorage
            AuthService.storeAuthData(authData);
        } catch (error) {
            throw error;
        } finally {
            setIsLoading(false);
        }
    };

    const logout = () => {
        setUser(null);
        setToken(null);
        AuthService.logout();
    };

    const updateUser = async (userData: Partial<User>): Promise<void> => {
        if (!user) {
            throw new Error('Not authenticated');
        }

        try {
            const updatedUser = await AuthService.updateUser(user.id, userData);
            setUser(updatedUser);
            localStorage.setItem('auth_user', JSON.stringify(updatedUser));
        } catch (error) {
            throw error;
        }
    };

    const refreshUser = async (): Promise<void> => {
        try {
            const currentUser = await AuthService.getCurrentUser();
            setUser(currentUser);
            localStorage.setItem('auth_user', JSON.stringify(currentUser));
        } catch (error) {
            throw error;
        }
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

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth(): AuthContextType {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
}
