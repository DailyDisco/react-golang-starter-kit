/**
 * Cross-tab authentication synchronization using BroadcastChannel API
 *
 * This module enables synchronization of authentication state across browser tabs.
 * When a user logs out in one tab, all other tabs are notified and also log out.
 *
 * Falls back to localStorage events for browsers that don't support BroadcastChannel
 * (e.g., older browsers, some private browsing modes).
 */

const AUTH_CHANNEL_NAME = "auth_sync";
const AUTH_STORAGE_KEY = "auth_sync_event";

// Auth event types that can be broadcast across tabs
export type AuthEventType = "logout" | "login" | "session_expired";

export interface AuthEvent {
  type: AuthEventType;
  userId?: number;
  timestamp: number;
}

// Singleton BroadcastChannel instance
let authChannel: BroadcastChannel | null = null;

// Track whether we're using fallback mode
let usingFallback = false;

// Track last processed event timestamp to prevent duplicate/stale events
let lastEventTimestamp = 0;

// Subscription callback type
type AuthEventCallback = (event: AuthEvent) => void;

// Active subscriptions
const subscribers = new Set<AuthEventCallback>();

// Storage event handler reference for cleanup
let storageHandler: ((e: StorageEvent) => void) | null = null;

/**
 * Initialize localStorage fallback for cross-tab sync
 * Used when BroadcastChannel is not available
 */
function initStorageFallback(): void {
  if (typeof window === "undefined" || storageHandler) {
    return;
  }

  storageHandler = (e: StorageEvent) => {
    if (e.key !== AUTH_STORAGE_KEY || !e.newValue) {
      return;
    }

    try {
      const event = JSON.parse(e.newValue) as AuthEvent;
      notifySubscribers(event);
    } catch {
      // Invalid JSON in storage event
    }
  };

  window.addEventListener("storage", storageHandler);
  usingFallback = true;
}

/**
 * Initialize the auth channel if BroadcastChannel is supported
 * Falls back to localStorage events if not available
 */
function getChannel(): BroadcastChannel | null {
  if (typeof window === "undefined") {
    return null;
  }

  // If already using fallback, return null
  if (usingFallback) {
    return null;
  }

  // BroadcastChannel not available - use fallback
  if (typeof BroadcastChannel === "undefined") {
    initStorageFallback();
    return null;
  }

  if (!authChannel) {
    try {
      authChannel = new BroadcastChannel(AUTH_CHANNEL_NAME);
      authChannel.addEventListener("message", handleMessage);
    } catch {
      // BroadcastChannel failed - use fallback
      initStorageFallback();
      return null;
    }
  }

  return authChannel;
}

/**
 * Notify all subscribers of an auth event
 * Includes timestamp validation to prevent processing stale events
 */
function notifySubscribers(authEvent: AuthEvent): void {
  // Validate the event structure
  if (!authEvent || typeof authEvent.type !== "string" || typeof authEvent.timestamp !== "number") {
    return;
  }

  // Ignore stale events (already processed or out of order)
  if (authEvent.timestamp <= lastEventTimestamp) {
    return;
  }
  lastEventTimestamp = authEvent.timestamp;

  // Notify all subscribers
  for (const callback of subscribers) {
    try {
      callback(authEvent);
    } catch {
      // Ignore errors in individual subscribers
    }
  }
}

/**
 * Handle incoming messages from other tabs (BroadcastChannel)
 */
function handleMessage(event: MessageEvent<AuthEvent>): void {
  notifySubscribers(event.data);
}

/**
 * Broadcast an auth event to all other tabs
 * Uses BroadcastChannel if available, otherwise localStorage
 */
export function broadcastAuthEvent(type: AuthEventType, userId?: number): void {
  if (typeof window === "undefined") {
    return;
  }

  const event: AuthEvent = {
    type,
    userId,
    timestamp: Date.now(),
  };

  const channel = getChannel();

  if (channel) {
    // Use BroadcastChannel
    try {
      channel.postMessage(event);
    } catch {
      // Failed to broadcast - channel might be closed
    }
  } else if (usingFallback) {
    // Use localStorage fallback
    try {
      // Write to localStorage - other tabs will receive via 'storage' event
      localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(event));
      // Clear immediately to allow same event type to be sent again
      localStorage.removeItem(AUTH_STORAGE_KEY);
    } catch {
      // localStorage might be full or disabled
    }
  }
}

/**
 * Subscribe to auth events from other tabs
 * Returns an unsubscribe function
 */
export function subscribeToAuthEvents(callback: AuthEventCallback): () => void {
  // Ensure channel is initialized
  getChannel();

  subscribers.add(callback);

  return () => {
    subscribers.delete(callback);
  };
}

/**
 * Broadcast a logout event to all tabs
 */
export function broadcastLogout(): void {
  broadcastAuthEvent("logout");
}

/**
 * Broadcast a login event to all tabs
 */
export function broadcastLogin(userId: number): void {
  broadcastAuthEvent("login", userId);
}

/**
 * Broadcast a session expired event to all tabs
 */
export function broadcastSessionExpired(): void {
  broadcastAuthEvent("session_expired");
}

/**
 * Close the auth channel (cleanup)
 */
export function closeAuthChannel(): void {
  // Clean up BroadcastChannel
  if (authChannel) {
    authChannel.removeEventListener("message", handleMessage);
    authChannel.close();
    authChannel = null;
  }

  // Clean up localStorage fallback
  if (storageHandler && typeof window !== "undefined") {
    window.removeEventListener("storage", storageHandler);
    storageHandler = null;
  }

  // Reset state
  subscribers.clear();
  usingFallback = false;
  lastEventTimestamp = 0;
}

/**
 * Check if the auth channel is available and healthy
 * Returns true if BroadcastChannel or fallback is active
 */
export function isChannelHealthy(): boolean {
  if (typeof window === "undefined") {
    return false;
  }

  // If we have an active BroadcastChannel, it's healthy
  if (authChannel) {
    return true;
  }

  // If we're using fallback, it's healthy
  if (usingFallback) {
    return true;
  }

  // Try to initialize and check
  getChannel();
  return authChannel !== null || usingFallback;
}

/**
 * Check if using localStorage fallback instead of BroadcastChannel
 */
export function isUsingFallback(): boolean {
  return usingFallback;
}
