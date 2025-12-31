import { logger } from "../../lib/logger";

/**
 * Custom error class that preserves API error codes for contextual handling
 */
export class ApiError extends Error {
  code: string;
  statusCode: number;

  constructor(message: string, code: string, statusCode: number) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.statusCode = statusCode;
  }
}

// Create API_BASE_URL safely for SSR and remote development
const getApiBaseUrl = () => {
  // In SSR, use absolute URL for server-side API calls
  if (typeof window === "undefined") {
    return "http://localhost:8080";
  }
  // In browser, use empty string for relative URLs (goes through Vite proxy)
  // This enables remote development where localhost:8080 isn't accessible
  // Set VITE_API_URL to an absolute URL only if you need to bypass the proxy
  return import.meta.env.VITE_API_URL || "";
};

export const API_BASE_URL = getApiBaseUrl();

// Generate a unique request ID for traceability
export const generateRequestId = (): string => {
  // Use crypto.randomUUID() if available (modern browsers)
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  // Fallback for older environments
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
};

// Get CSRF token from cookie
const getCSRFToken = (): string | null => {
  if (typeof document === "undefined") return null;
  const match = document.cookie.match(new RegExp("(^| )csrf_token=([^;]+)"));
  return match ? decodeURIComponent(match[2]) : null;
};

/**
 * Fetch a fresh CSRF token from the server.
 * Call this on app initialization to ensure CSRF protection is ready.
 */
