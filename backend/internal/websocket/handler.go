package websocket

import (
	"net/http"
	"strings"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/config"
	"react-golang-starter/internal/response"

	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
)

// Handler creates an HTTP handler for WebSocket connections
func Handler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract and validate JWT from cookie
		tokenString, err := extractTokenFromRequest(r)
		if err != nil {
			log.Debug().Err(err).Msg("WebSocket auth: no token found")
			response.Unauthorized(w, r, "Unauthorized")
			return
		}

		// Validate the token
		claims, err := auth.ValidateJWT(tokenString)
		if err != nil {
			log.Debug().Err(err).Msg("WebSocket auth: invalid token")
			response.TokenInvalid(w, r, "Invalid token")
			return
		}

		// Check if token is blacklisted
		if auth.IsTokenBlacklisted(tokenString) {
			log.Debug().Msg("WebSocket auth: token is blacklisted")
			response.TokenInvalid(w, r, "Token has been revoked")
			return
		}

		// Accept the WebSocket connection
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: getAllowedOrigins(),
		})
		if err != nil {
			log.Error().Err(err).Msg("WebSocket accept failed")
			return
		}

		// Create client and register with hub
		client := NewClient(claims.UserID, conn, hub)
		hub.register <- client

		log.Info().
			Uint("user_id", claims.UserID).
			Str("remote_addr", r.RemoteAddr).
			Msg("WebSocket connection established")

		// Start read and write pumps in goroutines
		go client.WritePump(r.Context())
		client.ReadPump(r.Context())
	}
}

// extractTokenFromRequest extracts JWT from cookie or Authorization header
func extractTokenFromRequest(r *http.Request) (string, error) {
	// First try cookie (preferred for browser clients)
	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	// Fall back to Authorization header (for API clients)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1], nil
		}
	}

	// Try query parameter as last resort (for WebSocket clients that can't set headers)
	token := r.URL.Query().Get("token")
	if token != "" {
		return token, nil
	}

	return "", http.ErrNoCookie
}

// getAllowedOrigins returns the list of allowed WebSocket origins.
// Uses the centralized CORS_ALLOWED_ORIGINS config and converts to WebSocket patterns.
func getAllowedOrigins() []string {
	originList := config.GetAllowedOrigins()
	patterns := make([]string, 0, len(originList))

	for _, origin := range originList {
		origin = strings.TrimSpace(origin)
		if origin == "*" {
			// Allow all origins (not recommended for production)
			return []string{"*"}
		}
		// Convert http(s)://host to just host pattern for WebSocket
		origin = strings.TrimPrefix(origin, "https://")
		origin = strings.TrimPrefix(origin, "http://")
		patterns = append(patterns, origin)
	}

	return patterns
}
