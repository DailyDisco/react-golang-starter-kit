package websocket

import (
	"testing"
)

// ============ Client Constants Tests ============

func TestClientConstants(t *testing.T) {
	// Verify constants are set to reasonable values
	tests := []struct {
		name     string
		value    interface{}
		minValue interface{}
		maxValue interface{}
	}{
		{"writeWait", writeWait.Seconds(), float64(5), float64(30)},
		{"pongWait", pongWait.Seconds(), float64(30), float64(120)},
		{"pingPeriod", pingPeriod.Seconds(), float64(30), float64(60)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := tt.value.(float64)
			min := tt.minValue.(float64)
			max := tt.maxValue.(float64)
			if val < min || val > max {
				t.Errorf("%s = %v, should be between %v and %v", tt.name, val, min, max)
			}
		})
	}
}

func TestPingPeriodLessThanPongWait(t *testing.T) {
	// pingPeriod must be less than pongWait for keep-alive to work correctly
	if pingPeriod >= pongWait {
		t.Errorf("pingPeriod (%v) should be less than pongWait (%v)", pingPeriod, pongWait)
	}
}

func TestMaxMessageSize(t *testing.T) {
	// maxMessageSize should be reasonable (not too small, not too large)
	minSize := int64(1024)     // 1KB minimum
	maxSize := int64(10485760) // 10MB maximum

	if maxMessageSize < minSize {
		t.Errorf("maxMessageSize = %d, should be at least %d", maxMessageSize, minSize)
	}

	if maxMessageSize > maxSize {
		t.Errorf("maxMessageSize = %d, should not exceed %d", maxMessageSize, maxSize)
	}
}

func TestSendBufferSize(t *testing.T) {
	// sendBufferSize should be reasonable
	if sendBufferSize < 16 {
		t.Errorf("sendBufferSize = %d, should be at least 16", sendBufferSize)
	}

	if sendBufferSize > 1024 {
		t.Errorf("sendBufferSize = %d, should not exceed 1024", sendBufferSize)
	}
}

// ============ NewClient Tests ============

func TestNewClient(t *testing.T) {
	hub := NewHub()

	client := NewClient(42, nil, hub)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.UserID != 42 {
		t.Errorf("NewClient().UserID = %d, want 42", client.UserID)
	}

	if client.hub != hub {
		t.Error("NewClient().hub should reference the provided hub")
	}

	if client.send == nil {
		t.Error("NewClient().send should not be nil")
	}

	// Verify send channel buffer size
	if cap(client.send) != sendBufferSize {
		t.Errorf("NewClient().send capacity = %d, want %d", cap(client.send), sendBufferSize)
	}
}

func TestNewClient_OrgIDsEmpty(t *testing.T) {
	hub := NewHub()

	client := NewClient(1, nil, hub)

	// OrgIDs should be nil/empty initially
	if len(client.OrgIDs) != 0 {
		t.Errorf("NewClient().OrgIDs length = %d, want 0", len(client.OrgIDs))
	}
}

// ============ IncomingMessage Tests ============

func TestIncomingMessage_Structure(t *testing.T) {
	msg := IncomingMessage{
		Type:    MessageTypePing,
		Payload: []byte(`{"test": "data"}`),
	}

	if msg.Type != MessageTypePing {
		t.Errorf("IncomingMessage.Type = %v, want %v", msg.Type, MessageTypePing)
	}

	if string(msg.Payload) != `{"test": "data"}` {
		t.Errorf("IncomingMessage.Payload = %s, want {\"test\": \"data\"}", msg.Payload)
	}
}

// ============ Client Structure Tests ============

func TestClient_Structure(t *testing.T) {
	hub := NewHub()

	client := &Client{
		UserID: 100,
		OrgIDs: []uint{1, 2, 3},
		conn:   nil,
		send:   make(chan Message, 10),
		hub:    hub,
	}

	if client.UserID != 100 {
		t.Errorf("Client.UserID = %d, want 100", client.UserID)
	}

	if len(client.OrgIDs) != 3 {
		t.Errorf("Client.OrgIDs length = %d, want 3", len(client.OrgIDs))
	}

	if cap(client.send) != 10 {
		t.Errorf("Client.send capacity = %d, want 10", cap(client.send))
	}
}
