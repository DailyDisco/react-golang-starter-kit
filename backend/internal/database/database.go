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
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDB initializes the GORM database connection.
// If the shared pgx pool is initialized (via InitPool), GORM will use it.
// Otherwise, GORM creates its own connection pool (backward compatible).
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

	// Configure GORM with prepared statements for better query performance
	gormConfig := &gorm.Config{
		// Enable prepared statement cache for faster query execution
		PrepareStmt: true,
	}

	// Configure query profiler for slow query detection
	profilerConfig := LoadProfilerConfig()
	if profilerConfig.Enabled {
		gormConfig.Logger = NewQueryProfiler()
		log.Info().
			Dur("slow_query_threshold", profilerConfig.SlowQueryThreshold).
			Bool("log_all_queries", profilerConfig.LogAllQueries).
			Msg("Query profiler enabled")
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	}

	// Check if shared pool is available
	useSharedPool := Pool != nil
	if useSharedPool {
		log.Info().Msg("Using shared pgx pool for GORM")
	}

	// Retry connection with exponential backoff
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), gormConfig)

		if err == nil {
			log.Info().
				Bool("prepared_stmt", true).
				Bool("shared_pool", useSharedPool).
				Msg("PostgreSQL database connected successfully")

			// Configure connection pool for optimal performance
			// Note: When using shared pool, these settings are managed by the pool
			sqlDB, poolErr := DB.DB()
			if poolErr != nil {
				log.Error().Err(poolErr).Msg("Failed to get underlying database connection")
				continue
			}

			if !useSharedPool {
				// Only configure pool settings if not using shared pool
				sqlDB.SetMaxOpenConns(25)
				sqlDB.SetMaxIdleConns(5)
				sqlDB.SetConnMaxLifetime(5 * time.Minute)
				sqlDB.SetConnMaxIdleTime(1 * time.Minute)

				log.Info().
					Int("max_open", 25).
					Int("max_idle", 5).
					Dur("max_lifetime", 5*time.Minute).
					Dur("max_idle_time", 1*time.Minute).
					Msg("Database connection pool configured")
			}

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
