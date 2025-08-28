import React, { createContext, useContext, useEffect, useState } from 'react';

export interface User {
    id: number;
    name: string;
    email: string;
    email_verified: boolean;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface AuthResponse {
    user: User;
    token: string;
}

export interface LoginRequest {
    email: string;
    password: string;
}

export interface RegisterRequest {
    name: string;
    email: string;
    password: string;
}

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
            const response = await fetch(`${import.meta.env.VITE_API_URL}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(credentials),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Login failed');
            }

            const authData: AuthResponse = await response.json();

            setUser(authData.user);
            setToken(authData.token);

            // Store in localStorage
            localStorage.setItem('auth_token', authData.token);
            localStorage.setItem('auth_user', JSON.stringify(authData.user));
        } catch (error) {
            throw error;
        } finally {
            setIsLoading(false);
        }
    };

    const register = async (userData: RegisterRequest): Promise<void> => {
        setIsLoading(true);
        try {
            const response = await fetch(`${import.meta.env.VITE_API_URL}/auth/register`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(userData),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Registration failed');
            }

            const authData: AuthResponse = await response.json();

            setUser(authData.user);
            setToken(authData.token);

            // Store in localStorage
            localStorage.setItem('auth_token', authData.token);
            localStorage.setItem('auth_user', JSON.stringify(authData.user));
        } catch (error) {
            throw error;
        } finally {
            setIsLoading(false);
        }
    };

    const logout = () => {
        setUser(null);
        setToken(null);
        localStorage.removeItem('auth_token');
        localStorage.removeItem('auth_user');
    };

    const updateUser = async (userData: Partial<User>): Promise<void> => {
        if (!user || !token) {
            throw new Error('Not authenticated');
        }

        try {
            const response = await fetch(`${import.meta.env.VITE_API_URL}/users/${user.id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`,
                },
                body: JSON.stringify(userData),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Update failed');
            }

            const updatedUser: User = await response.json();
            setUser(updatedUser);
            localStorage.setItem('auth_user', JSON.stringify(updatedUser));
        } catch (error) {
            throw error;
        }
    };

    const refreshUser = async (): Promise<void> => {
        if (!token) {
            throw new Error('No token available');
        }

        try {
            const response = await fetch(`${import.meta.env.VITE_API_URL}/auth/me`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                },
            });

            if (!response.ok) {
                throw new Error('Failed to refresh user data');
            }

            const currentUser: User = await response.json();
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
