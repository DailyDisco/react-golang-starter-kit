/**
 * Sentry Error Tracking Integration
 *
 * This module provides optional Sentry integration for error tracking.
 * Sentry is only initialized if VITE_SENTRY_DSN environment variable is set.
 *
 * Usage:
 * 1. Set VITE_SENTRY_DSN in your environment
 * 2. Import and call initSentry() early in your app
 * 3. Errors are automatically captured
 *
 * For manual error reporting:
 *   captureError(new Error('Something went wrong'))
 *   captureMessage('User did something unexpected')
 */

import { logger } from "./logger";

// Sentry SDK reference - dynamically imported when needed
// Using 'any' to avoid static type dependency on @sentry/react
// eslint-disable-next-line @typescript-eslint/no-explicit-any
let Sentry: any = null;
let isInitialized = false;

/**
 * Check if Sentry is configured (DSN is provided)
 */
export function isSentryConfigured(): boolean {
  return Boolean(import.meta.env.VITE_SENTRY_DSN);
}

/**
 * Initialize Sentry error tracking
 * This should be called as early as possible in your app, ideally before React renders.
 *
 * @returns Promise that resolves when Sentry is initialized (or skipped if not configured)
 */
export async function initSentry(): Promise<void> {
  if (isInitialized) {
    return;
  }

  const dsn = import.meta.env.VITE_SENTRY_DSN;

  if (!dsn) {
    logger.debug("Sentry DSN not configured, skipping initialization");
    return;
  }

  try {
    // Dynamically import Sentry to avoid bundling if not used
    Sentry = await import("@sentry/react");

    Sentry.init({
      dsn,
      environment: import.meta.env.MODE || "development",
      release: import.meta.env.VITE_APP_VERSION || "1.0.0",

      // Performance monitoring
      tracesSampleRate: import.meta.env.PROD ? 0.1 : 1.0,

      // Session replay (adjust as needed for your use case)
      replaysSessionSampleRate: 0.1,
      replaysOnErrorSampleRate: 1.0,

      // Integrations
      integrations: [
        Sentry.browserTracingIntegration(),
        Sentry.replayIntegration({
          // Mask all text content for privacy
          maskAllText: false,
          // Block all media for privacy
          blockAllMedia: false,
        }),
      ],

      // Don't send errors in development unless explicitly configured
      enabled: import.meta.env.PROD || Boolean(import.meta.env.VITE_SENTRY_ENABLE_DEV),

      // Filter out noisy errors
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      beforeSend(event: any, hint: any) {
        const error = hint?.originalException;

        // Filter out certain error types
        if (error instanceof Error) {
          // Skip network errors that are likely just connectivity issues
          if (error.message.includes("Failed to fetch") || error.message.includes("NetworkError")) {
            return null;
          }

          // Skip aborted requests
          if (error.name === "AbortError") {
            return null;
          }
        }

        return event;
      },

      // Add additional context
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      initialScope: (scope: any) => {
        // Add custom tags
        scope.setTag("app.version", import.meta.env.VITE_APP_VERSION || "unknown");

        return scope;
      },
    });

    isInitialized = true;
    logger.info("Sentry initialized successfully");
  } catch (error) {
    // Don't crash the app if Sentry fails to initialize
    logger.error("Failed to initialize Sentry", error);
  }
}

/**
 * Capture an exception and send it to Sentry
 */
export function captureError(error: Error, context?: Record<string, unknown>): void {
  if (Sentry && isInitialized) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    Sentry.withScope((scope: any) => {
      if (context) {
        scope.setExtras(context);
      }
      Sentry.captureException(error);
    });
  }

  // Always log locally too
  logger.error("Captured error", error, context);
}

/**
 * Capture a message and send it to Sentry
 */
export function captureMessage(message: string, level: "info" | "warning" | "error" = "info"): void {
  if (Sentry && isInitialized) {
    Sentry.captureMessage(message, level);
  }

  // Log locally
  if (level === "error") {
    logger.error(message);
  } else if (level === "warning") {
    logger.warn(message);
  } else {
    logger.info(message);
  }
}

/**
 * Set user context for Sentry
 * Call this after the user logs in
 */
export function setUser(user: { id: string | number; email?: string; username?: string } | null): void {
  if (Sentry && isInitialized) {
    if (user) {
      Sentry.setUser({
        id: String(user.id),
        email: user.email,
        username: user.username,
      });
    } else {
      Sentry.setUser(null);
    }
  }
}

/**
 * Add a breadcrumb for debugging
 */
export function addBreadcrumb(breadcrumb: {
  message: string;
  category?: string;
  data?: Record<string, unknown>;
}): void {
  if (Sentry && isInitialized) {
    Sentry.addBreadcrumb({
      message: breadcrumb.message,
      category: breadcrumb.category || "app",
      data: breadcrumb.data,
      level: "info",
    });
  }
}

/**
 * Set custom context for all future events
 */
export function setContext(name: string, context: Record<string, unknown>): void {
  if (Sentry && isInitialized) {
    Sentry.setContext(name, context);
  }
}

/**
 * Get the Sentry ErrorBoundary component if available
 * Use this to wrap parts of your app to catch and report errors
 *
 * @returns Sentry.ErrorBoundary or null if Sentry isn't available
 */
export function getSentryErrorBoundary() {
  if (Sentry && isInitialized) {
    return Sentry.ErrorBoundary;
  }
  return null;
}

/**
 * Start a new transaction for performance monitoring
 */
export function startTransaction(name: string, op: string): (() => void) | undefined {
  if (Sentry && isInitialized) {
    const span = Sentry.startInactiveSpan({ name, op });
    return () => span?.end();
  }
  return undefined;
}
