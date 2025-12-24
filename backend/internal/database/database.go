package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"react-golang-starter/internal/models"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var err error

	// Database configuration - supports both Railway and local development
	host := getEnv("PGHOST", getEnv("DB_HOST", "localhost"))
	port := getEnv("PGPORT", getEnv("DB_PORT", "5432"))
	user := getEnv("PGUSER", getEnv("DB_USER", "devuser"))
	password := getEnv("PGPASSWORD", getEnv("DB_PASSWORD", "devpass"))
	dbname := getEnv("PGDATABASE", getEnv("DB_NAME", "devdb"))
	sslmode := getEnv("DB_SSLMODE", "disable") // Default to disable for local development

	log.Debug().
		Str("DB_SSLMODE", os.Getenv("DB_SSLMODE")).
		Str("DB_HOST", os.Getenv("DB_HOST")).
		Str("DB_PORT", os.Getenv("DB_PORT")).
		Str("DB_USER", os.Getenv("DB_USER")).
		Str("DB_NAME", os.Getenv("DB_NAME")).
		Msg("Environment variables loaded")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	log.Info().
		Str("host", host).
		Str("port", port).
		Str("user", user).
		Str("dbname", dbname).
		Str("sslmode", sslmode).
		Msg("Connecting to database")

	// Retry connection with exponential backoff
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

		if err == nil {
			log.Info().Msg("PostgreSQL database connected successfully")

			// Configure connection pool for optimal performance
			sqlDB, poolErr := DB.DB()
			if poolErr != nil {
				log.Error().Err(poolErr).Msg("Failed to get underlying database connection")
				continue
			}

			// SetMaxOpenConns sets the maximum number of open connections to the database
			sqlDB.SetMaxOpenConns(25)

			// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
			sqlDB.SetMaxIdleConns(5)

			// SetConnMaxLifetime sets the maximum amount of time a connection may be reused
			sqlDB.SetConnMaxLifetime(5 * time.Minute)

			// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle
			sqlDB.SetConnMaxIdleTime(1 * time.Minute)

			log.Info().
				Int("max_open", 25).
				Int("max_idle", 5).
				Dur("max_lifetime", 5*time.Minute).
				Dur("max_idle_time", 1*time.Minute).
				Msg("Database connection pool configured")

			// Run migrations if enabled (via RUN_MIGRATIONS=true)
			// For production, run migrations manually using: make migrate-up
			if err := AutoRunMigrations(); err != nil {
				log.Warn().Err(err).Msg("Auto-migration failed - migrations can be run manually")
				// Continue without failing - migrations can be run manually
			}

			log.Info().Msg("Database ready")
			return
		}

		if i < maxRetries-1 {
			waitTime := time.Duration(i+1) * 2 * time.Second
			log.Error().
				Err(err).
				Int("attempt", i+1).
				Int("max_retries", maxRetries).
				Dur("retry_in", waitTime).
				Msg("Failed to connect to PostgreSQL database, retrying")
			time.Sleep(waitTime)
		}
	}

	log.Fatal().
		Err(err).
		Int("max_retries", maxRetries).
		Msg("Failed to connect to PostgreSQL database after max retries")
}

// CheckDatabaseHealth pings the database to check its connectivity
func CheckDatabaseHealth() models.ComponentStatus {
	sqlDB, err := DB.DB()
	if err != nil {
		return models.ComponentStatus{
			Name:    "database",
			Status:  "unhealthy",
			Message: fmt.Sprintf("failed to get underlying database connection: %v", err),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = sqlDB.PingContext(ctx)
	if err != nil {
		return models.ComponentStatus{
			Name:    "database",
			Status:  "unhealthy",
			Message: fmt.Sprintf("failed to ping database: %v", err),
		}
	}

	return models.ComponentStatus{
		Name:   "database",
		Status: "healthy",
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
