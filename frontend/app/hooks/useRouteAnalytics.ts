import { useEffect, useRef } from "react";

import { useRouter } from "@tanstack/react-router";

import { analytics, sanitizePath } from "../lib/analytics";

/**
 * Hook that tracks route changes for analytics
 * Should be used in a component near the root of the app
 */
export function useRouteAnalytics(): void {
  const router = useRouter();
  const previousPath = useRef<string | null>(null);

  useEffect(() => {
    // Track initial page load
    const initialPath = router.state.location.pathname;
    if (initialPath && initialPath !== previousPath.current) {
      analytics.pageView(initialPath);
      previousPath.current = initialPath;
    }

    // Subscribe to route changes
    const unsubscribe = router.subscribe("onResolved", (event) => {
      const currentPath = event.toLocation.pathname;

      // Avoid duplicate tracking
      if (currentPath === previousPath.current) return;

      // Track page view with sanitized path
      analytics.pageView(currentPath);
      previousPath.current = currentPath;
    });

    return unsubscribe;
  }, [router]);
}

/**
 * Component that initializes route analytics
 * Use this in __root.tsx alongside other initializers
 */
export function RouteAnalyticsInitializer(): null {
  useRouteAnalytics();
  return null;
}
