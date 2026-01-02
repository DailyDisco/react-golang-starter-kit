/**
 * Cross-tab authentication synchronization using BroadcastChannel API
 *
 * This module enables synchronization of authentication state across browser tabs.
 * When a user logs out in one tab, all other tabs are notified and also log out.
 */

const AUTH_CHANNEL_NAME = "auth_sync";

// Auth event types that can be broadcast across tabs
export type AuthEventType = "logout" | "login" | "session_expired";

export interface AuthEvent {
  type: AuthEventType;
  userId?: number;
  timestamp: number;
}

// Singleton BroadcastChannel instance
let authChannel: BroadcastChannel | null = null;

// Subscription callback type
type AuthEventCallback = (event: AuthEvent) => void;

// Active subscriptions
const subscribers = new Set<AuthEventCallback>();

/**
 * Initialize the auth channel if BroadcastChannel is supported
 */
function getChannel(): BroadcastChannel | null {
  if (typeof window === "undefined" || typeof BroadcastChannel === "undefined") {
    return null;
  }

  if (!authChannel) {
    try {
      authChannel = new BroadcastChannel(AUTH_CHANNEL_NAME);
      authChannel.addEventListener("message", handleMessage);
    } catch {
      // BroadcastChannel not supported or failed to create
      return null;
    }
  }

  return authChannel;
}

/**
 * Handle incoming messages from other tabs
 */
function handleMessage(event: MessageEvent<AuthEvent>): void {
  const authEvent = event.data;

  // Validate the event structure
  if (!authEvent || typeof authEvent.type !== "string" || typeof authEvent.timestamp !== "number") {
    return;
  }

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
 * Broadcast an auth event to all other tabs
 */
export function broadcastAuthEvent(type: AuthEventType, userId?: number): void {
  const channel = getChannel();
  if (!channel) return;

  const event: AuthEvent = {
    type,
    userId,
    timestamp: Date.now(),
  };

  try {
    channel.postMessage(event);
  } catch {
    // Failed to broadcast - channel might be closed
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
  if (authChannel) {
    authChannel.removeEventListener("message", handleMessage);
    authChannel.close();
    authChannel = null;
  }
  subscribers.clear();
}
