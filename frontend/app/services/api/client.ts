import { logger } from "../../lib/logger";

/**
 * Custom error class that preserves API error codes for contextual handling
 */
export class ApiError extends Error {
  code: string;
  statusCode: number;
  requestId?: string;
  /** Retry-After value in seconds (from 429 responses) */
  retryAfter?: number;

  constructor(message: string, code: string, statusCode: number, requestId?: string, retryAfter?: number) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.statusCode = statusCode;
    this.requestId = requestId;
    this.retryAfter = retryAfter;
  }
}

/**
 * Custom error for request timeouts
 */
export class TimeoutError extends Error {
  constructor(message: string = "Request timed out") {
    super(message);
    this.name = "TimeoutError";
  }
}

// Request timeout configuration (in milliseconds)
const DEFAULT_TIMEOUT_MS = 30000; // 30 seconds
const UPLOAD_TIMEOUT_MS = 120000; // 2 minutes for file uploads

/**
 * Fetch with timeout support using AbortController
 * Automatically aborts requests that exceed the timeout duration
 */
const fetchWithTimeout = async (url: string, options: RequestInit & { timeout?: number } = {}): Promise<Response> => {
  const { timeout = DEFAULT_TIMEOUT_MS, ...fetchOptions } = options;

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  try {
    const response = await fetch(url, {
      ...fetchOptions,
      signal: controller.signal,
    });
    return response;
  } catch (error) {
    if (error instanceof Error && error.name === "AbortError") {
      throw new TimeoutError(
        `Request to ${new URL(url, window?.location?.origin || "http://localhost").pathname} timed out after ${timeout}ms`
      );
    }
    throw error;
  } finally {
    clearTimeout(timeoutId);
  }
};

