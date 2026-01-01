import { useEffect, useRef } from "react";

import { useCommandContext } from "@/hooks/useCommandContext";
import { useCommandPaletteStore } from "@/hooks/useCommandPalette";
import { commandRegistry } from "@/services/command-palette";
import type { SearchResult } from "@/services/command-palette/types";

/**
 * Hook for managing async command search with debouncing
 *
 * Automatically performs search when the search query changes,
 * respecting the debounce settings of each search provider.
 */
export function useCommandSearch() {
  const { search, mode, setSearchResults, setSearching, setSearchError } = useCommandPaletteStore();
  const ctx = useCommandContext();

  // Track the latest search to avoid race conditions
  const searchIdRef = useRef(0);
  const abortControllerRef = useRef<AbortController | null>(null);

  useEffect(() => {
    // Only search in certain modes
    if (mode !== "impersonate" && mode !== "flag-toggle" && mode !== "search") {
      return;
    }

    // Clear results if search is too short
    if (search.length < 1) {
      setSearchResults([]);
      setSearching(false);
      return;
    }

    // Cancel previous search
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    // Create new abort controller for this search
    const abortController = new AbortController();
    abortControllerRef.current = abortController;

    // Increment search ID to track latest search
    const currentSearchId = ++searchIdRef.current;

    const performSearch = async () => {
      setSearching(true);
      setSearchError(null);

      try {
        let results: SearchResult[] = [];

        // Get providers based on mode
        if (mode === "impersonate") {
          // Only search users in impersonate mode
          const providers = commandRegistry.getSearchProviders(ctx);
          const userProvider = providers.find((p) => p.id === "users");

          if (userProvider && search.length >= (userProvider.minQueryLength ?? 2)) {
            results = await userProvider.search(search, ctx);
          }
        } else if (mode === "flag-toggle") {
          // Only search feature flags in flag-toggle mode
          const providers = commandRegistry.getSearchProviders(ctx);
          const flagProvider = providers.find((p) => p.id === "feature-flags");

          if (flagProvider && search.length >= (flagProvider.minQueryLength ?? 1)) {
            results = await flagProvider.search(search, ctx);
          }
        } else {
          // Full search across all providers
          results = await commandRegistry.search(search, ctx, {
            signal: abortController.signal,
          });
        }

        // Only update if this is still the latest search
        if (currentSearchId === searchIdRef.current && !abortController.signal.aborted) {
          setSearchResults(results);
        }
      } catch (error) {
        // Ignore abort errors
        if (error instanceof DOMException && error.name === "AbortError") {
          return;
        }

        // Only update if this is still the latest search
        if (currentSearchId === searchIdRef.current) {
          setSearchError(error instanceof Error ? error : new Error("Search failed"));
        }
      } finally {
        // Only update if this is still the latest search
        if (currentSearchId === searchIdRef.current && !abortController.signal.aborted) {
          setSearching(false);
        }
      }
    };

    // Debounce the search
    const debounceMs = mode === "flag-toggle" ? 200 : 300;
    const timeoutId = setTimeout(performSearch, debounceMs);

    return () => {
      clearTimeout(timeoutId);
      abortController.abort();
    };
  }, [search, mode, ctx, setSearchResults, setSearching, setSearchError]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);
}

/**
 * Hook for registering search providers on mount
 */
export function useSearchProviders() {
  useEffect(() => {
    // Import and register search providers
    const registerProviders = async () => {
      const { userSearchProvider, featureFlagSearchProvider, auditLogSearchProvider, pageSearchProvider } =
        await import("@/services/command-palette/searchProviders");

      const unregisterUser = commandRegistry.registerSearchProvider(userSearchProvider);
      const unregisterFlag = commandRegistry.registerSearchProvider(featureFlagSearchProvider);
      const unregisterAudit = commandRegistry.registerSearchProvider(auditLogSearchProvider);
      const unregisterPage = commandRegistry.registerSearchProvider(pageSearchProvider);

      return () => {
        unregisterUser();
        unregisterFlag();
        unregisterAudit();
        unregisterPage();
      };
    };

    let cleanup: (() => void) | undefined;
    registerProviders().then((fn) => {
      cleanup = fn;
    });

    return () => {
      cleanup?.();
    };
  }, []);
}
