package database

import (
	"fmt"
	"log"
	"os"
	"react-golang-starter/internal/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var err error

	// Database configuration - matches Docker container
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "devuser")
	password := getEnv("DB_PASSWORD", "devpass")
	dbname := getEnv("DB_NAME", "devdb")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	log.Printf("Connecting to database: host=%s port=%s user=%s dbname=%s", host, port, user, dbname)

	// Retry connection with exponential backoff
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

		if err == nil {
			log.Println("PostgreSQL database connected successfully")

			// Auto-migrate your models
			err = DB.AutoMigrate(&models.User{})
			if err != nil {
				log.Printf("Failed to migrate database: %v", err)
				continue
			}

			log.Println("Database migration completed")
			return
		}

		if i < maxRetries-1 {
			waitTime := time.Duration(i+1) * 2 * time.Second
			log.Printf("Failed to connect to PostgreSQL database (attempt %d/%d): %v", i+1, maxRetries, err)
			log.Printf("Retrying in %v...", waitTime)
			time.Sleep(waitTime)
		}
	}

	log.Fatal("Failed to connect to PostgreSQL database after", maxRetries, "attempts:", err)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
