package websocket

import (
	"context"
	"testing"
	"time"
)

// ============ NotificationPayload Tests ============

func TestNotificationPayload_Structure(t *testing.T) {
	now := time.Now().UTC()
	payload := NotificationPayload{
		ID:        "20250124120000.000000",
		Title:     "Test Title",
		Message:   "Test Message",
		Type:      "info",
		Timestamp: now,
		Data:      map[string]string{"key": "value"},
	}

	if payload.ID != "20250124120000.000000" {
		t.Errorf("ID = %q, want %q", payload.ID, "20250124120000.000000")
	}

	if payload.Title != "Test Title" {
		t.Errorf("Title = %q, want %q", payload.Title, "Test Title")
	}

	if payload.Message != "Test Message" {
		t.Errorf("Message = %q, want %q", payload.Message, "Test Message")
	}

	if payload.Type != "info" {
		t.Errorf("Type = %q, want %q", payload.Type, "info")
	}

	if !payload.Timestamp.Equal(now) {
		t.Errorf("Timestamp = %v, want %v", payload.Timestamp, now)
	}
}

func TestNotificationPayload_Types(t *testing.T) {
	types := []string{"info", "success", "warning", "error"}

	for _, typ := range types {
		t.Run(typ, func(t *testing.T) {
			payload := NotificationPayload{Type: typ}
			if payload.Type != typ {
				t.Errorf("Type = %q, want %q", payload.Type, typ)
			}
		})
	}
}

// ============ UserUpdatePayload Tests ============

func TestUserUpdatePayload_Structure(t *testing.T) {
	payload := UserUpdatePayload{
		Field: "email",
		Value: "new@example.com",
	}

	if payload.Field != "email" {
		t.Errorf("Field = %q, want %q", payload.Field, "email")
	}

	if payload.Value != "new@example.com" {
		t.Errorf("Value = %v, want %q", payload.Value, "new@example.com")
	}
}

func TestUserUpdatePayload_DifferentValueTypes(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value any
	}{
		{"string value", "name", "John Doe"},
		{"int value", "age", 30},
		{"bool value", "verified", true},
		{"nil value", "avatar", nil},
		{"map value", "preferences", map[string]bool{"dark_mode": true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := UserUpdatePayload{
				Field: tt.field,
				Value: tt.value,
			}
			if payload.Field != tt.field {
				t.Errorf("Field = %q, want %q", payload.Field, tt.field)
			}
		})
	}
}

// ============ NewNotificationService Tests ============

func TestNewNotificationService(t *testing.T) {
	hub := NewHub()
	service := NewNotificationService(hub)

	if service == nil {
		t.Fatal("NewNotificationService() returned nil")
	}

	if service.hub != hub {
		t.Error("NewNotificationService() did not set hub correctly")
	}
}

func TestNewNotificationService_NilHub(t *testing.T) {
	service := NewNotificationService(nil)

	if service == nil {
		t.Fatal("NewNotificationService() returned nil for nil hub")
	}

	if service.hub != nil {
		t.Error("service.hub should be nil")
	}
}

// ============ generateNotificationID Tests ============

func TestGenerateNotificationID_Format(t *testing.T) {
	id := generateNotificationID()

	if id == "" {
		t.Error("generateNotificationID() returned empty string")
	}

	// ID should be in format: 20060102150405.000000
	if len(id) != 21 {
		t.Errorf("generateNotificationID() length = %d, want 21", len(id))
	}

	// Should contain a dot
	if id[14] != '.' {
		t.Errorf("generateNotificationID() dot position wrong, got %q", id)
	}
}

func TestGenerateNotificationID_Unique(t *testing.T) {
	ids := make(map[string]bool)

	// Generate 100 IDs - they should mostly be unique (may have collisions within same microsecond)
	for i := 0; i < 100; i++ {
		id := generateNotificationID()
		ids[id] = true
		time.Sleep(time.Microsecond)
	}

	// With microsecond precision and sleep, we should have mostly unique IDs
	if len(ids) < 50 {
		t.Errorf("generateNotificationID() generated too few unique IDs: %d", len(ids))
	}
}

// ============ NotificationService Methods Tests ============

func TestNotificationService_IsUserOnline_NoUsers(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer cancel()

	service := NewNotificationService(hub)

	if service.IsUserOnline(1) {
		t.Error("IsUserOnline() should return false for non-connected user")
	}
}

func TestNotificationService_GetOnlineUserCount_Empty(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer cancel()

	service := NewNotificationService(hub)

	if count := service.GetOnlineUserCount(); count != 0 {
		t.Errorf("GetOnlineUserCount() = %d, want 0", count)
	}
}

func TestNotificationService_GetOnlineUserIDs_Empty(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer cancel()

	service := NewNotificationService(hub)

	ids := service.GetOnlineUserIDs()
	if len(ids) != 0 {
		t.Errorf("GetOnlineUserIDs() length = %d, want 0", len(ids))
	}
}

// ============ Integration-like Tests (no actual connections) ============

func TestNotificationService_SendNotification_NoClients(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer cancel()

	service := NewNotificationService(hub)

	// Should not panic when sending to non-existent user
	service.SendNotification(999, "Test", "Message", "info", nil)
}

func TestNotificationService_SendUserUpdate_NoClients(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer cancel()

	service := NewNotificationService(hub)

	// Should not panic when sending to non-existent user
	service.SendUserUpdate(999, "email", "test@example.com")
}

func TestNotificationService_BroadcastNotification_NoClients(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer cancel()

	service := NewNotificationService(hub)

	// Should not panic when broadcasting to no clients
	service.BroadcastNotification("Test", "Message", "info", nil)
}

func TestNotificationService_BroadcastMessage_NoClients(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)
	defer cancel()

	service := NewNotificationService(hub)

	// Should not panic when broadcasting to no clients
	service.BroadcastMessage(MessageTypeNotification, "test payload")
}
