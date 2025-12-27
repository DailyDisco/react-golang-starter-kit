import { logger } from "../../lib/logger";
import { API_BASE_URL, apiFetch, authenticatedFetchWithParsing, parseErrorResponse } from "../api/client";
import type { AuthResponse, LoginRequest, RegisterRequest, User } from "../types";

// Token refresh configuration
const REFRESH_BUFFER_SECONDS = 60; // Refresh token 60 seconds before expiry
let tokenExpiresAt: number | null = null;
let refreshTimeout: ReturnType<typeof setTimeout> | null = null;

// Callback for updating auth state after token refresh
let onTokenRefresh: ((authData: AuthResponse) => void) | null = null;

export class AuthService {
  /**
   * Set callback for when tokens are refreshed
   */
  static setTokenRefreshCallback(callback: (authData: AuthResponse) => void): void {
    onTokenRefresh = callback;
  }

  /**
   * Validate that data can be safely stringified to JSON
   */
  private static validateJsonData(data: unknown): boolean {
    try {
      JSON.stringify(data);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Schedule automatic token refresh before expiry
   */
  private static scheduleTokenRefresh(expiresInSeconds: number): void {
    // Clear any existing timeout
    if (refreshTimeout) {
      clearTimeout(refreshTimeout);
      refreshTimeout = null;
    }

    // Calculate when to refresh (before expiry with buffer)
    const refreshInMs = (expiresInSeconds - REFRESH_BUFFER_SECONDS) * 1000;

    if (refreshInMs <= 0) {
      // Token is already expired or about to expire, refresh immediately
      this.refreshAccessToken().catch((error) => {
        logger.warn("Immediate token refresh failed", { error });
      });
      return;
    }

    tokenExpiresAt = Date.now() + expiresInSeconds * 1000;

    refreshTimeout = setTimeout(async () => {
      try {
        await this.refreshAccessToken();
      } catch (error) {
        logger.warn("Scheduled token refresh failed", { error });
        // Token refresh failed, user will be logged out on next API call
      }
    }, refreshInMs);

    logger.isDev() && logger.info(`Token refresh scheduled in ${Math.round(refreshInMs / 1000)}s`);
  }

  /**
   * Refresh access token using refresh token
   */
  static async refreshAccessToken(): Promise<AuthResponse> {
    const refreshToken = this.getRefreshToken();

    if (!refreshToken) {
      throw new Error("No refresh token available");
    }

    const response = await apiFetch(`${API_BASE_URL}/api/auth/refresh`, {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) {
      // Refresh token is invalid or expired, clear storage
      this.clearStorage();
      const apiError = await parseErrorResponse(response, "Token refresh failed");
      throw apiError;
    }

    try {
      const authData: AuthResponse = await response.json();

      // Store the new auth data
      this.storeAuthData(authData);

      // Schedule next refresh
      if (authData.expires_in) {
        this.scheduleTokenRefresh(authData.expires_in);
      }

      // Notify callback if set
      if (onTokenRefresh) {
        onTokenRefresh(authData);
      }

      return authData;
    } catch (parseError) {
      logger.error("Failed to parse refresh response", parseError);
      throw new Error("Invalid response format from server");
    }
  }

  /**
   * Authenticate user with email and password
   */
  static async login(credentials: LoginRequest): Promise<AuthResponse> {
    const response = await apiFetch(`${API_BASE_URL}/api/auth/login`, {
      method: "POST",
      body: JSON.stringify(credentials),
    });

    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Login failed");
      throw apiError;
    }

    try {
      const authData: AuthResponse = await response.json();

      // Schedule token refresh if expires_in is provided
      if (authData.expires_in) {
        this.scheduleTokenRefresh(authData.expires_in);
      }

      return authData;
    } catch (parseError) {
      logger.error("Failed to parse login response", parseError);
      throw new Error("Invalid response format from server");
    }
  }

  /**
   * Register a new user account
   */
  static async register(userData: RegisterRequest): Promise<AuthResponse> {
    const response = await apiFetch(`${API_BASE_URL}/api/auth/register`, {
      method: "POST",
      body: JSON.stringify(userData),
    });

    if (!response.ok) {
      const apiError = await parseErrorResponse(response, "Registration failed");
      throw apiError;
    }

    try {
      const authData: AuthResponse = await response.json();

      // Schedule token refresh if expires_in is provided
      if (authData.expires_in) {
        this.scheduleTokenRefresh(authData.expires_in);
      }

      return authData;
    } catch (parseError) {
      logger.error("Failed to parse registration response", parseError);
      throw new Error("Invalid response format from server");
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
  static async updateUser(userId: number, userData: Partial<User>): Promise<User> {
    return authenticatedFetchWithParsing<User>(`${API_BASE_URL}/api/users/${userId}`, {
      method: "PUT",
      body: JSON.stringify(userData),
    });
  }

  /**
   * Logout user by calling the logout API endpoint (clears httpOnly cookie)
   */
  static async logout(): Promise<void> {
    // Clear the refresh timeout
    if (refreshTimeout) {
      clearTimeout(refreshTimeout);
      refreshTimeout = null;
    }
    tokenExpiresAt = null;

    try {
      // Call logout endpoint to clear the httpOnly cookie on the server
      await apiFetch(`${API_BASE_URL}/api/auth/logout`, {
        method: "POST",
      });
    } catch (error) {
      // Log but don't throw - we still want to clear local state even if API call fails
      logger.warn("Logout API call failed", { error });
    }

    // Clear any remaining local storage data
    this.clearStorage();
  }

  /**
   * Check if user is authenticated by verifying session with the server
   */
  static async isAuthenticated(): Promise<boolean> {
    if (typeof window === "undefined") return false;

    try {
      await this.getCurrentUser();
      return true;
    } catch {
      // Session is invalid, clean up local state
      this.clearStorage();
      return false;
    }
  }

  /**
   * Get stored refresh token
   */
  static getRefreshToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("refresh_token");
  }

  /**
   * Get time until token expires in seconds (or null if unknown)
   */
  static getTimeUntilExpiry(): number | null {
    if (!tokenExpiresAt) return null;
    return Math.max(0, Math.round((tokenExpiresAt - Date.now()) / 1000));
  }

  /**
   * Store authentication data in localStorage
   * Only stores refresh token and minimal user data - access token is in httpOnly cookie
   */
  static storeAuthData(authData: AuthResponse): void {
    // Ensure we're on the client side
    if (typeof window === "undefined") {
      logger.warn("Cannot store auth data on server side");
      return;
    }

    // Validate user data before storing
    if (!authData.user) {
      logger.error("Invalid auth data provided", null, { authData });
      throw new Error("Invalid authentication data");
    }

    if (!this.validateJsonData(authData.user)) {
      logger.error("User data cannot be serialized to JSON", null, { user: authData.user });
      throw new Error("User data is not serializable");
    }

    try {
      // Store minimal user data for UI purposes (non-sensitive fields only)
      const minimalUser = {
        id: authData.user.id,
        name: authData.user.name,
        email: authData.user.email,
        role: authData.user.role,
      };
      localStorage.setItem("auth_user", JSON.stringify(minimalUser));

      // Store refresh token (used to get new access tokens via API)
      if (authData.refresh_token) {
        localStorage.setItem("refresh_token", authData.refresh_token);
      }
      // Note: Access token is handled via httpOnly cookie set by the backend
    } catch (storageError) {
      logger.error("Failed to store auth data in localStorage", storageError);
      throw new Error("Failed to store authentication data");
    }
  }

  /**
   * Clear all authentication data from localStorage
   */
  static clearStorage(): void {
    if (typeof window === "undefined") {
      return;
    }

    localStorage.removeItem("auth_user");
    localStorage.removeItem("refresh_token");
    // Note: Access token in httpOnly cookie is cleared by the backend logout endpoint
  }

  /**
   * Initialize token refresh from stored data
   * Call this on app startup if user is authenticated
   * Returns true if session is valid, false otherwise
   */
  static async initializeFromStorage(): Promise<boolean> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) {
      return false;
    }

    try {
      await this.refreshAccessToken();
      return true;
    } catch (error) {
      logger.warn("Failed to refresh token on initialization", { error });
      this.clearStorage();
      return false;
    }
  }

  /**
   * Validate current session by checking with the server
   * Returns true if session is valid, false otherwise
   */
  static async validateSession(): Promise<boolean> {
    try {
      await this.getCurrentUser();
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Start periodic session heartbeat
   * Validates session every intervalMs and calls onInvalid if session expires
   */
  private static heartbeatInterval: ReturnType<typeof setInterval> | null = null;

  static startSessionHeartbeat(
    intervalMs: number = 5 * 60 * 1000, // Default: 5 minutes
    onInvalid?: () => void
  ): void {
    // Clear any existing heartbeat
    this.stopSessionHeartbeat();

    this.heartbeatInterval = setInterval(async () => {
      const isValid = await this.validateSession();
      if (!isValid) {
        logger.warn("Session heartbeat detected invalid session");
        this.stopSessionHeartbeat();
        if (onInvalid) {
          onInvalid();
        } else {
          // Default behavior: dispatch session-expired event
          if (typeof window !== "undefined") {
            window.dispatchEvent(new CustomEvent("session-expired"));
          }
        }
      }
    }, intervalMs);

    logger.isDev() && logger.info(`Session heartbeat started (interval: ${intervalMs / 1000}s)`);
  }

  static stopSessionHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }
}