export const initCSRFToken = async (): Promise<string | null> => {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/csrf-token`, {
      credentials: "include",
    });
    if (response.ok) {
      const data = (await response.json()) as { csrf_token: string };
      return data.csrf_token;
    }
  } catch {
    // Silently fail - CSRF token will be in cookie from response
  }
  return getCSRFToken();
};

// Create headers with request ID and optional CSRF token
// Authentication is handled via httpOnly cookies (credentials: "include")
export const createHeaders = (includeCSRF: boolean = false): Record<string, string> => {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "X-Request-ID": generateRequestId(),
  };

  // Include CSRF token for state-changing requests
  if (includeCSRF) {
    const csrfToken = getCSRFToken();
    if (csrfToken) {
      headers["X-CSRF-Token"] = csrfToken;
    }
  }

  return headers;
};

// Check if method is state-changing (needs CSRF protection)
const isStateChangingMethod = (method?: string): boolean => {
  const stateChangingMethods = ["POST", "PUT", "DELETE", "PATCH"];
  return stateChangingMethods.includes((method || "GET").toUpperCase());
};

// Track if we're currently refreshing to prevent concurrent refresh attempts
let isRefreshing = false;
let refreshPromise: Promise<boolean> | null = null;

// Grace period after login/registration to allow cookies to propagate
// During this window, we don't fire session-expired events on 401s
let authGraceUntil: number = 0;
const AUTH_GRACE_PERIOD_MS = 5000; // 5 seconds

// Also track in sessionStorage as backup (survives HMR)
const AUTH_GRACE_KEY = "auth_grace_until";

/**
 * Mark that authentication just happened (login/register).
 * During the grace period, 401 errors won't trigger session-expired events.
 * This prevents false positives while httpOnly cookies propagate in the browser.
 */
export const markAuthenticationComplete = (): void => {
  authGraceUntil = Date.now() + AUTH_GRACE_PERIOD_MS;
  // Also store in sessionStorage as backup
  if (typeof window !== "undefined") {
    sessionStorage.setItem(AUTH_GRACE_KEY, authGraceUntil.toString());
  }
};

/**
 * Check if we're within the post-authentication grace period
 */
const isInAuthGracePeriod = (): boolean => {
  // Check in-memory value first
  if (Date.now() < authGraceUntil) {
    return true;
  }
  // Fall back to sessionStorage (handles HMR/module reload)
  if (typeof window !== "undefined") {
    const stored = sessionStorage.getItem(AUTH_GRACE_KEY);
    if (stored) {
      const storedTime = parseInt(stored, 10);
      if (Date.now() < storedTime) {
        // Restore in-memory value
        authGraceUntil = storedTime;
        return true;
      } else {
        // Clean up expired value
        sessionStorage.removeItem(AUTH_GRACE_KEY);
      }
    }
  }
  return false;
};

// Import AuthService dynamically to avoid circular dependency
const getAuthService = async () => {
  const { AuthService } = await import("../auth/authService");
  return AuthService;
};

// Enhanced fetch with auth (uses httpOnly cookies via credentials: "include")
// Includes automatic retry on CSRF token errors and 401 token refresh
export const authenticatedFetch = async (url: string, options: RequestInit = {}): Promise<Response> => {
  const needsCSRF = isStateChangingMethod(options.method);
  let headers = createHeaders(needsCSRF);

  const response = await fetch(url, {
    ...options,
    credentials: "include", // Include httpOnly cookies for authentication
    headers: {
      ...headers,
      ...options.headers,
    },
  });

  // Retry once if CSRF token was missing or invalid (auto-refresh)
  if (response.status === 403 && needsCSRF) {
    try {
      const clonedResponse = response.clone();
      const errorData = (await clonedResponse.json()) as { error?: string };
      if (errorData.error === "CSRF_ERROR") {
        // Refresh CSRF token and retry
        await initCSRFToken();
        headers = createHeaders(true);
        return fetch(url, {
          ...options,
          credentials: "include",
          headers: {
            ...headers,
            ...options.headers,
          },
        });
      }
    } catch {
      // If we can't parse the error, just return the original response
    }
  }

  // Handle 401 Unauthorized - try to refresh token and retry once
  if (response.status === 401) {
    // Don't try to refresh on login/refresh endpoints to avoid infinite loops
    if (url.includes("/auth/login") || url.includes("/auth/refresh") || url.includes("/auth/register")) {
      return response;
    }

    try {
      // Use a single refresh promise to prevent concurrent refresh attempts
      if (!isRefreshing) {
        isRefreshing = true;
        const AuthService = await getAuthService();
        refreshPromise = AuthService.initializeFromStorage();
      }

      const refreshed = await refreshPromise;
      isRefreshing = false;
      refreshPromise = null;

      if (refreshed) {
        // Token refreshed successfully, retry the original request
        headers = createHeaders(needsCSRF);
        return fetch(url, {
          ...options,
          credentials: "include",
          headers: {
            ...headers,
            ...options.headers,
          },
        });
      }
    } catch (refreshError) {
      isRefreshing = false;
      refreshPromise = null;
      logger.warn("Token refresh failed during 401 recovery", { error: refreshError });
    }

    // Refresh failed or wasn't possible - dispatch session-expired event
    // But not during the grace period right after login (cookies may still be propagating)
    if (typeof window !== "undefined" && window.location.pathname !== "/login" && !isInAuthGracePeriod()) {
      window.dispatchEvent(new CustomEvent("session-expired"));
    }
  }

  return response;
};

// Simple fetch for auth endpoints (login/register) - includes credentials for cookie handling
// Includes automatic retry on CSRF token errors
export const apiFetch = async (url: string, options: RequestInit = {}): Promise<Response> => {
  const needsCSRF = isStateChangingMethod(options.method);
  let headers = createHeaders(needsCSRF);

  const response = await fetch(url, {
    ...options,
    credentials: "include", // Include httpOnly cookies
    headers: {
      ...headers,
      ...options.headers,
    },
  });

  // Retry once if CSRF token was missing or invalid (auto-refresh)
  if (response.status === 403 && needsCSRF) {
    try {
      const clonedResponse = response.clone();
      const errorData = (await clonedResponse.json()) as { error?: string };
      if (errorData.error === "CSRF_ERROR") {
        // Refresh CSRF token and retry
        await initCSRFToken();
        headers = createHeaders(true);
        return fetch(url, {
          ...options,
          credentials: "include",
          headers: {
            ...headers,
            ...options.headers,
          },
        });
      }
    } catch {
      // If we can't parse the error, just return the original response
    }
  }

  return response;
};

// Types for API responses
export interface ApiSuccessResponse<T = any> {
  success: true;
  message: string;
  data?: T;
}

export interface ApiErrorResponse {
  error: string;
  message: string;
  code: number;
  request_id?: string;
}

export type ApiResponse<T = any> = ApiSuccessResponse<T> | ApiErrorResponse;

/**
 * Safely parse error response from server and return ApiError
 */
export const parseErrorResponse = async (response: Response, defaultMessage: string): Promise<ApiError> => {
  try {
    // Get response as text first, then try to parse as JSON
    const responseText = await response.text();

    // Try to parse as JSON if it looks like JSON
    if (responseText.trim().startsWith("{") || responseText.trim().startsWith("[")) {
      const errorData = JSON.parse(responseText) as ApiErrorResponse;
      const message = errorData.message || errorData.error || defaultMessage;
      const code = errorData.error || "UNKNOWN_ERROR";
      return new ApiError(message, code, response.status);
    } else {
      // Not JSON, use the text directly
      const message = responseText || `${defaultMessage} with status ${response.status}`;
      return new ApiError(message, "UNKNOWN_ERROR", response.status);
    }
  } catch (parseError) {
    // If anything fails, use a generic error message
    logger.error("Failed to parse error response", parseError);
    return new ApiError(`${defaultMessage} with status ${response.status}`, "UNKNOWN_ERROR", response.status);
  }
};

/**
 * Parse API response and extract data from success responses
 */
export const parseApiResponse = async <T = any>(response: Response): Promise<T> => {
  try {
    const responseData = (await response.json()) as ApiResponse<T>;

    if (!response.ok) {
      // Handle error responses
      if ("error" in responseData) {
        throw new Error(responseData.message || responseData.error);
      }
      throw new Error(`Request failed with status ${response.status}`);
    }

    // Handle success responses
    if ("success" in responseData && responseData.success === true) {
      return responseData.data as T;
    }

    // If it's already the expected data format (for auth endpoints that return AuthResponse directly)
    return responseData as T;
  } catch (parseError) {
    logger.error("Failed to parse API response", parseError);
    throw new Error("Invalid response format from server");
  }
};

/**
 * Enhanced fetch with proper response parsing
 */
export const authenticatedFetchWithParsing = async <T = any>(url: string, options: RequestInit = {}): Promise<T> => {
  const response = await authenticatedFetch(url, options);
  return parseApiResponse<T>(response);
};

/**
 * Enhanced fetch without auth with proper response parsing
 */
export const apiFetchWithParsing = async <T = any>(url: string, options: RequestInit = {}): Promise<T> => {
  const response = await apiFetch(url, options);
  return parseApiResponse<T>(response);
};

/**
 * Simple API client for making authenticated requests
 */
export const apiClient = {
  async get<T = unknown>(path: string): Promise<T> {
    return authenticatedFetchWithParsing<T>(`${API_BASE_URL}/api/v1${path}`);
  },

  async post<T = unknown>(path: string, body: unknown): Promise<T> {
    return authenticatedFetchWithParsing<T>(`${API_BASE_URL}/api/v1${path}`, {
      method: "POST",
      body: JSON.stringify(body),
    });
  },

  async put<T = unknown>(path: string, body: unknown): Promise<T> {
    return authenticatedFetchWithParsing<T>(`${API_BASE_URL}/api/v1${path}`, {
      method: "PUT",
      body: JSON.stringify(body),
    });
  },

  async delete<T = unknown>(path: string): Promise<T> {
    return authenticatedFetchWithParsing<T>(`${API_BASE_URL}/api/v1${path}`, {
      method: "DELETE",
    });
  },
};
