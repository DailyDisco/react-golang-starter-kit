package auth

import (
	"context"
	"net/http"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
)

// ContextKey defines a custom type for context keys to avoid collisions
type ContextKey string

const (
	UserContextKey      ContextKey = "user"
	UserIDContextKey    ContextKey = "user_id"
	UserEmailContextKey ContextKey = "user_email"
	UserRoleContextKey  ContextKey = "user_role"
)

// AuthMiddleware validates JWT tokens and adds user context to requests
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Fetch user from database to ensure they still exist and are active
		var user models.User
		if err := database.DB.First(&user, claims.UserID).Error; err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		if !user.IsActive {
			http.Error(w, "Account is deactivated", http.StatusUnauthorized)
			return
		}

		// Add user context to request
		ctx := context.WithValue(r.Context(), UserContextKey, &user)
		ctx = context.WithValue(ctx, UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, UserEmailContextKey, user.Email)
		ctx = context.WithValue(ctx, UserRoleContextKey, user.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware adds user context if token is present but doesn't require it
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString, err := ExtractTokenFromHeader(authHeader)
			if err == nil {
				claims, err := ValidateJWT(tokenString)
				if err == nil {
					// Fetch user from database
					var user models.User
					if err := database.DB.First(&user, claims.UserID).Error; err == nil && user.IsActive {
						// Add user context to request
						ctx := context.WithValue(r.Context(), UserContextKey, &user)
						ctx = context.WithValue(ctx, UserIDContextKey, user.ID)
						ctx = context.WithValue(ctx, UserEmailContextKey, user.Email)
						ctx = context.WithValue(ctx, UserRoleContextKey, user.Role)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}
		}

		// No valid auth, continue without user context
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(uint)
	return userID, ok
}

// GetUserEmailFromContext retrieves the user email from the request context
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailContextKey).(string)
	return email, ok
}

// CORSMiddleware adds CORS headers for authentication
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // In production, specify your frontend URL
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AdminMiddleware checks if the user has admin privileges (extend User model for roles later)
func AdminMiddleware(next http.Handler) http.Handler {
	return AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		// TODO: Add role checking logic here when roles are implemented
		// For now, allow all authenticated users to pass through
		// Example: if user.Role != "admin" {
		//     http.Error(w, "Admin privileges required", http.StatusForbidden)
		//     return
		// }

		// Suppress unused variable warning - user will be used when role checking is implemented
		_ = user

		next.ServeHTTP(w, r)
	}))
}

// GetUserRoleFromContext retrieves the user role from the request context
func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleContextKey).(string)
	return role, ok
}
