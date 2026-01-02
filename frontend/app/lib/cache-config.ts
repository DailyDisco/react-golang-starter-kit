/**
 * Centralized cache timing configuration for TanStack Query
 *
 * These values control how long data is considered "fresh" (staleTime)
 * before TanStack Query will refetch on the next access.
 */
export const CACHE_TIMES = {
  /** Health checks - need frequent updates (15 seconds) */
  HEALTH: 15 * 1000,

  /** User subscription status - may change but not frequently (2 minutes) */
  SUBSCRIPTION: 2 * 60 * 1000,

  /** User profile data - changes occasionally (1 minute) */
  USER_DATA: 60 * 1000,

  /** User preferences - changes occasionally (1 minute) */
  PREFERENCES: 60 * 1000,

  /** User sessions - may change (1 minute) */
  SESSIONS: 60 * 1000,

  /** API keys - rarely change (5 minutes) */
  API_KEYS: 5 * 60 * 1000,

  /** Login history - rarely changes (5 minutes) */
  LOGIN_HISTORY: 5 * 60 * 1000,

  /** Feature flags - rarely change during a session (5 minutes) */
  FEATURE_FLAGS: 5 * 60 * 1000,

  /** Files list - may change (5 minutes) */
  FILES: 5 * 60 * 1000,

  /** File URLs - can be cached longer (10 minutes) */
  FILE_URL: 10 * 60 * 1000,

  /** Storage status - rarely changes (30 minutes) */
  STORAGE_STATUS: 30 * 60 * 1000,

  /** Billing configuration - almost never changes (1 hour) */
  BILLING_CONFIG: 60 * 60 * 1000,

  /** Available billing plans - rarely change (10 minutes) */
  BILLING_PLANS: 10 * 60 * 1000,

  /** Announcements - may be updated by admins (1 minute) */
  ANNOUNCEMENTS: 60 * 1000,

  /** Site settings - rarely change (5 minutes) */
  SITE_SETTINGS: 5 * 60 * 1000,

  /** Organization data - may change (2 minutes) */
  ORGANIZATIONS: 2 * 60 * 1000,
} as const;

/**
 * Garbage collection times - how long to keep data in cache after it becomes unused
 * These should generally be longer than staleTime
 */
export const GC_TIMES = {
  /** Default garbage collection time (5 minutes) */
  DEFAULT: 5 * 60 * 1000,

  /** Feature flags - keep longer since they're important (30 minutes) */
  FEATURE_FLAGS: 30 * 60 * 1000,

  /** Billing config - keep even longer (1 hour) */
  BILLING: 60 * 60 * 1000,

  /** Short-lived data like health checks (1 minute) */
  SHORT: 60 * 1000,
} as const;

export type CacheTimeKey = keyof typeof CACHE_TIMES;
export type GCTimeKey = keyof typeof GC_TIMES;

/**
 * Stale-While-Revalidate configurations for stable data.
 *
 * These patterns show cached data immediately (staleTime: Infinity)
 * while refreshing in the background at regular intervals.
 * Ideal for data that rarely changes but should stay fresh.
 */
export const SWR_CONFIG = {
  /**
   * Stable data pattern: show cached forever, refresh every 5 minutes.
   * Use for: billing config, feature flags, site settings, billing plans.
   */
  STABLE: {
    staleTime: Infinity,
    refetchInterval: 5 * 60 * 1000, // 5 minutes background refresh
    refetchIntervalInBackground: false, // Only when tab is active
  },

  /**
   * Semi-stable data pattern: cache for 10 minutes, refresh every 15 minutes.
   * Use for: changelog, announcements (admin-controlled updates).
   */
  SEMI_STABLE: {
    staleTime: 10 * 60 * 1000, // 10 minutes before considered stale
    refetchInterval: 15 * 60 * 1000, // 15 minutes background refresh
    refetchIntervalInBackground: false,
  },
} as const;
