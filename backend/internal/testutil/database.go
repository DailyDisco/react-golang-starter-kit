// Package testutil provides testing utilities including database helpers.
package testutil

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testDB      *gorm.DB
	testDBOnce  sync.Once
	migrateOnce sync.Once
	testDBMu    sync.Mutex
	migrateErr  error
)

// TestDBConfig holds test database configuration
type TestDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// DefaultTestDBConfig returns default test database configuration
func DefaultTestDBConfig() TestDBConfig {
	return TestDBConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_DB_PORT", "5433"),
		User:     getEnvOrDefault("TEST_DB_USER", "testuser"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "testpass"),
		DBName:   getEnvOrDefault("TEST_DB_NAME", "starter_kit_test"),
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// IsIntegrationTest returns true if integration tests should run
func IsIntegrationTest() bool {
	return os.Getenv("INTEGRATION_TEST") == "true"
}

// SkipIfNotIntegration skips the test if not running integration tests
func SkipIfNotIntegration(t *testing.T) {
	t.Helper()
	if !IsIntegrationTest() {
		t.Skip("Skipping integration test (set INTEGRATION_TEST=true to run)")
	}
}

// GetTestDB returns the shared test database connection.
// It initializes the connection on first call and reuses it for subsequent calls.
func GetTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	testDBOnce.Do(func() {
		cfg := DefaultTestDBConfig()
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
		)

		var err error
		testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to connect to test database")
		}

		// Configure connection pool for testing
		sqlDB, err := testDB.DB()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get underlying database connection")
		}
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(time.Minute)
	})

	return testDB
}

// SetupTestDB sets up the test database with migrations.
// Call this at the start of integration test suites.
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	SkipIfNotIntegration(t)

	db := GetTestDB(t)

	// Run migrations only once across all tests
	migrateOnce.Do(func() {
		migrateErr = db.AutoMigrate(
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
	})
	if migrateErr != nil {
		t.Fatalf("Failed to run migrations: %v", migrateErr)
	}

	// Set global DB for handlers that use it
	database.DB = db

	return db
}

// TestTransaction wraps a test in a database transaction that is rolled back.
// This provides isolation between tests while allowing parallel execution.
type TestTransaction struct {
	DB        *gorm.DB
	tx        *gorm.DB
	savepoint string
	t         *testing.T
}

// NewTestTransaction starts a new transaction for test isolation.
func NewTestTransaction(t *testing.T, db *gorm.DB) *TestTransaction {
	t.Helper()

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("Failed to begin transaction: %v", tx.Error)
	}

	// Create unique savepoint name
	savepoint := fmt.Sprintf("test_%d", time.Now().UnixNano())
	if err := tx.Exec(fmt.Sprintf("SAVEPOINT %s", savepoint)).Error; err != nil {
		tx.Rollback()
		t.Fatalf("Failed to create savepoint: %v", err)
	}

	tt := &TestTransaction{
		DB:        tx,
		tx:        tx,
		savepoint: savepoint,
		t:         t,
	}

	// Register cleanup
	t.Cleanup(func() {
		tt.Rollback()
	})

	return tt
}

// Rollback rolls back the transaction to the savepoint.
func (tt *TestTransaction) Rollback() {
	if tt.tx == nil {
		return
	}

	// Rollback to savepoint first
	tt.tx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", tt.savepoint))
	tt.tx.Rollback()
	tt.tx = nil
}

// Commit commits the transaction (use sparingly in tests).
func (tt *TestTransaction) Commit() error {
	if tt.tx == nil {
		return fmt.Errorf("transaction already completed")
	}

	err := tt.tx.Commit().Error
	tt.tx = nil
	return err
}

// TruncateTables truncates specified tables in the database.
// Use between test suites for complete isolation.
func TruncateTables(t *testing.T, db *gorm.DB, tables ...string) {
	t.Helper()

	testDBMu.Lock()
	defer testDBMu.Unlock()

	if len(tables) == 0 {
		// Truncate all tables except migrations
		tables = []string{
			"data_exports",
			"email_templates",
			"user_dismissed_announcements",
			"user_announcement_reads",
			"announcement_banners",
			"login_histories",
			"ip_blocklists",
			"system_settings",
			"user_sessions",
			"user_two_factors",
			"user_preferences",
			"user_feature_flags",
			"feature_flags",
			"audit_logs",
			"user_api_keys",
			"files",
			"oauth_providers",
			"subscriptions",
			"organization_invitations",
			"organization_members",
			"organizations",
			"token_blacklist",
			"users",
		}
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			// Ignore errors for non-existent tables
			t.Logf("Warning: Could not truncate table %s: %v", table, err)
		}
	}

	// Reset sequence counters
	ResetSequence()
}

// CleanupTestData removes test data by pattern.
func CleanupTestData(t *testing.T, db *gorm.DB, emailPattern string) {
	t.Helper()

	// Clean users with test email pattern
	db.Exec("DELETE FROM users WHERE email LIKE ?", emailPattern+"%")
	db.Exec("DELETE FROM organizations WHERE name LIKE ?", "Test%")
}

// WaitForDB waits for the database to be ready.
func WaitForDB(ctx context.Context, db *gorm.DB, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		sqlDB, err := db.DB()
		if err == nil {
			if err = sqlDB.PingContext(ctx); err == nil {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}

	return fmt.Errorf("database not ready within %v", timeout)
}

// WithTestDB is a helper that sets up a test database and runs a test function.
// It handles transaction rollback automatically.
func WithTestDB(t *testing.T, fn func(db *gorm.DB)) {
	t.Helper()
	SkipIfNotIntegration(t)

	db := SetupTestDB(t)
	tx := NewTestTransaction(t, db)

	fn(tx.DB)
	// Transaction is automatically rolled back by cleanup
}

// CreateTestUser is a helper to quickly create a user in the database for testing.
func CreateTestUser(t *testing.T, db *gorm.DB, opts ...func(*UserFactory)) *models.User {
	t.Helper()

	factory := NewUserFactory()
	for _, opt := range opts {
		opt(factory)
	}

	user := factory.Build()
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// CreateTestOrganization is a helper to quickly create an organization in the database.
func CreateTestOrganization(t *testing.T, db *gorm.DB, name string, ownerID uint) *models.Organization {
	t.Helper()

	org := &models.Organization{
		Name:            name,
		Slug:            fmt.Sprintf("test-org-%d", time.Now().UnixNano()),
		Plan:            models.OrgPlanFree,
		CreatedByUserID: ownerID,
	}

	if err := db.Create(org).Error; err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}

	return org
}

// CreateTestOrgMember is a helper to add a member to an organization.
func CreateTestOrgMember(t *testing.T, db *gorm.DB, orgID, userID uint, role models.OrganizationRole) *models.OrganizationMember {
	t.Helper()

	member := &models.OrganizationMember{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
		Status:         models.MemberStatusActive,
	}

	if err := db.Create(member).Error; err != nil {
		t.Fatalf("Failed to create organization member: %v", err)
	}

	return member
}
