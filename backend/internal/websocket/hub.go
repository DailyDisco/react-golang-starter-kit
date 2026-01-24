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
	MessageTypeNotification       MessageType = "notification"
	MessageTypeUserUpdate         MessageType = "user_update"
	MessageTypeBroadcast          MessageType = "broadcast"
	MessageTypePing               MessageType = "ping"
	MessageTypePong               MessageType = "pong"
	MessageTypeCacheInvalidate    MessageType = "cache_invalidate"
	MessageTypeUsageAlert         MessageType = "usage_alert"
	MessageTypeSubscriptionUpdate MessageType = "subscription_update"
	MessageTypeOrgUpdate          MessageType = "org_update"
	MessageTypeMemberUpdate       MessageType = "member_update"
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

	// CanUpgrade indicates if the user can upgrade to a higher plan
	CanUpgrade bool `json:"canUpgrade"`

	// CurrentPlan is the user's current subscription plan
	CurrentPlan string `json:"currentPlan,omitempty"`

	// SuggestedPlan is the recommended plan to upgrade to
	SuggestedPlan string `json:"suggestedPlan,omitempty"`

	// UpgradeURL is the URL to the upgrade/billing page
	UpgradeURL string `json:"upgradeUrl,omitempty"`
}

// SubscriptionUpdatePayload is sent to clients when subscription status changes
type SubscriptionUpdatePayload struct {
	// Event is the type of subscription event (created, updated, deleted, payment_failed)
	Event string `json:"event"`

	// Status is the new subscription status
	Status string `json:"status"`

	// Plan is the subscription plan name
	Plan string `json:"plan,omitempty"`

	// PriceID is the Stripe price ID
	PriceID string `json:"priceId,omitempty"`

	// CancelAtPeriodEnd indicates if subscription will cancel at period end
	CancelAtPeriodEnd bool `json:"cancelAtPeriodEnd"`

	// CurrentPeriodEnd is when the current billing period ends
	CurrentPeriodEnd string `json:"currentPeriodEnd,omitempty"`

	// Message is a human-readable description of the change
	Message string `json:"message"`

	// Timestamp is when this event occurred
	Timestamp int64 `json:"timestamp"`
}

// OrgUpdatePayload is sent to clients when organization settings change
type OrgUpdatePayload struct {
	// OrgSlug is the organization's slug identifier
	OrgSlug string `json:"orgSlug"`

	// Event is the type of organization event (settings_changed, billing_changed, deleted)
	Event string `json:"event"`

	// Field is the specific field that changed (optional)
	Field string `json:"field,omitempty"`
}

// MemberUpdatePayload is sent to clients when organization membership changes
type MemberUpdatePayload struct {
	// OrgSlug is the organization's slug identifier
	OrgSlug string `json:"orgSlug"`

	// Event is the type of membership event (added, removed, role_changed, invitation_sent, invitation_revoked)
	Event string `json:"event"`

	// UserID is the affected user's ID (optional)
	UserID uint `json:"userId,omitempty"`

	// Role is the new role (for role_changed events)
	Role string `json:"role,omitempty"`
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

	// Organization to user IDs mapping for org-scoped broadcasts
	orgClients map[uint]map[uint]struct{}

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
		orgClients: make(map[uint]map[uint]struct{}),
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

// SendToUsers sends a message to multiple specific users
func (h *Hub) SendToUsers(userIDs []uint, msgType MessageType, payload interface{}) {
	for _, userID := range userIDs {
		h.SendToUser(userID, msgType, payload)
	}
}

// SetUserOrgs updates the organization memberships for a connected user.
// This should be called when a user connects or when their org memberships change.
func (h *Hub) SetUserOrgs(userID uint, orgIDs []uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, ok := h.clients[userID]
	if !ok {
		return
	}

	// Remove user from old org mappings
	for _, oldOrgID := range client.OrgIDs {
		if orgUsers, exists := h.orgClients[oldOrgID]; exists {
			delete(orgUsers, userID)
			if len(orgUsers) == 0 {
				delete(h.orgClients, oldOrgID)
			}
		}
	}

	// Update client's org IDs
	client.OrgIDs = orgIDs

	// Add user to new org mappings
	for _, orgID := range orgIDs {
		if h.orgClients[orgID] == nil {
			h.orgClients[orgID] = make(map[uint]struct{})
		}
		h.orgClients[orgID][userID] = struct{}{}
	}

	log.Debug().
		Uint("user_id", userID).
		Uints("org_ids", orgIDs).
		Msg("Updated user org memberships")
}

// BroadcastToOrg sends a message to all users in a specific organization
func (h *Hub) BroadcastToOrg(orgID uint, msgType MessageType, payload interface{}) {
	h.mu.RLock()
	userIDs, exists := h.orgClients[orgID]
	if !exists || len(userIDs) == 0 {
		h.mu.RUnlock()
		return
	}

	// Collect user IDs while holding read lock
	targets := make([]uint, 0, len(userIDs))
	for userID := range userIDs {
		targets = append(targets, userID)
	}
	h.mu.RUnlock()

	// Send to all org members
	msg := Message{
		Type:    msgType,
		Payload: payload,
	}

	for _, userID := range targets {
		h.mu.RLock()
		client, ok := h.clients[userID]
		h.mu.RUnlock()

		if ok {
			select {
			case client.send <- msg:
			default:
				log.Warn().
					Uint("user_id", userID).
					Uint("org_id", orgID).
					Msg("Client send buffer full, org message dropped")
			}
		}
	}

	log.Debug().
		Uint("org_id", orgID).
		Int("recipients", len(targets)).
		Str("type", string(msgType)).
		Msg("Broadcast message to organization")
}

// GetOrgUserCount returns the number of connected users in an organization
func (h *Hub) GetOrgUserCount(orgID uint) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if orgUsers, exists := h.orgClients[orgID]; exists {
		return len(orgUsers)
	}
	return 0
}

// GetConnectedOrgIDs returns a list of all organization IDs with connected users
func (h *Hub) GetConnectedOrgIDs() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()

	orgIDs := make([]uint, 0, len(h.orgClients))
	for orgID := range h.orgClients {
		orgIDs = append(orgIDs, orgID)
	}
	return orgIDs
}
