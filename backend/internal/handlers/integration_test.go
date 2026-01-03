// Package handlers contains integration tests that test handlers with a real database.
//
// To run these tests, set the environment variable:
//
//	INTEGRATION_TEST=true go test -v -run Integration ./internal/handlers/...
//
// These tests require a running PostgreSQL database configured via environment variables.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// skipIfNotIntegration skips the test if INTEGRATION_TEST is not set.
func skipIfNotIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test (set INTEGRATION_TEST=true to run)")
	}
}

// testDB holds the test database connection.
var testDB *gorm.DB

// setupTestDB sets up a test database connection.
func setupTestDB(t *testing.T) *gorm.DB {
	skipIfNotIntegration(t)

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnvOrDefaultTest("TEST_DB_HOST", "localhost"),
		getEnvOrDefaultTest("TEST_DB_PORT", "5432"),
		getEnvOrDefaultTest("TEST_DB_USER", "postgres"),
		getEnvOrDefaultTest("TEST_DB_PASSWORD", "postgres"),
		getEnvOrDefaultTest("TEST_DB_NAME", "react_golang_starter_test"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to connect to test database")

	// Auto-migrate test tables
	err = db.AutoMigrate(&models.User{}, &models.TokenBlacklist{})
	require.NoError(t, err, "Failed to migrate test database")

	// Set global DB for handlers
	database.DB = db

	return db
}

// cleanupTestDB cleans up test data.
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	// Clean up test users
	db.Exec("DELETE FROM users WHERE email LIKE 'integration-test-%'")
	db.Exec("DELETE FROM token_blacklist WHERE reason = 'integration_test'")
}

