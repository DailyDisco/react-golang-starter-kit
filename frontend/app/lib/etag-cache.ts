/**
 * ETag cache for conditional HTTP requests.
 *
 * This module provides infrastructure for storing ETags and handling
 * 304 Not Modified responses to reduce bandwidth on cache revalidation.
 *
 * Usage with TanStack Query:
 * - Store ETags when responses are received
 * - Send If-None-Match header on refetches
 * - Handle 304 by returning previous data
 */

// In-memory ETag storage (persists across requests, cleared on page reload)
const etagCache = new Map<string, string>();

/**
 * Get a stored ETag for a given cache key.
 * @param key - The cache key (typically the URL or query key)
 */
export function getStoredETag(key: string): string | undefined {
  return etagCache.get(key);
}

/**
 * Store an ETag for a given cache key.
 * @param key - The cache key (typically the URL or query key)
 * @param etag - The ETag value from the response header
 */
export function storeETag(key: string, etag: string): void {
  etagCache.set(key, etag);
}

/**
 * Remove an ETag from the cache.
 * @param key - The cache key to remove
 */
export function clearETag(key: string): void {
  etagCache.delete(key);
}

/**
 * Clear all stored ETags.
 * Call this on logout to prevent stale ETags.
 */
export function clearAllETags(): void {
  etagCache.clear();
}

/**
 * Generate a cache key from a TanStack Query key.
 * @param queryKey - The query key array
 */
export function queryKeyToETagKey(queryKey: readonly unknown[]): string {
  return JSON.stringify(queryKey);
}

/**
 * Custom error thrown when a 304 Not Modified response is received.
 * TanStack Query can catch this and use the cached data.
 */
export class NotModifiedError extends Error {
  constructor() {
    super("Not Modified");
    this.name = "NotModifiedError";
  }
}

/**
 * Check if an error is a NotModifiedError.
 */
export function isNotModifiedError(error: unknown): error is NotModifiedError {
  return error instanceof NotModifiedError;
}

/**
 * Fetch with ETag support.
 * Sends If-None-Match header if we have a cached ETag.
 * Stores new ETags from responses.
 * Throws NotModifiedError on 304 responses.
 *
 * @param url - The URL to fetch
 * @param cacheKey - The cache key for ETag storage
 * @param options - Fetch options
 */
export async function fetchWithETag(url: string, cacheKey: string, options: RequestInit = {}): Promise<Response> {
  const headers = new Headers(options.headers);

  // Add If-None-Match header if we have a cached ETag
  const cachedETag = getStoredETag(cacheKey);
  if (cachedETag) {
    headers.set("If-None-Match", cachedETag);
  }

  const response = await fetch(url, {
    ...options,
    headers,
  });

  // Handle 304 Not Modified
  if (response.status === 304) {
    throw new NotModifiedError();
  }

  // Store new ETag if present
  const newETag = response.headers.get("ETag");
  if (newETag) {
    storeETag(cacheKey, newETag);
  }

  return response;
}

/**
 * Create a TanStack Query-compatible fetch function with ETag support.
 * Returns previous data on 304 responses.
 *
 * @param fetchFn - The original fetch function
 * @param cacheKey - The cache key for ETag storage
 * @param getPreviousData - Function to get previously cached data
 *
 * @example
 * ```typescript
 * const queryFn = createETagQueryFn(
 *   () => authenticatedFetch('/api/data'),
 *   'data-cache-key',
 *   () => queryClient.getQueryData(['data'])
 * );
 * ```
 */
export function createETagQueryFn<T>(
  fetchFn: () => Promise<Response>,
  cacheKey: string,
  getPreviousData: () => T | undefined
): () => Promise<T> {
  return async () => {
    try {
      const response = await fetchFn();

      // Store ETag if present
      const etag = response.headers.get("ETag");
      if (etag) {
        storeETag(cacheKey, etag);
      }

      return response.json() as Promise<T>;
    } catch (error) {
      if (isNotModifiedError(error)) {
        const previousData = getPreviousData();
        if (previousData !== undefined) {
          return previousData;
        }
      }
      throw error;
    }
  };
}
