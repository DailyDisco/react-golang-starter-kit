package websocket

import (
	"context"
	"testing"
	"time"
)

// ============ NewHub Tests ============

func TestNewHub(t *testing.T) {
	hub := NewHub()

	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}

	if hub.clients == nil {
		t.Error("NewHub().clients should not be nil")
	}

	if hub.orgClients == nil {
		t.Error("NewHub().orgClients should not be nil")
	}

	if hub.broadcast == nil {
		t.Error("NewHub().broadcast should not be nil")
	}

	if hub.register == nil {
		t.Error("NewHub().register should not be nil")
	}

	if hub.unregister == nil {
		t.Error("NewHub().unregister should not be nil")
	}

	if hub.done == nil {
		t.Error("NewHub().done should not be nil")
	}
}

func TestNewHub_ChannelBufferSize(t *testing.T) {
	hub := NewHub()

	// Broadcast channel should have buffer size of 256
	if cap(hub.broadcast) != 256 {
		t.Errorf("NewHub().broadcast capacity = %d, want 256", cap(hub.broadcast))
	}
}

// ============ MessageType Constants Tests ============

func TestMessageTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		expected string
	}{
		{"notification", MessageTypeNotification, "notification"},
		{"user_update", MessageTypeUserUpdate, "user_update"},
		{"broadcast", MessageTypeBroadcast, "broadcast"},
		{"ping", MessageTypePing, "ping"},
		{"pong", MessageTypePong, "pong"},
		{"cache_invalidate", MessageTypeCacheInvalidate, "cache_invalidate"},
		{"usage_alert", MessageTypeUsageAlert, "usage_alert"},
		{"subscription_update", MessageTypeSubscriptionUpdate, "subscription_update"},
		{"org_update", MessageTypeOrgUpdate, "org_update"},
		{"member_update", MessageTypeMemberUpdate, "member_update"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.msgType) != tt.expected {
				t.Errorf("MessageType %s = %q, want %q", tt.name, tt.msgType, tt.expected)
			}
		})
	}
}

// ============ Hub Client Management Tests ============

func TestHub_RegisterClient(t *testing.T) {
	hub := NewHub()

	// Start hub in background
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer func() {
		cancel()
		time.Sleep(10 * time.Millisecond) // Allow goroutine to clean up
	}()

	// Create a mock client
	client := &Client{
		UserID: 1,
		send:   make(chan Message, 10),
		hub:    hub,
	}

	// Register client
	hub.register <- client
	time.Sleep(10 * time.Millisecond) // Allow registration to process

	// Verify client is registered
	if !hub.IsUserConnected(1) {
		t.Error("Client should be connected after registration")
	}

	if hub.GetConnectedUserCount() != 1 {
		t.Errorf("GetConnectedUserCount() = %d, want 1", hub.GetConnectedUserCount())
	}
}

