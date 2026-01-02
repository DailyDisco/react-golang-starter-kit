// Package websocket provides real-time communication capabilities via WebSocket connections.
package websocket

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Message types for different events
	MessageTypeNotification    MessageType = "notification"
	MessageTypeUserUpdate      MessageType = "user_update"
	MessageTypeBroadcast       MessageType = "broadcast"
	MessageTypePing            MessageType = "ping"
	MessageTypePong            MessageType = "pong"
	MessageTypeCacheInvalidate MessageType = "cache_invalidate"
	MessageTypeUsageAlert      MessageType = "usage_alert"
)

// CacheInvalidatePayload is sent to clients when server-side cache is invalidated.
// Clients should invalidate their corresponding TanStack Query cache entries.
type CacheInvalidatePayload struct {
	// QueryKeys are the TanStack Query keys to invalidate (e.g., ["featureFlags"], ["settings"])
	QueryKeys []string `json:"queryKeys"`

	// Event is the cache event type (e.g., "feature_flags:updated")
	Event string `json:"event,omitempty"`

	// Timestamp is the Unix timestamp when invalidation occurred
	Timestamp int64 `json:"timestamp"`
}

// UsageAlertPayload is sent to clients when usage approaches or exceeds limits
type UsageAlertPayload struct {
	// AlertType is the type of alert (warning_80, warning_90, exceeded)
	AlertType string `json:"alertType"`

	// UsageType is the type of usage (api_calls, storage, etc.)
	UsageType string `json:"usageType"`

	// CurrentUsage is the current usage amount
	CurrentUsage int64 `json:"currentUsage"`

	// Limit is the maximum allowed usage
	Limit int64 `json:"limit"`

	// PercentageUsed is the percentage of the limit used
	PercentageUsed int `json:"percentageUsed"`

	// Message is a human-readable alert message
	Message string `json:"message"`
}

// Message represents a WebSocket message
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	UserID  uint        `json:"-"` // Target user ID (0 = broadcast to all)
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients by user ID
	clients map[uint]*Client

	// Channel for messages to broadcast
	broadcast chan Message

	// Channel for registering clients
	register chan *Client

	// Channel for unregistering clients
	unregister chan *Client

	// Mutex for thread-safe client map access
	mu sync.RWMutex

	// Done channel for graceful shutdown
	done chan struct{}
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint]*Client),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run(ctx context.Context) {
	log.Info().Msg("WebSocket hub started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("WebSocket hub shutting down")
			h.closeAll()
			return

		case <-h.done:
			log.Info().Msg("WebSocket hub received shutdown signal")
			h.closeAll()
			return

		case client := <-h.register:
			h.mu.Lock()
			// Close existing connection for the same user if any
			if existing, ok := h.clients[client.UserID]; ok {
				close(existing.send)
				log.Debug().Uint("user_id", client.UserID).Msg("Replaced existing WebSocket connection")
			}
			h.clients[client.UserID] = client
			h.mu.Unlock()
			log.Debug().Uint("user_id", client.UserID).Int("total_clients", len(h.clients)).Msg("Client registered")

		case client := <-h.unregister:
			h.mu.Lock()
			if existing, ok := h.clients[client.UserID]; ok && existing == client {
				delete(h.clients, client.UserID)
				close(client.send)
				log.Debug().Uint("user_id", client.UserID).Int("total_clients", len(h.clients)).Msg("Client unregistered")
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.sendMessage(msg)
		}
	}
}

// Stop gracefully shuts down the hub
func (h *Hub) Stop() {
	close(h.done)
}

// closeAll closes all client connections
func (h *Hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for userID, client := range h.clients {
		close(client.send)
		delete(h.clients, userID)
	}
	log.Info().Msg("All WebSocket connections closed")
}

// sendMessage sends a message to the appropriate client(s)
func (h *Hub) sendMessage(msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if msg.UserID == 0 {
		// Broadcast to all clients
		for _, client := range h.clients {
			select {
			case client.send <- msg:
			default:
				// Client's send buffer is full, skip
				log.Warn().Uint("user_id", client.UserID).Msg("Client send buffer full, message dropped")
			}
		}
	} else {
		// Send to specific user
		if client, ok := h.clients[msg.UserID]; ok {
			select {
			case client.send <- msg:
			default:
				log.Warn().Uint("user_id", msg.UserID).Msg("Client send buffer full, message dropped")
			}
		}
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID uint, msgType MessageType, payload interface{}) {
	h.broadcast <- Message{
		Type:    msgType,
		Payload: payload,
		UserID:  userID,
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(msgType MessageType, payload interface{}) {
	h.broadcast <- Message{
		Type:    msgType,
		Payload: payload,
		UserID:  0,
	}
}

// IsUserConnected checks if a user has an active WebSocket connection
func (h *Hub) IsUserConnected(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// GetConnectedUserCount returns the number of connected users
func (h *Hub) GetConnectedUserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetConnectedUserIDs returns a list of all connected user IDs
func (h *Hub) GetConnectedUserIDs() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userIDs := make([]uint, 0, len(h.clients))
	for userID := range h.clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}
