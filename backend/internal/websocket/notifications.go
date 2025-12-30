package websocket

import "time"

// NotificationPayload represents a notification message payload
type NotificationPayload struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // info, success, warning, error
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

// UserUpdatePayload represents a user data update payload
type UserUpdatePayload struct {
	Field string `json:"field"` // What field was updated
	Value any    `json:"value,omitempty"`
}

// NotificationService provides methods to send real-time notifications
type NotificationService struct {
	hub *Hub
}

// NewNotificationService creates a new notification service
func NewNotificationService(hub *Hub) *NotificationService {
	return &NotificationService{hub: hub}
}

// SendNotification sends a notification to a specific user
func (s *NotificationService) SendNotification(userID uint, title, message, notifType string, data any) {
	payload := NotificationPayload{
		ID:        generateNotificationID(),
		Title:     title,
		Message:   message,
		Type:      notifType,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
	s.hub.SendToUser(userID, MessageTypeNotification, payload)
}

// SendUserUpdate notifies a user that their data has been updated
func (s *NotificationService) SendUserUpdate(userID uint, field string, value any) {
	payload := UserUpdatePayload{
		Field: field,
		Value: value,
	}
	s.hub.SendToUser(userID, MessageTypeUserUpdate, payload)
}

// BroadcastNotification sends a notification to all connected users
func (s *NotificationService) BroadcastNotification(title, message, notifType string, data any) {
	payload := NotificationPayload{
		ID:        generateNotificationID(),
		Title:     title,
		Message:   message,
		Type:      notifType,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
	s.hub.Broadcast(MessageTypeNotification, payload)
}

// BroadcastMessage sends a custom message to all connected users
func (s *NotificationService) BroadcastMessage(msgType MessageType, payload any) {
	s.hub.Broadcast(msgType, payload)
}

// IsUserOnline checks if a user is currently connected
func (s *NotificationService) IsUserOnline(userID uint) bool {
	return s.hub.IsUserConnected(userID)
}

// GetOnlineUserCount returns the number of online users
func (s *NotificationService) GetOnlineUserCount() int {
	return s.hub.GetConnectedUserCount()
}

// GetOnlineUserIDs returns the IDs of all online users
func (s *NotificationService) GetOnlineUserIDs() []uint {
	return s.hub.GetConnectedUserIDs()
}

// generateNotificationID generates a unique notification ID
func generateNotificationID() string {
	return time.Now().Format("20060102150405.000000")
}
