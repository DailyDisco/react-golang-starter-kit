package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"react-golang-starter/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()

	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}

	// Require confirmation for non-development environments
	if env != "development" && env != "test" {
		fmt.Printf("⚠️  WARNING: You are about to seed the %s database!\n", strings.ToUpper(env))
		fmt.Println("This will create test users with known passwords.")
		fmt.Print("Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if strings.TrimSpace(input) != "yes" {
			fmt.Println("❌ Seeding cancelled.")
			os.Exit(1)
		}
	}

	// Initialize database
	database.ConnectDB()

	fmt.Println("🌱 Starting database seeding...")

	// Load seed config and run all seeds (bypasses env check for manual runs)
	config := database.LoadSeedConfig()
	if err := database.SeedUsersManual(config); err != nil {
		log.Fatalf("❌ Failed to seed users: %v", err)
	}

	if err := database.SeedFeatureFlags(); err != nil {
		log.Fatalf("❌ Failed to seed feature flags: %v", err)
	}

	if err := database.SeedFeatureFlagsExtended(); err != nil {
		log.Fatalf("❌ Failed to seed extended feature flags: %v", err)
	}

	fmt.Println("✅ Database seeding completed successfully!")
}
