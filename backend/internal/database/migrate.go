package database

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

// MigrationConfig holds migration configuration
type MigrationConfig struct {
	MigrationsPath string
	DatabaseName   string
}

// DefaultMigrationConfig returns default migration configuration
func DefaultMigrationConfig() *MigrationConfig {
	return &MigrationConfig{
		MigrationsPath: "./migrations",
		DatabaseName:   "postgres",
	}
}

// LoadMigrationConfig loads migration configuration from environment
func LoadMigrationConfig() *MigrationConfig {
	config := DefaultMigrationConfig()

	if path := os.Getenv("MIGRATIONS_PATH"); path != "" {
		config.MigrationsPath = path
	}

	return config
}

// NewMigrator creates a new migrate instance using the global DB connection
func NewMigrator(cfg *MigrationConfig) (*migrate.Migrate, error) {
	if DB == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	// Get underlying *sql.DB from GORM
	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying DB: %w", err)
	}

	// Create postgres driver
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{
		DatabaseName: cfg.DatabaseName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Create migrator with file source
	sourcePath := fmt.Sprintf("file://%s", cfg.MigrationsPath)
	m, err := migrate.NewWithDatabaseInstance(sourcePath, cfg.DatabaseName, driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return m, nil
}

// RunMigrations runs all pending migrations
func RunMigrations(migrationsPath string) error {
	cfg := &MigrationConfig{
		MigrationsPath: migrationsPath,
		DatabaseName:   "postgres",
	}

	m, err := NewMigrator(cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Info().Msg("no new migrations to apply")
	} else {
		log.Info().Msg("migrations applied successfully")
	}

	return nil
}

// MigrateDown rolls back n migrations
func MigrateDown(migrationsPath string, steps int) error {
	cfg := &MigrationConfig{
		MigrationsPath: migrationsPath,
		DatabaseName:   "postgres",
	}

	m, err := NewMigrator(cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rollback failed: %w", err)
	}

	log.Info().Int("steps", steps).Msg("migrations rolled back")
	return nil
}

// GetMigrationVersion returns current migration version
func GetMigrationVersion(migrationsPath string) (uint, bool, error) {
	cfg := &MigrationConfig{
		MigrationsPath: migrationsPath,
		DatabaseName:   "postgres",
	}

	m, err := NewMigrator(cfg)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, fmt.Errorf("failed to get version: %w", err)
	}

	return version, dirty, nil
}

// MigrateToVersion migrates to a specific version
func MigrateToVersion(migrationsPath string, version uint) error {
	cfg := &MigrationConfig{
		MigrationsPath: migrationsPath,
		DatabaseName:   "postgres",
	}

	m, err := NewMigrator(cfg)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Migrate(version); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration to version %d failed: %w", version, err)
	}

	log.Info().Uint("version", version).Msg("migrated to version")
	return nil
}

// AutoRunMigrations checks environment and runs migrations if enabled
func AutoRunMigrations() error {
	if os.Getenv("RUN_MIGRATIONS") != "true" {
		log.Debug().Msg("auto-migrations disabled (set RUN_MIGRATIONS=true to enable)")
		return nil
	}

	config := LoadMigrationConfig()
	log.Info().Str("path", config.MigrationsPath).Msg("running auto-migrations")

	return RunMigrations(config.MigrationsPath)
}
