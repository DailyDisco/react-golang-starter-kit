package websocket

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = 54 * time.Second

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB

	// Size of the client's send channel buffer
	sendBufferSize = 256
)

// Client represents a WebSocket client connection
type Client struct {
	// User ID associated with this connection
	UserID uint

	// The WebSocket connection
	conn *websocket.Conn

	// Channel for outbound messages
	send chan Message

	// Reference to the hub
	hub *Hub
}

// NewClient creates a new WebSocket client
func NewClient(userID uint, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		UserID: userID,
		conn:   conn,
		send:   make(chan Message, sendBufferSize),
		hub:    hub,
	}
}

// IncomingMessage represents a message received from the client
type IncomingMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close(websocket.StatusNormalClosure, "connection closed")
	}()

	c.conn.SetReadLimit(maxMessageSize)

	for {
		var msg IncomingMessage
		err := wsjson.Read(ctx, c.conn, &msg)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				log.Debug().Uint("user_id", c.UserID).Msg("WebSocket connection closed normally")
			} else {
				log.Debug().Err(err).Uint("user_id", c.UserID).Msg("WebSocket read error")
			}
			return
		}

		// Handle incoming messages
		c.handleMessage(ctx, msg)
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(ctx context.Context, msg IncomingMessage) {
	switch msg.Type {
	case MessageTypePing:
		// Respond with pong
		c.send <- Message{Type: MessageTypePong}
	default:
		// Log unknown message types for debugging
		log.Debug().
			Uint("user_id", c.UserID).
			Str("type", string(msg.Type)).
			Msg("Received WebSocket message")
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// Hub closed the channel
				return
			}

			writeCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := wsjson.Write(writeCtx, c.conn, msg)
			cancel()

			if err != nil {
				log.Debug().Err(err).Uint("user_id", c.UserID).Msg("WebSocket write error")
				return
			}

		case <-ticker.C:
			// Send ping to keep connection alive
			pingCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.conn.Ping(pingCtx)
			cancel()

			if err != nil {
				log.Debug().Err(err).Uint("user_id", c.UserID).Msg("WebSocket ping failed")
				return
			}

		case <-ctx.Done():
			return
		}
	}
}
