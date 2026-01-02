package middleware

import (
	"net/http"

	"react-golang-starter/internal/config"
)

// SetCORSErrorHeaders sets CORS headers on error responses.
// This is needed because middleware error responses (CSRF, rate limit)
// exit early before the CORS middleware can add headers to the response.
func SetCORSErrorHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}

	// Check if origin is allowed
	allowedOrigins := config.GetAllowedOrigins()
	originAllowed := false
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			originAllowed = true
			break
		}
	}

	if !originAllowed {
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
