import { inject } from "@vercel/analytics";

/**
 * Initialize Vercel Analytics
 * Only runs in browser environment
 */
export function initVercelAnalytics(): void {
  if (typeof window === "undefined") return;

  inject({
    mode: import.meta.env.PROD ? "production" : "development",
    debug: import.meta.env.DEV,
  });
}
