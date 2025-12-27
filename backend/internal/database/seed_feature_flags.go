package database

import (
	"time"

	"react-golang-starter/internal/models"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// SeedFeatureFlags creates initial feature flags if they don't exist
// This is idempotent - flags that already exist are not modified
func SeedFeatureFlags() error {
	now := time.Now().Format(time.RFC3339)

	initialFlags := []models.FeatureFlag{
		{
			Key:               "dark_mode",
			Name:              "Dark Mode",
			Description:       "Enable dark mode theme toggle",
			Enabled:           true,
			RolloutPercentage: 100,
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			Key:               "beta_features",
			Name:              "Beta Features",
			Description:       "Access to beta features for testing",
			Enabled:           true,
			RolloutPercentage: 0,
			AllowedRoles:      pq.StringArray{"admin", "super_admin"},
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			Key:               "premium_features",
			Name:              "Premium Features",
			Description:       "Premium tier features for paid subscribers",
			Enabled:           true,
			RolloutPercentage: 100,
			AllowedRoles:      pq.StringArray{"premium", "admin", "super_admin"},
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			Key:               "new_dashboard",
			Name:              "New Dashboard",
			Description:       "Experimental new dashboard UI",
			Enabled:           true,
			RolloutPercentage: 0,
			AllowedRoles:      pq.StringArray{"admin", "super_admin"},
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			Key:               "file_preview",
			Name:              "File Preview",
			Description:       "Enable file preview in file manager",
			Enabled:           true,
			RolloutPercentage: 100,
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			Key:               "oauth_login",
			Name:              "OAuth Login",
			Description:       "Enable OAuth social login options",
			Enabled:           true,
			RolloutPercentage: 100,
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			Key:               "admin_impersonation",
			Name:              "Admin Impersonation",
			Description:       "Allow admins to impersonate users for support",
			Enabled:           true,
			RolloutPercentage: 100,
			AllowedRoles:      pq.StringArray{"admin", "super_admin"},
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			Key:               "advanced_analytics",
			Name:              "Advanced Analytics",
			Description:       "Advanced analytics dashboard",
			Enabled:           true,
			RolloutPercentage: 100,
			AllowedRoles:      pq.StringArray{"admin", "super_admin"},
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	}

	seededCount := 0
	for _, flag := range initialFlags {
		// Check if flag already exists
		var existing models.FeatureFlag
		result := DB.Where("key = ?", flag.Key).First(&existing)
		if result.Error == nil {
			// Flag already exists, skip
			continue
		}

		// Create the flag
		if err := DB.Create(&flag).Error; err != nil {
			log.Warn().
				Err(err).
				Str("key", flag.Key).
				Msg("Failed to seed feature flag")
			continue
		}
		seededCount++
	}

	if seededCount > 0 {
		log.Info().
			Int("count", seededCount).
			Msg("Feature flags seeded")
	}

	return nil
}
