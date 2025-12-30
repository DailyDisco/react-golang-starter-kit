package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// generateUniqueToken generates a random unique token
func generateUniqueToken() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("seed_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// SeedConfig contains configuration for seeding
type SeedConfig struct {
	AdminPassword   string
	DefaultPassword string
}

func main() {
	// Load environment variables
	_ = godotenv.Load()

	// Initialize database
	database.ConnectDB()

	config := SeedConfig{
		AdminPassword:   getEnv("SEED_ADMIN_PASSWORD", "admin123!"),
		DefaultPassword: getEnv("SEED_DEFAULT_PASSWORD", "password123!"),
	}

	fmt.Println("🌱 Starting database seeding...")

	// Seed in order
	if err := seedUsers(config); err != nil {
		log.Fatalf("❌ Failed to seed users: %v", err)
	}

	if err := seedFeatureFlags(); err != nil {
		log.Fatalf("❌ Failed to seed feature flags: %v", err)
	}

	fmt.Println("✅ Database seeding completed successfully!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	return string(hash)
}

func seedUsers(config SeedConfig) error {
	fmt.Println("👤 Seeding users...")

	users := []models.User{
		{
			Name:               "Super Admin",
			Email:              "superadmin@example.com",
			Password:           hashPassword(config.AdminPassword),
			Role:               "super_admin",
			IsActive:           true,
			EmailVerified:      true,
			VerificationToken:  generateUniqueToken(),
			PasswordResetToken: generateUniqueToken(),
			StripeCustomerID:   "cus_seed_superadmin",
		},
		{
			Name:               "Admin User",
			Email:              "admin@example.com",
			Password:           hashPassword(config.AdminPassword),
			Role:               "admin",
			IsActive:           true,
			EmailVerified:      true,
			VerificationToken:  generateUniqueToken(),
			PasswordResetToken: generateUniqueToken(),
			StripeCustomerID:   "cus_seed_admin",
		},
		{
			Name:               "Premium User",
			Email:              "premium@example.com",
			Password:           hashPassword(config.DefaultPassword),
			Role:               "premium",
			IsActive:           true,
			EmailVerified:      true,
			VerificationToken:  generateUniqueToken(),
			PasswordResetToken: generateUniqueToken(),
			StripeCustomerID:   "cus_seed_premium",
		},
		{
			Name:               "Regular User",
			Email:              "user@example.com",
			Password:           hashPassword(config.DefaultPassword),
			Role:               "user",
			IsActive:           true,
			EmailVerified:      true,
			VerificationToken:  generateUniqueToken(),
			PasswordResetToken: generateUniqueToken(),
			StripeCustomerID:   "cus_seed_user",
		},
		{
			Name:               "Unverified User",
			Email:              "unverified@example.com",
			Password:           hashPassword(config.DefaultPassword),
			Role:               "user",
			IsActive:           true,
			EmailVerified:      false,
			VerificationToken:  generateUniqueToken(),
			PasswordResetToken: generateUniqueToken(),
			StripeCustomerID:   "cus_seed_unverified",
		},
		{
			Name:               "Inactive User",
			Email:              "inactive@example.com",
			Password:           hashPassword(config.DefaultPassword),
			Role:               "user",
			IsActive:           false,
			EmailVerified:      true,
			VerificationToken:  generateUniqueToken(),
			PasswordResetToken: generateUniqueToken(),
			StripeCustomerID:   "cus_seed_inactive",
		},
	}

	for _, user := range users {
		// Check if user already exists
		var existing models.User
		result := database.DB.Where("email = ?", user.Email).First(&existing)
		if result.Error == nil {
			fmt.Printf("   ⏭️  User %s already exists, skipping\n", user.Email)
			continue
		}

		// Create user
		if err := database.DB.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Email, err)
		}
		fmt.Printf("   ✓ Created user: %s (%s)\n", user.Email, user.Role)
	}

	return nil
}

func seedFeatureFlags() error {
	fmt.Println("🚩 Seeding feature flags...")

	now := time.Now().Format(time.RFC3339)
	flags := []models.FeatureFlag{
		{
			Key:               "dark_mode",
			Name:              "Dark Mode",
			Description:       "Enable dark mode theme",
			Enabled:           true,
			RolloutPercentage: 100,
			AllowedRoles:      pq.StringArray{},
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			Key:               "new_dashboard",
			Name:              "New Dashboard",
			Description:       "Enable the redesigned dashboard UI",
			Enabled:           true,
			RolloutPercentage: 50, // 50% rollout
			AllowedRoles:      pq.StringArray{},
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			Key:               "beta_features",
			Name:              "Beta Features",
			Description:       "Access to beta features",
			Enabled:           true,
			RolloutPercentage: 100,
			AllowedRoles:      pq.StringArray{"premium", "admin", "super_admin"},
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			Key:               "advanced_analytics",
			Name:              "Advanced Analytics",
			Description:       "Show advanced analytics dashboard",
			Enabled:           false,
			RolloutPercentage: 0,
			AllowedRoles:      pq.StringArray{"admin", "super_admin"},
			Metadata:          "{}",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
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

	for _, flag := range flags {
		// Check if flag already exists
		var existing models.FeatureFlag
		result := database.DB.Where("key = ?", flag.Key).First(&existing)
		if result.Error == nil {
			fmt.Printf("   ⏭️  Feature flag '%s' already exists, skipping\n", flag.Key)
			continue
		}

		// Create flag
		if err := database.DB.Create(&flag).Error; err != nil {
			return fmt.Errorf("failed to create feature flag %s: %w", flag.Key, err)
		}
		fmt.Printf("   ✓ Created feature flag: %s (enabled=%v, rollout=%d%%)\n", flag.Key, flag.Enabled, flag.RolloutPercentage)
	}

	return nil
}
