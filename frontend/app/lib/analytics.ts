import { track } from "@vercel/analytics";

import { logger } from "./logger";
import { addBreadcrumb, isSentryConfigured } from "./sentry";

// Event categories for organization
type EventCategory = "auth" | "navigation" | "feature" | "error" | "performance" | "engagement";

// Privacy-safe properties only (no PII)
type SafeEventProperties = Record<string, string | number | boolean | undefined>;

interface TrackEventOptions extends SafeEventProperties {
  category?: EventCategory;
}

/**
 * Sanitize paths to remove PII (IDs, slugs, etc.)
 */
export function sanitizePath(path: string): string {
  return path
    .replace(/\/\d+/g, "/:id") // /users/123 -> /users/:id
    .replace(/\/[a-f0-9-]{36}/g, "/:uuid") // UUID patterns
    .replace(/\/@[\w-]+/g, "/:slug"); // /orgs/@acme -> /orgs/:slug
}

/**
 * Track a custom event across all analytics providers
 * @param eventName - Snake_case event name (e.g., 'button_click', 'feature_used')
 * @param properties - Event properties (must not contain PII)
 */
export function trackEvent(eventName: string, properties?: TrackEventOptions): void {
  const { category = "feature", ...eventProperties } = properties || {};

  // 1. Log locally (development visibility)
  logger.debug(`Analytics: ${eventName}`, eventProperties);

  // 2. Report to Sentry as breadcrumb
  if (isSentryConfigured()) {
    addBreadcrumb({
      message: eventName,
      category: `analytics.${category}`,
      data: eventProperties,
    });
  }

  // 3. Report to Vercel Analytics
  if (typeof window !== "undefined") {
    track(eventName, eventProperties);
  }
}

/**
 * Pre-defined analytics functions for common events
 */
export const analytics = {
  // Auth events
  login: (method: "email" | "oauth" = "email") => trackEvent("user_login", { category: "auth", method }),

  logout: () => trackEvent("user_logout", { category: "auth" }),

  register: (method: "email" | "oauth" = "email") => trackEvent("user_register", { category: "auth", method }),

  // Feature usage
  featureUsed: (featureName: string, metadata?: SafeEventProperties) =>
    trackEvent("feature_used", { category: "feature", feature: featureName, ...metadata }),

  // Navigation
  pageView: (path: string) =>
    trackEvent("page_view", {
      category: "navigation",
      path: sanitizePath(path),
    }),

  // Errors (supplement Sentry's automatic capture)
  errorOccurred: (errorType: string, errorCode?: string) =>
    trackEvent("error_occurred", { category: "error", errorType, errorCode }),

  // Performance
  performanceMark: (markName: string, durationMs: number) =>
    trackEvent("performance_mark", { category: "performance", mark: markName, duration: durationMs }),

  // Engagement
  buttonClick: (buttonId: string, context?: string) =>
    trackEvent("button_click", { category: "engagement", buttonId, context }),

  // Subscription events
  subscriptionCreated: (plan: string) => trackEvent("subscription_created", { category: "feature", plan }),

  subscriptionCanceled: (plan: string) => trackEvent("subscription_canceled", { category: "feature", plan }),

  // File events
  fileUploaded: (fileType: string, sizeKb: number) =>
    trackEvent("file_uploaded", { category: "feature", fileType, sizeKb }),

  fileDeleted: () => trackEvent("file_deleted", { category: "feature" }),

  // API key events
  apiKeyCreated: () => trackEvent("api_key_created", { category: "feature" }),

  apiKeyDeleted: () => trackEvent("api_key_deleted", { category: "feature" }),

  // 2FA events
  twoFactorEnabled: () => trackEvent("2fa_enabled", { category: "auth" }),

  twoFactorDisabled: () => trackEvent("2fa_disabled", { category: "auth" }),
};
