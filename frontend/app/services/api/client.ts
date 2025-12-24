import { logger } from "../../lib/logger";

// Create API_BASE_URL safely for SSR
const getApiBaseUrl = () => {
  // In SSR, import.meta.env might not be available or populated
  if (typeof window === "undefined") {
    return "http://localhost:8080";
  }
  return import.meta.env.VITE_API_URL || "http://localhost:8080";
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
    const response = await fetch(`${API_BASE_URL}/api/csrf-token`, {
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

// Enhanced fetch with auth (uses httpOnly cookies via credentials: "include")
export const authenticatedFetch = async (url: string, options: RequestInit = {}): Promise<Response> => {
  const needsCSRF = isStateChangingMethod(options.method);
  const headers = createHeaders(needsCSRF);

  const response = await fetch(url, {
    ...options,
    credentials: "include", // Include httpOnly cookies for authentication
    headers: {
      ...headers,
      ...options.headers,
    },
  });

  // Handle 401 Unauthorized - session might be expired
  if (response.status === 401) {
    // Clear auth data from localStorage (user info for UI)
    if (typeof window !== "undefined") {
      localStorage.removeItem("auth_user");
      localStorage.removeItem("refresh_token");
      // Redirect to login if on a protected page
      if (window.location.pathname !== "/login") {
        window.location.href = "/login";
      }
    }
  }

  return response;
};

// Simple fetch for auth endpoints (login/register) - includes credentials for cookie handling
export const apiFetch = async (url: string, options: RequestInit = {}): Promise<Response> => {
  const needsCSRF = isStateChangingMethod(options.method);
  const headers = createHeaders(needsCSRF);

  return fetch(url, {
    ...options,
    credentials: "include", // Include httpOnly cookies
    headers: {
      ...headers,
      ...options.headers,
    },
  });
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
 * Safely parse error response from server
 */
export const parseErrorResponse = async (response: Response, defaultMessage: string): Promise<string> => {
  try {
    // Get response as text first, then try to parse as JSON
    const responseText = await response.text();

    // Try to parse as JSON if it looks like JSON
    if (responseText.trim().startsWith("{") || responseText.trim().startsWith("[")) {
      const errorData = JSON.parse(responseText) as ApiErrorResponse;
      return errorData.error || errorData.message || defaultMessage;
    } else {
      // Not JSON, use the text directly
      return responseText || `${defaultMessage} with status ${response.status}`;
    }
  } catch (parseError) {
    // If anything fails, use a generic error message
    logger.error("Failed to parse error response", parseError);
    return `${defaultMessage} with status ${response.status}`;
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
    return authenticatedFetchWithParsing<T>(`${API_BASE_URL}/api${path}`);
  },

  async post<T = unknown>(path: string, body: unknown): Promise<T> {
    return authenticatedFetchWithParsing<T>(`${API_BASE_URL}/api${path}`, {
      method: "POST",
      body: JSON.stringify(body),
    });
  },

  async put<T = unknown>(path: string, body: unknown): Promise<T> {
    return authenticatedFetchWithParsing<T>(`${API_BASE_URL}/api${path}`, {
      method: "PUT",
      body: JSON.stringify(body),
    });
  },

  async delete<T = unknown>(path: string): Promise<T> {
    return authenticatedFetchWithParsing<T>(`${API_BASE_URL}/api${path}`, {
      method: "DELETE",
    });
  },
};
