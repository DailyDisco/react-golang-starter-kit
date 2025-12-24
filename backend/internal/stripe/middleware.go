package stripe

import (
	"encoding/json"
	"net/http"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
)

// RequirePremium middleware ensures the user has an active premium subscription
func RequirePremium(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.GetUserIDFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not authenticated",
				Code:    http.StatusUnauthorized,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Get user from database
		var user models.User
		if err := database.DB.First(&user, userID).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not found",
				Code:    http.StatusUnauthorized,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Check if user has premium or higher role
		if models.RoleHierarchy[user.Role] < models.RoleHierarchy[models.RolePremium] {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Forbidden",
				Message: "Premium subscription required",
				Code:    http.StatusForbidden,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireActiveSubscription middleware checks for an active subscription in the database
func RequireActiveSubscription(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.GetUserIDFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not authenticated",
				Code:    http.StatusUnauthorized,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Check for active subscription
		var subscription models.Subscription
		if err := database.DB.Where("user_id = ? AND status IN ?", userID, []string{
			models.SubscriptionStatusActive,
			models.SubscriptionStatusTrialing,
			models.SubscriptionStatusPastDue, // Grace period
		}).First(&subscription).Error; err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			if err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Error:   "Payment Required",
				Message: "Active subscription required",
				Code:    http.StatusPaymentRequired,
			}); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}
