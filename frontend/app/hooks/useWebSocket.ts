import { useCallback, useEffect, useRef, useState } from "react";

import { useQueryClient } from "@tanstack/react-query";

import { logger } from "../lib/logger";
import { useAuthStore } from "../stores/auth-store";
import { useNotificationStore } from "../stores/notification-store";

// WebSocket message types (must match backend)
type MessageType = "notification" | "user_update" | "broadcast" | "ping" | "pong";

interface WebSocketMessage {
  type: MessageType;
  payload?: unknown;
}

interface NotificationPayload {
  id: string;
  title: string;
  message: string;
  type: "info" | "success" | "warning" | "error";
  timestamp: string;
  data?: unknown;
}

interface UserUpdatePayload {
  field: string;
  value?: unknown;
}

interface UseWebSocketOptions {
  /** Custom message handler */
  onMessage?: (message: WebSocketMessage) => void;
  /** Reconnection interval in ms (default: 3000) */
  reconnectInterval?: number;
  /** Maximum reconnection attempts (default: 5) */
  maxRetries?: number;
  /** Enable auto-connect when authenticated (default: true) */
  autoConnect?: boolean;
}

interface UseWebSocketReturn {
  /** Whether the WebSocket is connected */
  isConnected: boolean;
  /** Manually connect to WebSocket */
  connect: () => void;
  /** Manually disconnect from WebSocket */
  disconnect: () => void;
  /** Send a message through WebSocket */
  sendMessage: (type: MessageType, payload?: unknown) => void;
}

/**
 * Hook for managing WebSocket connection with automatic reconnection
 * and integration with TanStack Query for cache invalidation
 */
export function useWebSocket(options: UseWebSocketOptions = {}): UseWebSocketReturn {
  const { reconnectInterval = 3000, maxRetries = 5, autoConnect = true, onMessage } = options;

  const wsRef = useRef<WebSocket | null>(null);
  const retriesRef = useRef(0);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const connectRef = useRef<(() => void) | null>(null);

  const [isConnected, setIsConnected] = useState(false);

  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const queryClient = useQueryClient();
  const addNotification = useNotificationStore((state) => state.addNotification);

  /**
   * Get WebSocket URL based on current environment
   */
  const getWebSocketUrl = useCallback(() => {
    // Determine protocol
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";

    // Use API URL if configured, otherwise use current host
    const apiUrl = import.meta.env.VITE_API_URL;
    let host: string;

    if (apiUrl) {
      try {
        const url = new URL(apiUrl);
        host = url.host;
      } catch {
        host = window.location.host;
      }
    } else {
      // In development with Vite proxy, WebSocket needs to go to backend directly
      host = window.location.host;
    }

    return `${protocol}//${host}/ws`;
  }, []);

  /**
   * Handle incoming WebSocket messages
   */
  const handleMessage = useCallback(
    (message: WebSocketMessage) => {
      switch (message.type) {
        case "notification": {
          const payload = message.payload as NotificationPayload;
          addNotification({
            title: payload.title,
            message: payload.message,
            type: payload.type,
            data: payload.data,
          });
          break;
        }

        case "user_update": {
          const payload = message.payload as UserUpdatePayload;
          // Invalidate relevant queries based on what was updated
          if (payload.field === "profile" || payload.field === "role") {
            queryClient.invalidateQueries({ queryKey: ["auth", "user"] });
          } else if (payload.field === "preferences") {
            queryClient.invalidateQueries({ queryKey: ["preferences"] });
          } else if (payload.field === "sessions") {
            queryClient.invalidateQueries({ queryKey: ["sessions"] });
          }
          break;
        }

        case "broadcast": {
          // Handle broadcast messages (e.g., system announcements)
          const payload = message.payload as NotificationPayload;
          if (payload) {
            addNotification({
              title: payload.title || "System",
              message: payload.message,
              type: payload.type || "info",
              data: payload.data,
            });
          }
          break;
        }

        case "pong":
          // Server responded to our ping
          break;

        default:
          logger.debug("Unknown WebSocket message type", { type: message.type });
      }

      // Call custom handler if provided
      onMessage?.(message);
    },
    [addNotification, queryClient, onMessage]
  );

  /**
   * Connect to WebSocket server
   */
  const connect = useCallback(() => {
    // Don't connect if already connected or not authenticated
    if (wsRef.current?.readyState === WebSocket.OPEN || !isAuthenticated) {
      return;
    }

    // Clear any pending reconnect
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    const url = getWebSocketUrl();
    logger.debug("Connecting to WebSocket", { url });

    try {
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        retriesRef.current = 0;
        logger.info("WebSocket connected");
      };

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as WebSocketMessage;
          handleMessage(message);
        } catch (err) {
          logger.error("Failed to parse WebSocket message", err);
        }
      };

      ws.onclose = (event) => {
        setIsConnected(false);
        wsRef.current = null;

        // Only attempt reconnect if we should be connected
        if (isAuthenticated && retriesRef.current < maxRetries) {
          retriesRef.current++;
          const delay = reconnectInterval * Math.min(retriesRef.current, 5); // Cap exponential backoff

          logger.debug("WebSocket closed, reconnecting...", {
            attempt: retriesRef.current,
            delay,
            code: event.code,
          });

          reconnectTimeoutRef.current = setTimeout(() => connectRef.current?.(), delay);
        } else if (retriesRef.current >= maxRetries) {
          logger.warn("WebSocket max reconnection attempts reached");
        }
      };

      ws.onerror = (error) => {
        logger.error("WebSocket error", error);
      };
    } catch (err) {
      logger.error("Failed to create WebSocket connection", err);
    }
  }, [isAuthenticated, getWebSocketUrl, handleMessage, reconnectInterval, maxRetries]);

  // Keep connectRef updated so reconnect timeout can call the latest version
  useEffect(() => {
    connectRef.current = connect;
  }, [connect]);

  /**
   * Disconnect from WebSocket server
   */
  const disconnect = useCallback(() => {
    // Clear any pending reconnect
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    // Reset retry counter
    retriesRef.current = maxRetries; // Prevent auto-reconnect

    if (wsRef.current) {
      wsRef.current.close(1000, "Client disconnecting");
      wsRef.current = null;
    }

    setIsConnected(false);
  }, [maxRetries]);

  /**
   * Send a message through WebSocket
   */
  const sendMessage = useCallback((type: MessageType, payload?: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, payload }));
    } else {
      logger.warn("Cannot send message, WebSocket not connected");
    }
  }, []);

  // Auto-connect when authenticated
  useEffect(() => {
    if (autoConnect && isAuthenticated) {
      connect();
    } else if (!isAuthenticated) {
      disconnect();
    }

    return () => {
      disconnect();
    };
  }, [autoConnect, isAuthenticated, connect, disconnect]);

  // Send periodic pings to keep connection alive
  useEffect(() => {
    if (!isConnected) return;

    const pingInterval = setInterval(() => {
      sendMessage("ping");
    }, 30000); // Every 30 seconds

    return () => clearInterval(pingInterval);
  }, [isConnected, sendMessage]);

  return {
    isConnected,
    connect,
    disconnect,
    sendMessage,
  };
}