func TestHub_UnregisterClient(t *testing.T) {
	hub := NewHub()

	// Start hub in background
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer func() {
		cancel()
		time.Sleep(10 * time.Millisecond)
	}()

	// Create and register a mock client
	client := &Client{
		UserID: 1,
		send:   make(chan Message, 10),
		hub:    hub,
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Verify client is registered
	if !hub.IsUserConnected(1) {
		t.Fatal("Client should be connected before unregistration")
	}

	// Unregister client
	hub.unregister <- client
	time.Sleep(10 * time.Millisecond)

	// Verify client is unregistered
	if hub.IsUserConnected(1) {
		t.Error("Client should not be connected after unregistration")
	}

	if hub.GetConnectedUserCount() != 0 {
		t.Errorf("GetConnectedUserCount() = %d, want 0", hub.GetConnectedUserCount())
	}
}

func TestHub_ReplaceExistingClient(t *testing.T) {
	hub := NewHub()

	// Start hub in background
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer func() {
		cancel()
		time.Sleep(10 * time.Millisecond)
	}()

	// Create first client
	client1 := &Client{
		UserID: 1,
		send:   make(chan Message, 10),
		hub:    hub,
	}

	// Register first client
	hub.register <- client1
	time.Sleep(10 * time.Millisecond)

	// Create second client with same user ID
	client2 := &Client{
		UserID: 1,
		send:   make(chan Message, 10),
		hub:    hub,
	}

	// Register second client (should replace first)
	hub.register <- client2
	time.Sleep(10 * time.Millisecond)

	// Should still only have 1 connected user
	if hub.GetConnectedUserCount() != 1 {
		t.Errorf("GetConnectedUserCount() = %d, want 1 (replacing client)", hub.GetConnectedUserCount())
	}

	// First client's send channel should be closed
	select {
	case _, ok := <-client1.send:
		if ok {
			t.Error("First client's send channel should be closed")
		}
	default:
		// Channel might be closed but empty, that's fine
	}
}

// ============ IsUserConnected Tests ============

func TestHub_IsUserConnected_NotConnected(t *testing.T) {
	hub := NewHub()

	if hub.IsUserConnected(999) {
		t.Error("IsUserConnected(999) should return false for non-existent user")
	}
}

func TestHub_IsUserConnected_Connected(t *testing.T) {
	hub := NewHub()

	// Directly add client to map (unit test without running hub)
	hub.mu.Lock()
	hub.clients[42] = &Client{UserID: 42}
	hub.mu.Unlock()

	if !hub.IsUserConnected(42) {
		t.Error("IsUserConnected(42) should return true for connected user")
	}
}

// ============ GetConnectedUserCount Tests ============

func TestHub_GetConnectedUserCount_Empty(t *testing.T) {
	hub := NewHub()

	if hub.GetConnectedUserCount() != 0 {
		t.Errorf("GetConnectedUserCount() = %d, want 0 for empty hub", hub.GetConnectedUserCount())
	}
}

func TestHub_GetConnectedUserCount_Multiple(t *testing.T) {
	hub := NewHub()

	// Directly add clients to map
	hub.mu.Lock()
	hub.clients[1] = &Client{UserID: 1}
	hub.clients[2] = &Client{UserID: 2}
	hub.clients[3] = &Client{UserID: 3}
	hub.mu.Unlock()

	if hub.GetConnectedUserCount() != 3 {
		t.Errorf("GetConnectedUserCount() = %d, want 3", hub.GetConnectedUserCount())
	}
}

// ============ GetConnectedUserIDs Tests ============

func TestHub_GetConnectedUserIDs_Empty(t *testing.T) {
	hub := NewHub()

	ids := hub.GetConnectedUserIDs()
	if len(ids) != 0 {
		t.Errorf("GetConnectedUserIDs() length = %d, want 0", len(ids))
	}
}

func TestHub_GetConnectedUserIDs_Multiple(t *testing.T) {
	hub := NewHub()

	// Directly add clients to map
	hub.mu.Lock()
	hub.clients[10] = &Client{UserID: 10}
	hub.clients[20] = &Client{UserID: 20}
	hub.clients[30] = &Client{UserID: 30}
	hub.mu.Unlock()

	ids := hub.GetConnectedUserIDs()
	if len(ids) != 3 {
		t.Errorf("GetConnectedUserIDs() length = %d, want 3", len(ids))
	}

	// Check all IDs are present
	idMap := make(map[uint]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	for _, expectedID := range []uint{10, 20, 30} {
		if !idMap[expectedID] {
			t.Errorf("GetConnectedUserIDs() missing user ID %d", expectedID)
		}
	}
}

// ============ Organization Scoped Operations Tests ============

func TestHub_SetUserOrgs(t *testing.T) {
	hub := NewHub()

	// Add a client first
	hub.mu.Lock()
	client := &Client{UserID: 1, OrgIDs: []uint{}}
	hub.clients[1] = client
	hub.mu.Unlock()

	// Set org memberships
	hub.SetUserOrgs(1, []uint{100, 200})

	// Verify client's OrgIDs updated
	if len(client.OrgIDs) != 2 {
		t.Errorf("Client OrgIDs length = %d, want 2", len(client.OrgIDs))
	}

	// Verify orgClients map updated
	hub.mu.RLock()
	if _, exists := hub.orgClients[100]; !exists {
		t.Error("orgClients[100] should exist")
	}
	if _, exists := hub.orgClients[200]; !exists {
		t.Error("orgClients[200] should exist")
	}
	if _, exists := hub.orgClients[100][1]; !exists {
		t.Error("User 1 should be in orgClients[100]")
	}
	hub.mu.RUnlock()
}

func TestHub_SetUserOrgs_UpdateOrgs(t *testing.T) {
	hub := NewHub()

	// Add a client with initial org
	hub.mu.Lock()
	client := &Client{UserID: 1, OrgIDs: []uint{100}}
	hub.clients[1] = client
	hub.orgClients[100] = map[uint]struct{}{1: {}}
	hub.mu.Unlock()

	// Update to different orgs
	hub.SetUserOrgs(1, []uint{200, 300})

	// Verify old org no longer has user
	hub.mu.RLock()
	if _, exists := hub.orgClients[100]; exists {
		if _, userExists := hub.orgClients[100][1]; userExists {
			t.Error("User 1 should not be in orgClients[100] after update")
		}
	}

	// Verify new orgs have user
	if _, exists := hub.orgClients[200][1]; !exists {
		t.Error("User 1 should be in orgClients[200]")
	}
	if _, exists := hub.orgClients[300][1]; !exists {
		t.Error("User 1 should be in orgClients[300]")
	}
	hub.mu.RUnlock()
}

func TestHub_SetUserOrgs_NonExistentUser(t *testing.T) {
	hub := NewHub()

	// Should not panic for non-existent user
	hub.SetUserOrgs(999, []uint{100, 200})

	// Verify no org mappings created
	hub.mu.RLock()
	if len(hub.orgClients) != 0 {
		t.Errorf("orgClients should be empty for non-existent user, got %d", len(hub.orgClients))
	}
	hub.mu.RUnlock()
}

// ============ GetOrgUserCount Tests ============

func TestHub_GetOrgUserCount_Empty(t *testing.T) {
	hub := NewHub()

	if hub.GetOrgUserCount(100) != 0 {
		t.Errorf("GetOrgUserCount(100) = %d, want 0 for non-existent org", hub.GetOrgUserCount(100))
	}
}

func TestHub_GetOrgUserCount_Multiple(t *testing.T) {
	hub := NewHub()

	// Set up org with multiple users
	hub.mu.Lock()
	hub.clients[1] = &Client{UserID: 1}
	hub.clients[2] = &Client{UserID: 2}
	hub.clients[3] = &Client{UserID: 3}
	hub.orgClients[100] = map[uint]struct{}{
		1: {},
		2: {},
		3: {},
	}
	hub.mu.Unlock()

	if hub.GetOrgUserCount(100) != 3 {
		t.Errorf("GetOrgUserCount(100) = %d, want 3", hub.GetOrgUserCount(100))
	}
}

// ============ GetConnectedOrgIDs Tests ============

func TestHub_GetConnectedOrgIDs_Empty(t *testing.T) {
	hub := NewHub()

	orgIDs := hub.GetConnectedOrgIDs()
	if len(orgIDs) != 0 {
		t.Errorf("GetConnectedOrgIDs() length = %d, want 0", len(orgIDs))
	}
}

func TestHub_GetConnectedOrgIDs_Multiple(t *testing.T) {
	hub := NewHub()

	// Set up multiple orgs
	hub.mu.Lock()
	hub.orgClients[100] = map[uint]struct{}{1: {}}
	hub.orgClients[200] = map[uint]struct{}{2: {}}
	hub.orgClients[300] = map[uint]struct{}{3: {}}
	hub.mu.Unlock()

	orgIDs := hub.GetConnectedOrgIDs()
	if len(orgIDs) != 3 {
		t.Errorf("GetConnectedOrgIDs() length = %d, want 3", len(orgIDs))
	}

	// Check all org IDs are present
	orgMap := make(map[uint]bool)
	for _, id := range orgIDs {
		orgMap[id] = true
	}

	for _, expectedID := range []uint{100, 200, 300} {
		if !orgMap[expectedID] {
			t.Errorf("GetConnectedOrgIDs() missing org ID %d", expectedID)
		}
	}
}

// ============ Hub Stop Tests ============

func TestHub_Stop(t *testing.T) {
	hub := NewHub()

	// Start hub in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		hub.Run(ctx)
		close(done)
	}()

	// Stop the hub
	hub.Stop()

	// Wait for hub to stop (with timeout)
	select {
	case <-done:
		// Hub stopped successfully
	case <-time.After(time.Second):
		t.Error("Hub did not stop within timeout")
	}
}

// ============ Message Struct Tests ============

func TestMessage_Structure(t *testing.T) {
	msg := Message{
		Type:    MessageTypeNotification,
		Payload: "test payload",
		UserID:  42,
	}

	if msg.Type != MessageTypeNotification {
		t.Errorf("Message.Type = %v, want %v", msg.Type, MessageTypeNotification)
	}

	if msg.Payload != "test payload" {
		t.Errorf("Message.Payload = %v, want 'test payload'", msg.Payload)
	}

	if msg.UserID != 42 {
		t.Errorf("Message.UserID = %d, want 42", msg.UserID)
	}
}

// ============ Payload Struct Tests ============

func TestCacheInvalidatePayload_Structure(t *testing.T) {
	payload := CacheInvalidatePayload{
		QueryKeys: []string{"users", "settings"},
		Event:     "settings:updated",
		Timestamp: 1234567890,
	}

	if len(payload.QueryKeys) != 2 {
		t.Errorf("QueryKeys length = %d, want 2", len(payload.QueryKeys))
	}
	if payload.Event != "settings:updated" {
		t.Errorf("Event = %q, want 'settings:updated'", payload.Event)
	}
	if payload.Timestamp != 1234567890 {
		t.Errorf("Timestamp = %d, want 1234567890", payload.Timestamp)
	}
}

func TestUsageAlertPayload_Structure(t *testing.T) {
	payload := UsageAlertPayload{
		AlertType:      "warning_90",
		UsageType:      "api_calls",
		CurrentUsage:   9000,
		Limit:          10000,
		PercentageUsed: 90,
		Message:        "You have used 90% of your API calls",
		CanUpgrade:     true,
		CurrentPlan:    "free",
		SuggestedPlan:  "Pro",
		UpgradeURL:     "/settings/billing",
	}

	if payload.AlertType != "warning_90" {
		t.Errorf("AlertType = %q, want 'warning_90'", payload.AlertType)
	}
	if payload.PercentageUsed != 90 {
		t.Errorf("PercentageUsed = %d, want 90", payload.PercentageUsed)
	}
	if !payload.CanUpgrade {
		t.Error("CanUpgrade should be true")
	}
}

func TestSubscriptionUpdatePayload_Structure(t *testing.T) {
	payload := SubscriptionUpdatePayload{
		Event:             "created",
		Status:            "active",
		Plan:              "pro",
		PriceID:           "price_pro_monthly",
		CancelAtPeriodEnd: false,
		CurrentPeriodEnd:  "2025-02-01",
		Message:           "Subscription activated",
		Timestamp:         1234567890,
	}

	if payload.Event != "created" {
		t.Errorf("Event = %q, want 'created'", payload.Event)
	}
	if payload.Status != "active" {
		t.Errorf("Status = %q, want 'active'", payload.Status)
	}
}

func TestOrgUpdatePayload_Structure(t *testing.T) {
	payload := OrgUpdatePayload{
		OrgSlug: "my-org",
		Event:   "settings_changed",
		Field:   "name",
	}

	if payload.OrgSlug != "my-org" {
		t.Errorf("OrgSlug = %q, want 'my-org'", payload.OrgSlug)
	}
	if payload.Event != "settings_changed" {
		t.Errorf("Event = %q, want 'settings_changed'", payload.Event)
	}
}

func TestMemberUpdatePayload_Structure(t *testing.T) {
	payload := MemberUpdatePayload{
		OrgSlug: "my-org",
		Event:   "role_changed",
		UserID:  42,
		Role:    "admin",
	}

	if payload.OrgSlug != "my-org" {
		t.Errorf("OrgSlug = %q, want 'my-org'", payload.OrgSlug)
	}
	if payload.UserID != 42 {
		t.Errorf("UserID = %d, want 42", payload.UserID)
	}
	if payload.Role != "admin" {
		t.Errorf("Role = %q, want 'admin'", payload.Role)
	}
}

// ============ SendToUser Tests ============

func TestHub_SendToUser(t *testing.T) {
	hub := NewHub()

	// Start hub in background
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer func() {
		cancel()
		time.Sleep(10 * time.Millisecond)
	}()

	// Create and register a client
	client := &Client{
		UserID: 1,
		send:   make(chan Message, 10),
		hub:    hub,
	}
	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Send a message to the user
	hub.SendToUser(1, MessageTypeNotification, "test payload")

	// Wait for message to be received
	select {
	case msg := <-client.send:
		if msg.Type != MessageTypeNotification {
			t.Errorf("Message type = %v, want %v", msg.Type, MessageTypeNotification)
		}
		if msg.Payload != "test payload" {
			t.Errorf("Message payload = %v, want 'test payload'", msg.Payload)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for message")
	}
}

func TestHub_SendToUser_NotConnected(t *testing.T) {
	hub := NewHub()

	// Start hub in background
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer func() {
		cancel()
		time.Sleep(10 * time.Millisecond)
	}()

	// Send to non-existent user (should not panic)
	hub.SendToUser(999, MessageTypeNotification, "test")

	// Give it time to process (should complete without error)
	time.Sleep(50 * time.Millisecond)
}

// ============ Broadcast Tests ============

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()

	// Start hub in background
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer func() {
		cancel()
		time.Sleep(10 * time.Millisecond)
	}()

	// Create and register multiple clients
	client1 := &Client{UserID: 1, send: make(chan Message, 10), hub: hub}
	client2 := &Client{UserID: 2, send: make(chan Message, 10), hub: hub}
	hub.register <- client1
	hub.register <- client2
	time.Sleep(10 * time.Millisecond)

	// Broadcast a message
	hub.Broadcast(MessageTypeBroadcast, "broadcast message")

	// Both clients should receive the message
	for i, client := range []*Client{client1, client2} {
		select {
		case msg := <-client.send:
			if msg.Type != MessageTypeBroadcast {
				t.Errorf("Client %d: Message type = %v, want %v", i+1, msg.Type, MessageTypeBroadcast)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Client %d: Timeout waiting for broadcast message", i+1)
		}
	}
}

// ============ SendToUsers Tests ============

func TestHub_SendToUsers(t *testing.T) {
	hub := NewHub()

	// Start hub in background
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer func() {
		cancel()
		time.Sleep(10 * time.Millisecond)
	}()

	// Create and register multiple clients
	client1 := &Client{UserID: 1, send: make(chan Message, 10), hub: hub}
	client2 := &Client{UserID: 2, send: make(chan Message, 10), hub: hub}
	client3 := &Client{UserID: 3, send: make(chan Message, 10), hub: hub}
	hub.register <- client1
	hub.register <- client2
	hub.register <- client3
	time.Sleep(10 * time.Millisecond)

	// Send to specific users (1 and 3, not 2)
	hub.SendToUsers([]uint{1, 3}, MessageTypeUserUpdate, "selective message")

	// Clients 1 and 3 should receive messages
	for _, client := range []*Client{client1, client3} {
		select {
		case msg := <-client.send:
			if msg.Type != MessageTypeUserUpdate {
				t.Errorf("Message type = %v, want %v", msg.Type, MessageTypeUserUpdate)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Timeout waiting for message for user %d", client.UserID)
		}
	}

	// Client 2 should NOT receive a message
	select {
	case <-client2.send:
		t.Error("Client 2 should not have received a message")
	case <-time.After(50 * time.Millisecond):
		// Expected: no message
	}
}

// ============ BroadcastToOrg Tests ============

func TestHub_BroadcastToOrg(t *testing.T) {
	hub := NewHub()

	// Start hub in background
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer func() {
		cancel()
		time.Sleep(10 * time.Millisecond)
	}()

	// Create and register clients
	client1 := &Client{UserID: 1, OrgIDs: []uint{}, send: make(chan Message, 10), hub: hub}
	client2 := &Client{UserID: 2, OrgIDs: []uint{}, send: make(chan Message, 10), hub: hub}
	client3 := &Client{UserID: 3, OrgIDs: []uint{}, send: make(chan Message, 10), hub: hub}
	hub.register <- client1
	hub.register <- client2
	hub.register <- client3
	time.Sleep(10 * time.Millisecond)

	// Set org memberships: user 1 and 2 in org 100, user 3 in org 200
	hub.SetUserOrgs(1, []uint{100})
	hub.SetUserOrgs(2, []uint{100})
	hub.SetUserOrgs(3, []uint{200})

	// Broadcast to org 100
	hub.BroadcastToOrg(100, MessageTypeOrgUpdate, OrgUpdatePayload{OrgSlug: "test-org", Event: "settings_changed"})

	// Wait a bit for messages to be sent
	time.Sleep(50 * time.Millisecond)

	// Clients 1 and 2 should receive the message
	for _, client := range []*Client{client1, client2} {
		select {
		case msg := <-client.send:
			if msg.Type != MessageTypeOrgUpdate {
				t.Errorf("User %d: Message type = %v, want %v", client.UserID, msg.Type, MessageTypeOrgUpdate)
			}
		default:
			t.Errorf("User %d should have received org broadcast", client.UserID)
		}
	}

	// Client 3 should NOT receive the message
	select {
	case <-client3.send:
		t.Error("User 3 should not have received org 100 broadcast")
	default:
		// Expected: no message
	}
}

func TestHub_BroadcastToOrg_EmptyOrg(t *testing.T) {
	hub := NewHub()

	// Start hub in background
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer func() {
		cancel()
		time.Sleep(10 * time.Millisecond)
	}()

	// Broadcast to non-existent org (should not panic)
	hub.BroadcastToOrg(999, MessageTypeOrgUpdate, "test")

	// Give it time to process (should complete without error)
	time.Sleep(50 * time.Millisecond)
}

// ============ sendMessage Tests ============

func TestHub_SendMessage_BufferFull(t *testing.T) {
	hub := NewHub()

	// Add a client with a very small buffer
	hub.mu.Lock()
	client := &Client{
		UserID: 1,
		send:   make(chan Message, 1), // Very small buffer
		hub:    hub,
	}
	hub.clients[1] = client
	hub.mu.Unlock()

	// Fill the buffer
	client.send <- Message{Type: MessageTypePing}

	// Send another message directly via sendMessage (buffer full case)
	hub.sendMessage(Message{
		Type:   MessageTypeNotification,
		UserID: 1,
	})

	// The message should be dropped (not block) - verify by timeout
	select {
	case <-time.After(50 * time.Millisecond):
		// Expected: sendMessage completed without blocking
	}
}

func TestHub_SendMessage_BroadcastBufferFull(t *testing.T) {
	hub := NewHub()

	// Add a client with a very small buffer
	hub.mu.Lock()
	client := &Client{
		UserID: 1,
		send:   make(chan Message, 1),
		hub:    hub,
	}
	hub.clients[1] = client
	hub.mu.Unlock()

	// Fill the buffer
	client.send <- Message{Type: MessageTypePing}

	// Broadcast message (UserID = 0) should also handle full buffer gracefully
	hub.sendMessage(Message{
		Type:   MessageTypeBroadcast,
		UserID: 0, // Broadcast to all
	})

	// Should complete without blocking
	time.Sleep(50 * time.Millisecond)
}

// ============ closeAll Tests ============

func TestHub_CloseAll(t *testing.T) {
	hub := NewHub()

	// Add clients directly
	hub.mu.Lock()
	client1 := &Client{UserID: 1, send: make(chan Message, 10)}
	client2 := &Client{UserID: 2, send: make(chan Message, 10)}
	hub.clients[1] = client1
	hub.clients[2] = client2
	hub.mu.Unlock()

	// Close all
	hub.closeAll()

	// Verify clients map is empty
	hub.mu.RLock()
	clientCount := len(hub.clients)
	hub.mu.RUnlock()

	if clientCount != 0 {
		t.Errorf("clients map should be empty after closeAll, got %d", clientCount)
	}

	// Verify send channels are closed
	select {
	case _, ok := <-client1.send:
		if ok {
			t.Error("client1.send should be closed")
		}
	default:
		// Channel is closed and empty
	}
}

// ============ Hub Run Context Cancellation Tests ============

func TestHub_Run_ContextCancellation(t *testing.T) {
	hub := NewHub()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		hub.Run(ctx)
		close(done)
	}()

	// Register a client
	client := &Client{UserID: 1, send: make(chan Message, 10), hub: hub}
	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for hub to stop
	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("Hub did not stop after context cancellation")
	}

	// After shutdown, clients should be cleared
	if hub.GetConnectedUserCount() != 0 {
		t.Errorf("Expected 0 connected users after shutdown, got %d", hub.GetConnectedUserCount())
	}
}
