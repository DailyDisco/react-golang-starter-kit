package database

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"react-golang-starter/internal/models"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedConfig contains configuration for seeding
type SeedConfig struct {
	Enabled         bool
	AdminPassword   string
	DefaultPassword string
}

// LoadSeedConfig loads seed configuration from environment variables
func LoadSeedConfig() *SeedConfig {
	return &SeedConfig{
		Enabled:         seedGetEnvBool("AUTO_SEED", false),
		AdminPassword:   seedGetEnv("SEED_ADMIN_PASSWORD", "admin123!"),
		DefaultPassword: seedGetEnv("SEED_DEFAULT_PASSWORD", "password123!"),
	}
}

func seedGetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func seedGetEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

// SeedAll runs all seed functions (users and feature flags)
// This is idempotent - existing data is not modified
// WARNING: Only runs in development/test environments
func SeedAll(config *SeedConfig) error {
	env := seedGetEnv("GO_ENV", "development")
	if env != "development" && env != "test" {
		log.Warn().Str("env", env).Msg("AUTO_SEED ignored - only allowed in development/test")
		return nil
	}

	log.Info().Str("env", env).Msg("Starting database seeding...")

	// Warn about default passwords
	if config.AdminPassword == "admin123!" || config.DefaultPassword == "password123!" {
		log.Warn().Msg("Using default seed passwords - change SEED_ADMIN_PASSWORD and SEED_DEFAULT_PASSWORD for security")
	}

	if err := SeedUsers(config); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	// Feature flags are seeded separately (already called in main.go)
	// but we include them here for completeness when running full seed
	if err := SeedFeatureFlags(); err != nil {
		return fmt.Errorf("failed to seed feature flags: %w", err)
	}

	log.Info().Msg("Database seeding completed")
	return nil
}

// generateUniqueToken generates a random unique token
func generateUniqueToken() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("seed_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// SeedUsersManual creates initial test users (for manual seed command)
// Bypasses environment check since the command has its own confirmation
func SeedUsersManual(config *SeedConfig) error {
	return seedUsersInternal(config)
}

// SeedUsers creates initial test users if they don't exist
// This is idempotent - users that already exist are not modified
func SeedUsers(config *SeedConfig) error {
	return seedUsersInternal(config)
}

// seedUsersInternal is the internal implementation for seeding users
// Uses a transaction to ensure atomic seeding - all users are created or none
func seedUsersInternal(config *SeedConfig) error {
	adminHash, err := hashPassword(config.AdminPassword)
	if err != nil {
		return err
	}

	defaultHash, err := hashPassword(config.DefaultPassword)
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	farFuture := time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339)

	users := []models.User{
		{
			Name:                 "Super Admin",
			Email:                "superadmin@example.com",
			Password:             adminHash,
			Role:                 "super_admin",
			IsActive:             true,
			EmailVerified:        true,
			VerificationToken:    generateUniqueToken(),
			VerificationExpires:  farFuture,
			PasswordResetToken:   generateUniqueToken(),
			PasswordResetExpires: farFuture,
			StripeCustomerID:     "cus_seed_superadmin",
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		{
			Name:                 "Admin User",
			Email:                "admin@example.com",
			Password:             adminHash,
			Role:                 "admin",
			IsActive:             true,
			EmailVerified:        true,
			VerificationToken:    generateUniqueToken(),
			VerificationExpires:  farFuture,
			PasswordResetToken:   generateUniqueToken(),
			PasswordResetExpires: farFuture,
			StripeCustomerID:     "cus_seed_admin",
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		{
			Name:                 "Premium User",
			Email:                "premium@example.com",
			Password:             defaultHash,
			Role:                 "premium",
			IsActive:             true,
			EmailVerified:        true,
			VerificationToken:    generateUniqueToken(),
			VerificationExpires:  farFuture,
			PasswordResetToken:   generateUniqueToken(),
			PasswordResetExpires: farFuture,
			StripeCustomerID:     "cus_seed_premium",
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		{
			Name:                 "Regular User",
			Email:                "user@example.com",
			Password:             defaultHash,
			Role:                 "user",
			IsActive:             true,
			EmailVerified:        true,
			VerificationToken:    generateUniqueToken(),
			VerificationExpires:  farFuture,
			PasswordResetToken:   generateUniqueToken(),
			PasswordResetExpires: farFuture,
			StripeCustomerID:     "cus_seed_user",
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		{
			Name:                 "Unverified User",
			Email:                "unverified@example.com",
			Password:             defaultHash,
			Role:                 "user",
			IsActive:             true,
			EmailVerified:        false,
			VerificationToken:    generateUniqueToken(),
			VerificationExpires:  farFuture,
			PasswordResetToken:   generateUniqueToken(),
			PasswordResetExpires: farFuture,
			StripeCustomerID:     "cus_seed_unverified",
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		{
			Name:                 "Inactive User",
			Email:                "inactive@example.com",
			Password:             defaultHash,
			Role:                 "user",
			IsActive:             false,
			EmailVerified:        true,
			VerificationToken:    generateUniqueToken(),
			VerificationExpires:  farFuture,
			PasswordResetToken:   generateUniqueToken(),
			PasswordResetExpires: farFuture,
			StripeCustomerID:     "cus_seed_inactive",
			CreatedAt:            now,
			UpdatedAt:            now,
		},
	}

	seededCount := 0

	// Use a transaction for atomic seeding
	err = DB.Transaction(func(tx *gorm.DB) error {
		for _, user := range users {
			// Check if user already exists
			var existing models.User
			result := tx.Where("email = ?", user.Email).First(&existing)
			if result.Error == nil {
				continue
			}

			// Create user
			if err := tx.Create(&user).Error; err != nil {
				log.Warn().
					Err(err).
					Str("email", user.Email).
					Msg("Failed to seed user")
				// Return error to rollback transaction
				return fmt.Errorf("failed to seed user %s: %w", user.Email, err)
			}
			seededCount++
			log.Debug().
				Str("email", user.Email).
				Str("role", user.Role).
				Msg("Seeded user")
		}
		return nil
	})

	if err != nil {
		return err
	}

	if seededCount > 0 {
		log.Info().
			Int("count", seededCount).
			Msg("Users seeded")
	}

	return nil
}

// SeedFeatureFlagsExtended seeds additional feature flags beyond the base set
// This is called by the seed command for a more complete dataset
func SeedFeatureFlagsExtended() error {
	now := time.Now().Format(time.RFC3339)

	extendedFlags := []models.FeatureFlag{
		{
			Key:               "file_sharing",
			Name:              "File Sharing",
			Description:       "Enable file sharing functionality",
			Enabled:           true,
			RolloutPercentage: 100,
			AllowedRoles:      pq.StringArray{},
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	}

	seededCount := 0
	for _, flag := range extendedFlags {
		var existing models.FeatureFlag
		result := DB.Where("key = ?", flag.Key).First(&existing)
		if result.Error == nil {
			continue
		}

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
			Msg("Extended feature flags seeded")
	}

	return nil
}