// createTestUser creates a user for testing.
func createTestUser(t *testing.T, db *gorm.DB, email, password string) *models.User {
	hashedPassword, err := auth.HashPassword(password)
	require.NoError(t, err)

	user := &models.User{
		Name:          "Integration Test User",
		Email:         email,
		Password:      hashedPassword,
		EmailVerified: true,
		IsActive:      true,
		Role:          models.RoleUser,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = db.Create(user).Error
	require.NoError(t, err)

	return user
}

// getEnvOrDefault gets an environment variable or returns a default value.
func getEnvOrDefaultTest(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// TestIntegration_UserCRUD tests the full user CRUD lifecycle.
func TestIntegration_UserCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	email := fmt.Sprintf("integration-test-%d@example.com", time.Now().UnixNano())
	password := "TestPassword123!"

	t.Run("Create User via API", func(t *testing.T) {
		reqBody := models.RegisterRequest{
			Name:     "Integration Test User",
			Email:    email,
			Password: password,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := CreateUser()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var response models.SuccessResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("Get User by ID", func(t *testing.T) {
		// First get the user from DB
		var user models.User
		err := db.Where("email = ?", email).First(&user).Error
		require.NoError(t, err)

		// Create authenticated request
		token, err := auth.GenerateJWT(&user)
		require.NoError(t, err)

		// Use chi router to handle URL params
		r := chi.NewRouter()
		r.Get("/api/v1/users/{id}", GetUser())

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%d", user.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Add user to context
		ctx := auth.SetUserContext(req.Context(), &user)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response models.SuccessResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("Update User", func(t *testing.T) {
		var user models.User
		err := db.Where("email = ?", email).First(&user).Error
		require.NoError(t, err)

		token, err := auth.GenerateJWT(&user)
		require.NoError(t, err)

		updateData := map[string]string{
			"name": "Updated Integration Test User",
		}
		body, _ := json.Marshal(updateData)

		r := chi.NewRouter()
		r.Put("/api/v1/users/{id}", UpdateUser())

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/users/%d", user.ID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		ctx := auth.SetUserContext(req.Context(), &user)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify update in database
		var updatedUser models.User
		err = db.First(&updatedUser, user.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "Updated Integration Test User", updatedUser.Name)
	})

	t.Run("Delete User", func(t *testing.T) {
		var user models.User
		err := db.Where("email = ?", email).First(&user).Error
		require.NoError(t, err)

		token, err := auth.GenerateJWT(&user)
		require.NoError(t, err)

		r := chi.NewRouter()
		r.Delete("/api/v1/users/{id}", DeleteUser())

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d", user.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		ctx := auth.SetUserContext(req.Context(), &user)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify soft delete
		var deletedUser models.User
		err = db.Unscoped().First(&deletedUser, user.ID).Error
		require.NoError(t, err)
		assert.NotNil(t, deletedUser.DeletedAt)
	})
}

// TestIntegration_UserPagination tests pagination of user list.
func TestIntegration_UserPagination(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create test users
	for i := 0; i < 25; i++ {
		email := fmt.Sprintf("integration-test-pagination-%d-%d@example.com", time.Now().UnixNano(), i)
		createTestUser(t, db, email, "TestPassword123!")
	}

	t.Run("Paginated List - Page 1", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=1&limit=10", nil)
		rr := httptest.NewRecorder()

		handler := GetUsers()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response models.SuccessResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response.Data.(map[string]interface{})
		users := data["users"].([]interface{})
		assert.Len(t, users, 10)
		assert.GreaterOrEqual(t, int(data["total"].(float64)), 25)
	})

	t.Run("Paginated List - Page 2", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=2&limit=10", nil)
		rr := httptest.NewRecorder()

		handler := GetUsers()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response models.SuccessResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		data := response.Data.(map[string]interface{})
		users := data["users"].([]interface{})
		assert.Len(t, users, 10)
		assert.Equal(t, 2, int(data["page"].(float64)))
	})

	t.Run("Invalid Pagination Params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=-1&limit=10", nil)
		rr := httptest.NewRecorder()

		handler := GetUsers()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Limit Exceeds Maximum", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=1&limit=200", nil)
		rr := httptest.NewRecorder()

		handler := GetUsers()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestIntegration_AccountLockout tests the account lockout feature.
func TestIntegration_AccountLockout(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	email := fmt.Sprintf("integration-test-lockout-%d@example.com", time.Now().UnixNano())
	password := "TestPassword123!"
	createTestUser(t, db, email, password)

	t.Run("Account locks after failed attempts", func(t *testing.T) {
		// Make 5 failed login attempts
		for i := 0; i < 5; i++ {
			reqBody := models.LoginRequest{
				Email:    email,
				Password: "WrongPassword123!",
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			auth.LoginUser(rr, req)

			// First 4 attempts should return 401
			if i < 4 {
				assert.Equal(t, http.StatusUnauthorized, rr.Code, "Attempt %d should return 401", i+1)
			}
		}

		// 6th attempt should be locked
		reqBody := models.LoginRequest{
			Email:    email,
			Password: "WrongPassword123!",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		auth.LoginUser(rr, req)

		assert.Equal(t, http.StatusTooManyRequests, rr.Code, "Should return 429 when account is locked")

		// Verify account is locked in database
		var user models.User
		err := db.Where("email = ?", email).First(&user).Error
		require.NoError(t, err)
		assert.NotNil(t, user.LockedUntil)
		assert.GreaterOrEqual(t, user.FailedLoginAttempts, 5)
	})

	t.Run("Successful login resets counter", func(t *testing.T) {
		// Create a fresh user for this test
		freshEmail := fmt.Sprintf("integration-test-reset-%d@example.com", time.Now().UnixNano())
		createTestUser(t, db, freshEmail, password)

		// Make 2 failed attempts
		for i := 0; i < 2; i++ {
			reqBody := models.LoginRequest{
				Email:    freshEmail,
				Password: "WrongPassword123!",
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			auth.LoginUser(rr, req)
		}

		// Verify failed attempts recorded
		var user models.User
		err := db.Where("email = ?", freshEmail).First(&user).Error
		require.NoError(t, err)
		assert.Equal(t, 2, user.FailedLoginAttempts)

		// Successful login
		reqBody := models.LoginRequest{
			Email:    freshEmail,
			Password: password,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		auth.LoginUser(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Verify counter is reset
		err = db.Where("email = ?", freshEmail).First(&user).Error
		require.NoError(t, err)
		assert.Equal(t, 0, user.FailedLoginAttempts)
		assert.Nil(t, user.LockedUntil)
	})
}

// TestIntegration_HealthCheck tests the health check endpoint with real dependencies.
func TestIntegration_HealthCheck(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	service := NewService()

	t.Run("Health check returns healthy", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		rr := httptest.NewRecorder()

		service.HealthCheck(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response models.HealthStatus
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "healthy", response.OverallStatus)
	})

	t.Run("Health check with verbose mode", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health?verbose=true", nil)
		rr := httptest.NewRecorder()

		service.HealthCheck(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response models.HealthStatus
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotNil(t, response.Runtime)
		assert.Greater(t, response.Runtime.Goroutines, 0)
	})
}

// TestIntegration_N1Detection tests N+1 query detection.
func TestIntegration_N1Detection(t *testing.T) {
	skipIfNotIntegration(t)

	// This test verifies that N+1 detection is working by checking
	// that queries are tracked when N1DetectionEnabled is true

	t.Run("N+1 tracker is created in development context", func(t *testing.T) {
		ctx := context.Background()
		ctx = database.WithN1Detection(ctx)

		tracker := database.GetN1Tracker(ctx)
		assert.NotNil(t, tracker)
	})

	t.Run("Tracker detects repeated queries", func(t *testing.T) {
		ctx := context.Background()
		ctx = database.WithN1Detection(ctx)

		tracker := database.GetN1Tracker(ctx)
		require.NotNil(t, tracker)

		// Simulate repeated queries (would happen in N+1 scenario)
		for i := 0; i < 10; i++ {
			tracker.Track(fmt.Sprintf("SELECT * FROM users WHERE id = %d", i))
		}

		violations := tracker.GetN1Violations(5)
		assert.Contains(t, violations, "users")
		assert.GreaterOrEqual(t, violations["users"], 10)
	})
}
