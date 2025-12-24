import { QueryClient } from "@tanstack/react-query";

import { logger } from "./logger";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Cache data for 1 minute before considering it stale
      staleTime: 1000 * 60, // 1 minute
      // Keep unused data in cache for 5 minutes
      gcTime: 1000 * 60 * 5, // 5 minutes (formerly cacheTime)
      retry: (failureCount, error) => {
        // Don't retry on 4xx client errors (bad request, unauthorized, etc.)
        if (error instanceof Error && error.message.includes("4")) {
          return false;
        }
        // Retry up to 2 times for other errors (network issues, 5xx, etc.)
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
