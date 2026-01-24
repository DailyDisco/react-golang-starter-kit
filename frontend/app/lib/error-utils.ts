import type { ApiError } from "../services/api/client";

export type ErrorCategory = "auth" | "validation" | "network" | "server" | "unknown";

export interface CategorizedError {
  category: ErrorCategory;
  message: string;
  details?: string;
  retryable: boolean;
  code?: string;
  statusCode?: number;
  /** Seconds until retry is allowed (for 429 rate limit errors) */
  retryAfter?: number;
}

// WeakMap cache for memoizing categorization of error objects
// Using WeakMap allows garbage collection when error objects are no longer referenced
const errorCache = new WeakMap<object, CategorizedError>();

/**
 * Categorize an error for consistent handling across the application.
 * Results are memoized per error object to avoid repeated categorization.
 *
 * Categories:
 * - auth: 401/403 errors requiring re-authentication or lacking permissions
 * - validation: 400/422 errors from invalid input
 * - network: Connection failures, timeouts
 * - server: 500+ errors from backend issues
 * - unknown: Everything else
 */
export function categorizeError(error: Error | ApiError | unknown): CategorizedError {
  // Check cache for object errors
  if (error && typeof error === "object") {
    const cached = errorCache.get(error);
    if (cached) {
      return cached;
    }
  }

  // Helper to cache and return result for object errors
  const cacheAndReturn = (result: CategorizedError): CategorizedError => {
    if (error && typeof error === "object") {
      errorCache.set(error, result);
    }
    return result;
  };

  // Handle non-Error objects
  if (!error || typeof error !== "object") {
    return {
      category: "unknown",
      message: String(error) || "An unexpected error occurred",
      retryable: true,
    };
  }

  const err = error as Error & { statusCode?: number; code?: string };

  // Check if it's an ApiError with status code
  if ("statusCode" in err && typeof err.statusCode === "number") {
    // Auth errors (401, 403)
    if (err.statusCode === 401) {
      return cacheAndReturn({
        category: "auth",
        message: "Please log in to continue",
        details: err.message,
        retryable: false,
        code: err.code,
        statusCode: err.statusCode,
      });
    }

    if (err.statusCode === 403) {
      return cacheAndReturn({
        category: "auth",
        message: "You don't have permission for this action",
        details: err.message,
        retryable: false,
        code: err.code,
        statusCode: err.statusCode,
      });
    }

    // Validation errors (400, 422)
    if (err.statusCode === 400 || err.statusCode === 422) {
      return cacheAndReturn({
        category: "validation",
        message: err.message || "Please check your input",
        retryable: false,
        code: err.code,
        statusCode: err.statusCode,
      });
    }

    // Rate limiting (429)
    if (err.statusCode === 429) {
      const retryAfter = "retryAfter" in err ? (err as { retryAfter?: number }).retryAfter : undefined;
      return cacheAndReturn({
        category: "server",
        message: retryAfter
          ? `Too many requests. Please wait ${retryAfter} seconds.`
          : "Too many requests. Please wait and try again.",
        details: err.message,
        retryable: true,
        code: err.code,
        statusCode: err.statusCode,
        retryAfter,
      });
    }

    // Server errors (500+)
    if (err.statusCode >= 500) {
      return cacheAndReturn({
        category: "server",
        message: "Something went wrong on our end",
        details: err.message,
        retryable: true,
        code: err.code,
        statusCode: err.statusCode,
      });
    }

    // Other 4xx errors
    if (err.statusCode >= 400 && err.statusCode < 500) {
      return cacheAndReturn({
        category: "validation",
        message: err.message || "Request failed",
        retryable: false,
        code: err.code,
        statusCode: err.statusCode,
      });
    }
  }

  // Network errors (check message patterns)
  const message = err.message?.toLowerCase() || "";
  if (
    message.includes("failed to fetch") ||
    message.includes("network") ||
    message.includes("fetch failed") ||
    message.includes("networkerror") ||
    message.includes("timeout") ||
    message.includes("econnrefused") ||
    message.includes("enotfound") ||
    message.includes("connection refused")
  ) {
    return cacheAndReturn({
      category: "network",
      message: "Unable to connect. Check your internet connection.",
      retryable: true,
    });
  }

  // Default unknown error
  return cacheAndReturn({
    category: "unknown",
    message: err.message || "An unexpected error occurred",
    retryable: true,
  });
}

/**
 * Check if an error is a specific category
 */
export function isAuthError(error: Error | ApiError | unknown): boolean {
  return categorizeError(error).category === "auth";
}

export function isValidationError(error: Error | ApiError | unknown): boolean {
  return categorizeError(error).category === "validation";
}

export function isNetworkError(error: Error | ApiError | unknown): boolean {
  return categorizeError(error).category === "network";
}

export function isServerError(error: Error | ApiError | unknown): boolean {
  return categorizeError(error).category === "server";
}

export function isRetryableError(error: Error | ApiError | unknown): boolean {
  return categorizeError(error).retryable;
}
