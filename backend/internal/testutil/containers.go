//go:build integration

// Package testutil provides testing utilities including container-based infrastructure.
package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"react-golang-starter/internal/models"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresContainer wraps a testcontainers PostgreSQL instance.
type PostgresContainer struct {
	Container testcontainers.Container
	DB        *gorm.DB
	DSN       string
}

// SetupPostgresContainer starts a PostgreSQL container for integration tests.
// The container is automatically terminated when the test completes.
func SetupPostgresContainer(t *testing.T) *PostgresContainer {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("test_db"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Warning: Failed to terminate PostgreSQL container: %v", err)
		}
	})

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect with GORM
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return &PostgresContainer{
		Container: container,
		DB:        db,
		DSN:       connStr,
	}
}

// RedisContainer wraps a testcontainers Redis instance.
type RedisContainer struct {
	Container testcontainers.Container
	Addr      string
}

// SetupRedisContainer starts a Redis container for integration tests.
func SetupRedisContainer(t *testing.T) *RedisContainer {
	t.Helper()
	ctx := context.Background()

	container, err := tcredis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Warning: Failed to terminate Redis container: %v", err)
		}
	})

	// Get connection address
	addr, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("Failed to get Redis address: %v", err)
	}

	return &RedisContainer{
		Container: container,
		Addr:      addr,
	}
}

// TestInfrastructure holds all test infrastructure containers.
type TestInfrastructure struct {
	Postgres *PostgresContainer
	Redis    *RedisContainer
}

// SetupTestInfrastructure starts all required containers for integration tests.
// Use this when tests need both database and cache.
func SetupTestInfrastructure(t *testing.T) *TestInfrastructure {
	t.Helper()

	return &TestInfrastructure{
		Postgres: SetupPostgresContainer(t),
		Redis:    SetupRedisContainer(t),
	}
}

// runMigrations runs all database migrations.
func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.TokenBlacklist{},
		&models.Organization{},
		&models.OrganizationMember{},
		&models.OrganizationInvitation{},
		&models.Subscription{},
		&models.OAuthProvider{},
		&models.AuditLog{},
		&models.FeatureFlag{},
		&models.UserFeatureFlag{},
		&models.File{},
		&models.UserAPIKey{},
		&models.UserPreferences{},
		&models.UserTwoFactor{},
		&models.UserSession{},
		&models.SystemSetting{},
		&models.IPBlocklist{},
		&models.LoginHistory{},
		&models.AnnouncementBanner{},
		&models.UserAnnouncementRead{},
		&models.UserDismissedAnnouncement{},
		&models.EmailTemplate{},
		&models.DataExport{},
	)
}

// WithPostgresDB is a helper that runs a test function with a clean database.
// Each call gets a fresh transaction that is rolled back after the test.
func WithPostgresDB(t *testing.T, pg *PostgresContainer, fn func(db *gorm.DB)) {
	t.Helper()

	// Start a transaction for isolation
	tx := pg.DB.Begin()
	if tx.Error != nil {
		t.Fatalf("Failed to begin transaction: %v", tx.Error)
	}

	// Create savepoint
	savepoint := fmt.Sprintf("test_%d", time.Now().UnixNano())
	if err := tx.Exec(fmt.Sprintf("SAVEPOINT %s", savepoint)).Error; err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create savepoint: %v", err)
	}

	// Ensure rollback on completion
	defer func() {
		tx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", savepoint))
		tx.Rollback()
	}()

	fn(tx)
}
