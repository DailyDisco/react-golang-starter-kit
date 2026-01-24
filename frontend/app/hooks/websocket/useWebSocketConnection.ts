import { useCallback, useEffect, useRef, useState } from "react";

import { logger } from "../../lib/logger";
import { AuthService } from "../../services/auth/authService";
import { useAuthStore } from "../../stores/auth-store";
import type { MessageType, WebSocketMessage } from "./types";

interface UseWebSocketConnectionOptions {
  /** Reconnection interval in ms (default: 3000) */
  reconnectInterval?: number;
  /** Maximum reconnection attempts (default: 5) */
  maxRetries?: number;
  /** Enable auto-connect when authenticated (default: true) */
  autoConnect?: boolean;
  /** Handler called when a message is received */
  onMessage?: (message: WebSocketMessage) => void;
}

interface UseWebSocketConnectionReturn {
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
 * Get WebSocket URL based on current environment
 */
function getWebSocketUrl(): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
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
    host = window.location.host;
  }

  return `${protocol}//${host}/ws`;
}

/**
 * Hook for managing WebSocket connection lifecycle with automatic reconnection.
 * Handles connection, disconnection, ping/pong, and reconnection logic.
 */
export function useWebSocketConnection(options: UseWebSocketConnectionOptions = {}): UseWebSocketConnectionReturn {
  const { reconnectInterval = 3000, maxRetries = 5, autoConnect = true, onMessage } = options;

  const wsRef = useRef<WebSocket | null>(null);
  const retriesRef = useRef(0);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const connectRef = useRef<(() => void) | null>(null);

  const [isConnected, setIsConnected] = useState(false);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  /**
   * Connect to WebSocket server
   */
  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN || !isAuthenticated) {
      return;
    }

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
          onMessage?.(message);
        } catch (err) {
          logger.error("Failed to parse WebSocket message", err);
        }
      };

      ws.onclose = (event) => {
        setIsConnected(false);
        wsRef.current = null;

        if (isAuthenticated && retriesRef.current < maxRetries) {
          retriesRef.current++;
          const delay = reconnectInterval * Math.min(retriesRef.current, 5);

          logger.debug("WebSocket closed, reconnecting...", {
            attempt: retriesRef.current,
            delay,
            code: event.code,
          });

          reconnectTimeoutRef.current = setTimeout(async () => {
            try {
              const isValid = await AuthService.validateSession();
              if (isValid) {
                connectRef.current?.();
              } else {
                logger.warn("Session invalid during WebSocket reconnect, skipping reconnection");
                retriesRef.current = maxRetries;
              }
            } catch (err) {
              logger.error("Failed to validate session during WebSocket reconnect", err);
              connectRef.current?.();
            }
          }, delay);
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
  }, [isAuthenticated, onMessage, reconnectInterval, maxRetries]);

  // Keep connectRef updated so reconnect timeout can call the latest version
  useEffect(() => {
    connectRef.current = connect;
  }, [connect]);

  /**
   * Disconnect from WebSocket server
   */
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    retriesRef.current = maxRetries;

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
    }, 30000);

    return () => clearInterval(pingInterval);
  }, [isConnected, sendMessage]);

  return {
    isConnected,
    connect,
    disconnect,
    sendMessage,
  };
}
