package middleware

import (
	"net/http"
	"os"
	"strings"
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
	allowedOrigins := getAllowedOrigins()
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

// getAllowedOrigins returns the allowed CORS origins from environment variables
func getAllowedOrigins() []string {
	originsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	if originsEnv != "" {
		return strings.Split(originsEnv, ",")
	}

	// Default development origins
	return []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:3002",
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:5175",
		"http://localhost:5193",
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
	}
}
