import { API_BASE_URL, apiFetch, authenticatedFetch } from '../api/client';
import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  User,
} from '../types';

export class AuthService {
  /**
   * Authenticate user with email and password
   */
  static async login(credentials: LoginRequest): Promise<AuthResponse> {
    const response = await apiFetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      body: JSON.stringify(credentials),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Login failed');
    }

    return response.json();
  }

  /**
   * Register a new user account
   */
  static async register(userData: RegisterRequest): Promise<AuthResponse> {
    const response = await apiFetch(`${API_BASE_URL}/auth/register`, {
      method: 'POST',
      body: JSON.stringify(userData),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Registration failed');
    }

    return response.json();
  }

  /**
   * Get current authenticated user information
   */
  static async getCurrentUser(): Promise<User> {
    const response = await authenticatedFetch(`${API_BASE_URL}/auth/me`);
    if (!response.ok) {
      throw new Error('Failed to fetch current user');
    }
    return response.json();
  }

  /**
   * Update user profile information
   */
  static async updateUser(
    userId: number,
    userData: Partial<User>
  ): Promise<User> {
    const response = await authenticatedFetch(
      `${API_BASE_URL}/users/${userId}`,
      {
        method: 'PUT',
        body: JSON.stringify(userData),
      }
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Update failed');
    }

    return response.json();
  }

  /**
   * Logout user by clearing stored tokens
   */
  static logout(): void {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('auth_user');
  }

  /**
   * Check if user is authenticated by verifying token exists and is valid
   */
  static async isAuthenticated(): Promise<boolean> {
    const token = localStorage.getItem('auth_token');
    if (!token) return false;

    try {
      await this.getCurrentUser();
      return true;
    } catch {
      // Token is invalid, clean up
      this.logout();
      return false;
    }
  }

  /**
   * Get stored authentication token
   */
  static getToken(): string | null {
    return localStorage.getItem('auth_token');
  }

  /**
   * Store authentication data in localStorage
   */
  static storeAuthData(authData: AuthResponse): void {
    localStorage.setItem('auth_token', authData.token);
    localStorage.setItem('auth_user', JSON.stringify(authData.user));
  }
}
