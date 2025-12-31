import { QueryClient } from "@tanstack/react-query";

import { ApiError } from "../services/api/client";
import { logger } from "./logger";

/**
 * Extract HTTP status code from error.
 * Checks ApiError.statusCode first, then falls back to parsing error message.
 */
const extractStatusCode = (error: Error): number | null => {
  // Check for ApiError with explicit statusCode
  if (error instanceof ApiError) {
    return error.statusCode;
  }

  // Check for statusCode property on generic errors
  if ("statusCode" in error && typeof error.statusCode === "number") {
    return error.statusCode;
  }

  if (!error.message) return null;

  // Match common patterns: "status 404", "status: 500", "with status 401", etc.
  const match = error.message.match(/\bstatus[:\s]+(\d{3})\b/i);
  if (match) {
    return parseInt(match[1], 10);
  }

  // Also check for "status code" pattern
  const codeMatch = error.message.match(/\bstatus\s+code[:\s]+(\d{3})\b/i);
  if (codeMatch) {
    return parseInt(codeMatch[1], 10);
  }

  return null;
};

/**
 * Check if error represents a 4xx client error (should not retry).
 * Returns true for 400-499 status codes.
 * Auth errors (401/403) are specifically not retried - the auth recovery
 * in client.ts handles token refresh separately.
 */
const isClientError = (error: Error): boolean => {
  const status = extractStatusCode(error);
  return status !== null && status >= 400 && status < 500;
};

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Cache data for 1 minute before considering it stale
      staleTime: 1000 * 60, // 1 minute
      // Keep unused data in cache for 5 minutes
      gcTime: 1000 * 60 * 5, // 5 minutes (formerly cacheTime)
      retry: (failureCount, error) => {
        // Don't retry on 4xx client errors (bad request, unauthorized, etc.)
        // These are client-side issues that won't be fixed by retrying
        if (error instanceof Error && isClientError(error)) {
          return false;
        }
        // Retry up to 2 times for other errors (network issues, 5xx server errors, etc.)
        return failureCount < 2;
      },
      // Enable refetch on window focus for better UX - data stays fresh
      refetchOnWindowFocus: true,
      // Refetch on network reconnection
      refetchOnReconnect: true,
      // Network mode - continue to show cached data if offline
      networkMode: "online",
    },
    mutations: {
      // Don't retry mutations by default - they could have side effects
      retry: false,
      // Network mode for mutations
      networkMode: "online",
      onError: (error) => {
        logger.error("Mutation error", error);
      },
    },
  },
});
