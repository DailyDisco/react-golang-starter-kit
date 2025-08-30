import {
  API_BASE_URL,
  apiFetch,
  authenticatedFetch,
  parseErrorResponse,
  authenticatedFetchWithParsing,
  apiFetchWithParsing,
  createHeaders,
} from '../api/client';
import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  User,
} from '../types';

export class AuthService {
  /**
   * Validate that data can be safely stringified to JSON
   */
  private static validateJsonData(data: any): boolean {
    try {
      JSON.stringify(data);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Authenticate user with email and password
   */
  static async login(credentials: LoginRequest): Promise<AuthResponse> {
    const response = await apiFetch(`${API_BASE_URL}/api/auth/login`, {
      method: 'POST',
      body: JSON.stringify(credentials),
    });

    if (!response.ok) {
      const errorMessage = await parseErrorResponse(response, 'Login failed');
      throw new Error(errorMessage);
    }

    try {
      return await response.json();
    } catch (parseError) {
      console.error('Failed to parse login response:', parseError);
      throw new Error('Invalid response format from server');
    }
  }

  /**
   * Register a new user account
   */
  static async register(userData: RegisterRequest): Promise<AuthResponse> {
    const response = await apiFetch(`${API_BASE_URL}/api/auth/register`, {
      method: 'POST',
      body: JSON.stringify(userData),
    });

    if (!response.ok) {
      const errorMessage = await parseErrorResponse(
        response,
        'Registration failed'
      );
      throw new Error(errorMessage);
    }

    try {
      return await response.json();
    } catch (parseError) {
      console.error('Failed to parse registration response:', parseError);
      throw new Error('Invalid response format from server');
    }
  }

  /**
   * Get current authenticated user information
   */
  static async getCurrentUser(): Promise<User> {
    return authenticatedFetchWithParsing<User>(`${API_BASE_URL}/api/auth/me`);
  }

  /**
   * Update user profile information
   */
  static async updateUser(
    userId: number,
    userData: Partial<User>
  ): Promise<User> {
    return authenticatedFetchWithParsing<User>(
      `${API_BASE_URL}/api/users/${userId}`,
      {
        method: 'PUT',
        body: JSON.stringify(userData),
      }
    );
  }

  /**
   * Logout user by clearing stored tokens
   */
  static logout(): void {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('auth_user');
    }
  }

  /**
   * Check if user is authenticated by verifying token exists and is valid
   */
  static async isAuthenticated(): Promise<boolean> {
    if (typeof window === 'undefined') return false;

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
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('auth_token');
  }

  /**
   * Store authentication data in localStorage
   */
  static storeAuthData(authData: AuthResponse): void {
    // Ensure we're on the client side
    if (typeof window === 'undefined') {
      console.warn('Cannot store auth data on server side');
      return;
    }

    // Validate data before storing
    if (!authData.token || !authData.user) {
      console.error('Invalid auth data provided:', authData);
      throw new Error('Invalid authentication data');
    }

    if (!this.validateJsonData(authData.user)) {
      console.error('User data cannot be serialized to JSON:', authData.user);
      throw new Error('User data is not serializable');
    }

    try {
      localStorage.setItem('auth_token', authData.token);
      localStorage.setItem('auth_user', JSON.stringify(authData.user));
    } catch (storageError) {
      console.error('Failed to store auth data in localStorage:', storageError);
      throw new Error('Failed to store authentication data');
    }
  }

  /**
   * Debug utility: Inspect current localStorage auth data
   */
  static debugInspectStorage(): void {
    if (typeof window === 'undefined') {
      console.log(
        '🔍 AuthService Debug: Running on server side, no localStorage available'
      );
      return;
    }

    const token = localStorage.getItem('auth_token');
    const user = localStorage.getItem('auth_user');

    console.group('🔍 AuthService Debug - localStorage Inspection');
    console.log('auth_token:', token ? `${token.substring(0, 20)}...` : 'null');
    console.log('auth_user raw:', user);

    if (user) {
      try {
        const parsed = JSON.parse(user);
        console.log('auth_user parsed:', parsed);
        console.log('Is valid JSON: ✅ Yes');
      } catch (parseError) {
        console.error('Is valid JSON: ❌ No - Parse error:', parseError);
        console.log('Raw user data inspection:');
        console.log('Length:', user.length);
        console.log('First 50 chars:', user.substring(0, 50));
        console.log(
          'Position 4 char:',
          user.charAt(4),
          `(code: ${user.charCodeAt(4)})`
        );
      }
    } else {
      console.log('auth_user: null');
    }
    console.groupEnd();
  }

  /**
   * Clear all authentication data from localStorage
   */
  static clearStorage(): void {
    if (typeof window === 'undefined') {
      console.log('🧹 AuthService: Cannot clear storage on server side');
      return;
    }

    console.log('🧹 Clearing authentication data from localStorage');
    localStorage.removeItem('auth_token');
    localStorage.removeItem('auth_user');
  }

  /**
   * Debug utility: Test API connectivity and inspect raw responses
   */
  static async debugApiResponse(
    endpoint: string,
    method: string = 'GET',
    body?: any
  ): Promise<void> {
    try {
      console.group(`🔍 AuthService Debug - ${method} ${endpoint}`);

      const headers = createHeaders(method !== 'GET');
      const options: RequestInit = {
        method,
        headers,
      };

      if (body && method !== 'GET') {
        options.body = JSON.stringify(body);
      }

      console.log('Request options:', options);

      const response = await fetch(`${API_BASE_URL}/api${endpoint}`, options);
      console.log('Response status:', response.status);
      console.log(
        'Response headers:',
        Object.fromEntries(response.headers.entries())
      );

      const responseText = await response.text();
      console.log('Raw response text:', responseText);
      console.log('Response text length:', responseText.length);

      // Try to parse as JSON
      try {
        const jsonData = JSON.parse(responseText);
        console.log('Parsed JSON:', jsonData);
        console.log('✅ Response is valid JSON');
      } catch (jsonError) {
        console.log('❌ Response is not valid JSON:', jsonError);
        console.log('First 100 chars:', responseText.substring(0, 100));
        console.log(
          'Last 100 chars:',
          responseText.substring(Math.max(0, responseText.length - 100))
        );
      }

      console.groupEnd();
    } catch (error) {
      console.error('❌ API Debug failed:', error);
      console.groupEnd();
    }
  }
}
