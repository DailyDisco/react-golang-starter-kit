package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/response"
	"time"

	"github.com/rs/zerolog/log"
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
// Tries httpOnly cookie first, then falls back to Authorization header for backwards compatibility
// Also checks if the token has been blacklisted (revoked)
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		var err error

		// Try cookie first (preferred for security)
		tokenString, err = ExtractTokenFromCookie(r)
		if err != nil {
			// Fall back to Authorization header for backwards compatibility
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, r, "Authentication required")
				return
			}
			tokenString, err = ExtractTokenFromHeader(authHeader)
			if err != nil {
				response.Unauthorized(w, r, err.Error())
				return
			}
		}

		// Check if token is blacklisted (revoked)
		if IsTokenBlacklisted(tokenString) {
			response.TokenInvalid(w, r, "Token has been revoked")
			return
		}

		claims, err := ValidateJWT(tokenString)
		if err != nil {
			response.TokenInvalid(w, r, "Invalid token")
			return
		}

		// Try to get user from cache first, fallback to database
		cacheKey := fmt.Sprintf("user:%d", claims.UserID)
		var user models.User

		if err := cache.GetJSON(r.Context(), cacheKey, &user); err != nil {
			// Log cache errors (but not cache misses) for debugging
			var cacheErr *cache.CacheError
			if errors.As(err, &cacheErr) && cacheErr.Err != nil {
				log.Warn().Err(err).Str("key", cacheKey).Msg("cache lookup failed, falling back to database")
			}
			// Fetch from database
			if err := database.DB.First(&user, claims.UserID).Error; err != nil {
				response.Unauthorized(w, r, "User not found")
				return
			}
			// Store in cache for 2 minutes
			_ = cache.SetJSON(r.Context(), cacheKey, &user, 2*time.Minute)
		}

		if !user.IsActive {
			response.AccountInactive(w, r, "Account is deactivated")
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
// Tries httpOnly cookie first, then falls back to Authorization header
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// Try cookie first (preferred for security)
		tokenString, err := ExtractTokenFromCookie(r)
		if err != nil {
			// Fall back to Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				tokenString, _ = ExtractTokenFromHeader(authHeader)
			}
		}

		if tokenString != "" {
			claims, err := ValidateJWT(tokenString)
			if err == nil {
				// Try to get user from cache first, fallback to database
				cacheKey := fmt.Sprintf("user:%d", claims.UserID)
				var user models.User

				if err := cache.GetJSON(r.Context(), cacheKey, &user); err != nil {
					// Log cache errors (but not cache misses) for debugging
					var cacheErr *cache.CacheError
					if errors.As(err, &cacheErr) && cacheErr.Err != nil {
						log.Warn().Err(err).Str("key", cacheKey).Msg("cache lookup failed, falling back to database")
					}
					// Fetch from database
					if err := database.DB.First(&user, claims.UserID).Error; err != nil {
						// User not found, continue without context
						next.ServeHTTP(w, r)
						return
					}
					// Store in cache for 2 minutes
					_ = cache.SetJSON(r.Context(), cacheKey, &user, 2*time.Minute)
				}

				if user.IsActive {
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

// AdminMiddleware checks if the user has admin privileges
// Requires user to have admin or super_admin role
func AdminMiddleware(next http.Handler) http.Handler {
	return AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			response.Unauthorized(w, r, "User not found in context")
			return
		}

		// Check if user has admin or super_admin role
		if !HasRole(user.Role, models.RoleAdmin, models.RoleSuperAdmin) {
			response.Forbidden(w, r, "Admin privileges required")
			return
		}

		next.ServeHTTP(w, r)
	}))
}

// SuperAdminMiddleware checks if the user has super admin privileges
// Requires user to have super_admin role specifically
func SuperAdminMiddleware(next http.Handler) http.Handler {
	return AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			response.Unauthorized(w, r, "User not found in context")
			return
		}

		// Check if user has super_admin role
		if user.Role != models.RoleSuperAdmin {
			response.Forbidden(w, r, "Super admin privileges required")
			return
		}

		next.ServeHTTP(w, r)
	}))
}

// MinRoleLevelMiddleware checks if user's role level meets the minimum required
// Uses the RoleHierarchy from models to compare role levels
func MinRoleLevelMiddleware(minRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				response.Unauthorized(w, r, "User not found in context")
				return
			}

			userLevel, userExists := models.RoleHierarchy[user.Role]
			minLevel, minExists := models.RoleHierarchy[minRole]

			if !userExists || !minExists {
				response.InternalError(w, r, "Invalid role configuration")
				return
			}

			if userLevel < minLevel {
				response.Forbidden(w, r, "Insufficient role level")
				return
			}

			next.ServeHTTP(w, r)
		}))
	}
}

// GetUserRoleFromContext retrieves the user role from the request context
func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleContextKey).(string)
	return role, ok
}

// InvalidateUserCache removes a user from the cache
// Call this when user data changes (update, logout, password change)
func InvalidateUserCache(ctx context.Context, userID uint) error {
	cacheKey := fmt.Sprintf("user:%d", userID)
	return cache.Delete(ctx, cacheKey)
}