// Create API_BASE_URL safely for SSR and remote development
const getApiBaseUrl = () => {
  // In SSR, use environment variable for server-side API calls
  // SSR_API_URL allows internal Docker network routing (e.g., http://backend:8080)
  if (typeof window === "undefined") {
    return process.env.SSR_API_URL || process.env.VITE_API_URL || "http://localhost:8080";
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
// Uses a queue-based approach to batch all 401 requests and retry them together
let refreshPromise: Promise<boolean> | null = null;
type QueuedRequest = { resolve: (response: Response) => void; retry: () => Promise<Response> };
const failedRequestsQueue: QueuedRequest[] = [];

// Circuit breaker configuration
const MAX_CONSECUTIVE_401_FAILURES = 3;
const CIRCUIT_BREAKER_RESET_MS = 10000; // 10 seconds
const CIRCUIT_BREAKER_KEY = "auth_circuit_breaker";

/**
 * Circuit breaker state stored in sessionStorage for tab isolation
 * Each browser tab maintains its own circuit breaker state
 */
interface CircuitBreakerState {
  failures: number;
  resetAt: number | null;
}

const getCircuitBreakerState = (): CircuitBreakerState => {
  if (typeof window === "undefined") {
    return { failures: 0, resetAt: null };
  }
  try {
    const stored = sessionStorage.getItem(CIRCUIT_BREAKER_KEY);
    if (stored) {
      const state = JSON.parse(stored) as CircuitBreakerState;
      // Check if reset time has passed
      if (state.resetAt && Date.now() >= state.resetAt) {
        // Auto-reset expired circuit breaker
        sessionStorage.removeItem(CIRCUIT_BREAKER_KEY);
        return { failures: 0, resetAt: null };
      }
      return state;
    }
  } catch {
    // Ignore parse errors
  }
  return { failures: 0, resetAt: null };
};

const setCircuitBreakerState = (state: CircuitBreakerState): void => {
  if (typeof window === "undefined") return;
  sessionStorage.setItem(CIRCUIT_BREAKER_KEY, JSON.stringify(state));
};

/**
 * Reset the 401 circuit breaker counter
 * Called when authentication succeeds
 */
export const resetAuthCircuitBreaker = (): void => {
  if (typeof window === "undefined") return;
  sessionStorage.removeItem(CIRCUIT_BREAKER_KEY);
};

/**
 * Increment circuit breaker failure count
 */
const incrementCircuitBreaker = (): void => {
  const state = getCircuitBreakerState();
  const newState: CircuitBreakerState = {
    failures: state.failures + 1,
    resetAt: Date.now() + CIRCUIT_BREAKER_RESET_MS,
  };
  setCircuitBreakerState(newState);
};

/**
 * Check if the circuit breaker is tripped
 */
const isCircuitBreakerOpen = (): boolean => {
  const state = getCircuitBreakerState();
  return state.failures >= MAX_CONSECUTIVE_401_FAILURES;
};

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
// Includes automatic retry on CSRF token errors, 401 token refresh, and request timeouts
export const authenticatedFetch = async (
  url: string,
  options: RequestInit & { timeout?: number } = {}
): Promise<Response> => {
  const needsCSRF = isStateChangingMethod(options.method);
  let headers = createHeaders(needsCSRF);
  const { timeout, ...fetchOptions } = options;

  const response = await fetchWithTimeout(url, {
    ...fetchOptions,
    timeout,
    credentials: "include", // Include httpOnly cookies for authentication
    headers: {
      ...headers,
      ...fetchOptions.headers,
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
        return fetchWithTimeout(url, {
          ...fetchOptions,
          timeout,
          credentials: "include",
          headers: {
            ...headers,
            ...fetchOptions.headers,
          },
        });
      }
    } catch {
      // If we can't parse the error, just return the original response
    }
  }

  // Handle 401 Unauthorized - queue request and retry after token refresh
  if (response.status === 401) {
    // Don't try to refresh on auth endpoints to avoid infinite loops
    if (url.includes("/auth/login") || url.includes("/auth/refresh") || url.includes("/auth/register")) {
      return response;
    }

    // Circuit breaker: stop retrying after too many consecutive 401 failures
    if (isCircuitBreakerOpen()) {
      logger.warn("Circuit breaker open: too many consecutive 401s, skipping retry");
      if (typeof window !== "undefined" && window.location.pathname !== "/login" && !isInAuthGracePeriod()) {
        window.dispatchEvent(new CustomEvent("session-expired"));
      }
      return response;
    }

    // Create a retry function for this request
    const retryRequest = async (): Promise<Response> => {
      const retryHeaders = createHeaders(needsCSRF);
      return fetchWithTimeout(url, {
        ...fetchOptions,
        timeout,
        credentials: "include",
        headers: {
          ...retryHeaders,
          ...fetchOptions.headers,
        },
      });
    };

    // Queue this request and return a promise that resolves when it's retried
    return new Promise<Response>((resolve) => {
      failedRequestsQueue.push({ resolve, retry: retryRequest });

      // If no refresh is in progress, start one
      if (!refreshPromise) {
        const startRefresh = async () => {
          try {
            const AuthService = await getAuthService();
            const refreshed = await AuthService.initializeFromStorage();

            // Delay to ensure cookies are fully propagated through Docker/Vite proxy
            await new Promise((resolve) => setTimeout(resolve, 200));

            // Process all queued requests
            const queue = [...failedRequestsQueue];
            failedRequestsQueue.length = 0;

            if (refreshed) {
              // Reset circuit breaker on successful refresh
              resetAuthCircuitBreaker();
              // Retry all queued requests with new token
              for (const request of queue) {
                try {
                  let retryResponse = await request.retry();

                  // If retry still gets 401, wait longer for cookie propagation and try once more
                  if (retryResponse.status === 401) {
                    logger.info("Retry got 401, waiting for cookie propagation...");
                    await new Promise((resolve) => setTimeout(resolve, 150));
                    retryResponse = await request.retry();
                  }

                  request.resolve(retryResponse);
                } catch (retryError) {
                  logger.warn("Failed to retry request after token refresh", { error: retryError });
                  // Resolve with an error response
                  request.resolve(
                    new Response(JSON.stringify({ error: "RETRY_FAILED", message: "Failed to retry request" }), {
                      status: 500,
                      headers: { "Content-Type": "application/json" },
                    })
                  );
                }
              }
            } else {
              // Increment circuit breaker counter on refresh failure
              incrementCircuitBreaker();
              // Refresh failed - dispatch session-expired and reject all queued requests
              if (typeof window !== "undefined" && window.location.pathname !== "/login" && !isInAuthGracePeriod()) {
                window.dispatchEvent(new CustomEvent("session-expired"));
              }
              // Return original 401 response to all queued requests
              for (const request of queue) {
                request.resolve(response);
              }
            }
          } catch (refreshError) {
            logger.warn("Token refresh failed during 401 recovery", { error: refreshError });
            // Increment circuit breaker counter on refresh exception
            incrementCircuitBreaker();
            // Dispatch session-expired event
            if (typeof window !== "undefined" && window.location.pathname !== "/login" && !isInAuthGracePeriod()) {
              window.dispatchEvent(new CustomEvent("session-expired"));
            }
            // Return original 401 response to all queued requests
            const queue = [...failedRequestsQueue];
            failedRequestsQueue.length = 0;
            for (const request of queue) {
              request.resolve(response);
            }
          } finally {
            refreshPromise = null;
          }
        };

        refreshPromise = startRefresh().then(() => true);
      }
    });
  }

  // Handle rate limiting - log for debugging
  if (response.status === 429) {
    const retryAfter = response.headers.get("Retry-After");
    logger.warn("Rate limited", { retryAfter, endpoint: url });
  }

  return response;
};

// Simple fetch for auth endpoints (login/register) - includes credentials for cookie handling
// Includes automatic retry on CSRF token errors and request timeouts
export const apiFetch = async (url: string, options: RequestInit & { timeout?: number } = {}): Promise<Response> => {
  const needsCSRF = isStateChangingMethod(options.method);
  let headers = createHeaders(needsCSRF);
  const { timeout, ...fetchOptions } = options;

  const response = await fetchWithTimeout(url, {
    ...fetchOptions,
    timeout,
    credentials: "include", // Include httpOnly cookies
    headers: {
      ...headers,
      ...fetchOptions.headers,
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
        return fetchWithTimeout(url, {
          ...fetchOptions,
          timeout,
          credentials: "include",
          headers: {
            ...headers,
            ...fetchOptions.headers,
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
  // Extract Retry-After header for rate limit errors
  let retryAfter: number | undefined;
  if (response.status === 429) {
    const retryAfterHeader = response.headers.get("Retry-After");
    if (retryAfterHeader) {
      // Retry-After can be seconds or HTTP-date; we only handle seconds
      const parsed = parseInt(retryAfterHeader, 10);
      if (!isNaN(parsed)) {
        retryAfter = parsed;
      }
    }
  }

  try {
    // Get response as text first, then try to parse as JSON
    const responseText = await response.text();

    // Try to parse as JSON if it looks like JSON
    if (responseText.trim().startsWith("{") || responseText.trim().startsWith("[")) {
      const errorData = JSON.parse(responseText) as ApiErrorResponse;
      const message = errorData.message || errorData.error || defaultMessage;
      const code = errorData.error || "UNKNOWN_ERROR";
      return new ApiError(message, code, response.status, errorData.request_id, retryAfter);
    } else {
      // Not JSON, use the text directly
      const message = responseText || `${defaultMessage} with status ${response.status}`;
      return new ApiError(message, "UNKNOWN_ERROR", response.status, undefined, retryAfter);
    }
  } catch (parseError) {
    // If anything fails, use a generic error message
    logger.error("Failed to parse error response", parseError);
    return new ApiError(
      `${defaultMessage} with status ${response.status}`,
      "UNKNOWN_ERROR",
      response.status,
      undefined,
      retryAfter
    );
  }
};

/**
 * Parse API response and extract data from success responses
 */
export const parseApiResponse = async <T = any>(response: Response): Promise<T> => {
  // Handle error responses first, before trying to parse JSON
  if (!response.ok) {
    const error = await parseErrorResponse(response, "Request failed");
    throw error;
  }

  // For successful responses, parse JSON
  try {
    const responseData = (await response.json()) as ApiResponse<T>;

    // Handle success responses with { success: true, data: ... } format
    if ("success" in responseData && responseData.success === true) {
      return responseData.data as T;
    }

    // If it's already the expected data format (for auth endpoints that return AuthResponse directly)
    return responseData as T;
  } catch (parseError) {
    logger.error("Failed to parse API response", parseError);
    throw new ApiError("Invalid response format from server", "PARSE_ERROR", response.status);
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
