package jobs

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/rs/zerolog/log"
)

// Client wraps the River client with our configuration
type Client struct {
	river  *river.Client[pgx.Tx]
	pool   *pgxpool.Pool
	config *Config
}

var instance *Client

// Initialize sets up the River job client
func Initialize(config *Config) error {
	if !config.Enabled {
		log.Info().Msg("job system disabled")
		instance = nil
		return nil
	}

	// Build database URL from environment (same vars as GORM uses)
	dbURL := buildDatabaseURL()

	// Create pgx pool for River
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return fmt.Errorf("failed to create pgx pool: %w", err)
	}

	// Run River's migrations to ensure schema is up to date
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		pool.Close()
		return fmt.Errorf("failed to create River migrator: %w", err)
	}

	_, err = migrator.Migrate(context.Background(), rivermigrate.DirectionUp, nil)
	if err != nil {
		pool.Close()
		return fmt.Errorf("failed to run River migrations: %w", err)
	}
	log.Info().Msg("River schema migrations completed")

	// Create workers registry
	workers := river.NewWorkers()

	// Register all job workers
	river.AddWorker(workers, &SendVerificationEmailWorker{})
	river.AddWorker(workers, &SendPasswordResetEmailWorker{})
	river.AddWorker(workers, &ProcessStripeWebhookWorker{})

	// Create River client
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: config.WorkerCount},
			"email":            {MaxWorkers: 5},
			"webhooks":         {MaxWorkers: 3},
		},
		Workers:              workers,
		JobTimeout:           config.JobTimeout,
		RescueStuckJobsAfter: config.RescueStuckJobsAfter,
	})
	if err != nil {
		pool.Close()
		return fmt.Errorf("failed to create River client: %w", err)
	}

	instance = &Client{
		river:  riverClient,
		pool:   pool,
		config: config,
	}

	log.Info().
		Int("workers", config.WorkerCount).
		Msg("River job client initialized")

	return nil
}

// Start begins processing jobs
func Start(ctx context.Context) error {
	if instance == nil {
		return nil // Jobs disabled
	}

	if err := instance.river.Start(ctx); err != nil {
		return fmt.Errorf("failed to start River client: %w", err)
	}

	log.Info().Msg("job processing started")
	return nil
}

// Stop gracefully shuts down job processing
func Stop(ctx context.Context) error {
	if instance == nil {
		return nil
	}

	if err := instance.river.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop River client: %w", err)
	}

	// Close the pgx pool
	instance.pool.Close()

	log.Info().Msg("job processing stopped")
	return nil
}

// Insert adds a job to the queue
func Insert[T river.JobArgs](ctx context.Context, args T, opts *river.InsertOpts) error {
	if instance == nil {
		return fmt.Errorf("job system not initialized or disabled")
	}

	_, err := instance.river.Insert(ctx, args, opts)
	return err
}

// InsertMany adds multiple jobs to the queue
func InsertMany(ctx context.Context, params []river.InsertManyParams) error {
	if instance == nil {
		return fmt.Errorf("job system not initialized or disabled")
	}

	_, err := instance.river.InsertMany(ctx, params)
	return err
}

// IsAvailable returns whether the job system is available
func IsAvailable() bool {
	return instance != nil
}

// GetClient returns the underlying River client (for advanced usage)
func GetClient() *river.Client[pgx.Tx] {
	if instance == nil {
		return nil
	}
	return instance.river
}

// buildDatabaseURL constructs the database URL from environment variables
func buildDatabaseURL() string {
	host := getEnv("PGHOST", getEnv("DB_HOST", "localhost"))
	port := getEnv("PGPORT", getEnv("DB_PORT", "5432"))
	user := getEnv("PGUSER", getEnv("DB_USER", "devuser"))
	password := getEnv("PGPASSWORD", getEnv("DB_PASSWORD", "devpass"))
	dbname := getEnv("PGDATABASE", getEnv("DB_NAME", "devdb"))
	sslmode := getEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
