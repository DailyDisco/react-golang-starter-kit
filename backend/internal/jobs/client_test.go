package jobs

import (
	"context"
	"strings"
	"testing"
)

// ============ getEnv Tests ============

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback string
		envValue string
		setEnv   bool
		expected string
	}{
		{"returns env value when set", "TEST_GET_ENV_1", "default", "custom", true, "custom"},
		{"returns fallback when not set", "TEST_GET_ENV_2", "fallback", "", false, "fallback"},
		{"returns fallback when empty", "TEST_GET_ENV_3", "fallback", "", true, "fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}
			result := getEnv(tt.key, tt.fallback)
			if result != tt.expected {
				t.Errorf("getEnv(%q, %q) = %q, want %q", tt.key, tt.fallback, result, tt.expected)
			}
		})
	}
}

// ============ buildDatabaseURL Tests ============

func TestBuildDatabaseURL_Defaults(t *testing.T) {
	// Clear all database env vars to test defaults
	t.Setenv("PGHOST", "")
	t.Setenv("PGPORT", "")
	t.Setenv("PGUSER", "")
	t.Setenv("PGPASSWORD", "")
	t.Setenv("PGDATABASE", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_SSLMODE", "")

	url := buildDatabaseURL()

	// Should use default values
	expected := "postgres://devuser:devpass@localhost:5432/starter_kit_db?sslmode=disable"
	if url != expected {
		t.Errorf("buildDatabaseURL() = %q, want %q", url, expected)
	}
}

func TestBuildDatabaseURL_DBVars(t *testing.T) {
	// Clear PG* vars
	t.Setenv("PGHOST", "")
	t.Setenv("PGPORT", "")
	t.Setenv("PGUSER", "")
	t.Setenv("PGPASSWORD", "")
	t.Setenv("PGDATABASE", "")

	// Set DB_* vars
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "testuser")
	t.Setenv("DB_PASSWORD", "testpass")
	t.Setenv("DB_NAME", "testdb")
	t.Setenv("DB_SSLMODE", "require")

	url := buildDatabaseURL()

	expected := "postgres://testuser:testpass@db.example.com:5433/testdb?sslmode=require"
	if url != expected {
		t.Errorf("buildDatabaseURL() = %q, want %q", url, expected)
	}
}

func TestBuildDatabaseURL_PGVars(t *testing.T) {
	// Set PG* vars (should take precedence)
	t.Setenv("PGHOST", "pg.example.com")
	t.Setenv("PGPORT", "5434")
	t.Setenv("PGUSER", "pguser")
	t.Setenv("PGPASSWORD", "pgpass")
	t.Setenv("PGDATABASE", "pgdb")
	t.Setenv("DB_SSLMODE", "verify-full")

	url := buildDatabaseURL()

	expected := "postgres://pguser:pgpass@pg.example.com:5434/pgdb?sslmode=verify-full"
	if url != expected {
		t.Errorf("buildDatabaseURL() = %q, want %q", url, expected)
	}
}

func TestBuildDatabaseURL_Format(t *testing.T) {
	t.Setenv("DB_HOST", "myhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "myuser")
	t.Setenv("DB_PASSWORD", "mypass")
	t.Setenv("DB_NAME", "mydb")
	t.Setenv("DB_SSLMODE", "disable")

	// Clear PG* vars
	t.Setenv("PGHOST", "")
	t.Setenv("PGPORT", "")
	t.Setenv("PGUSER", "")
	t.Setenv("PGPASSWORD", "")
	t.Setenv("PGDATABASE", "")

	url := buildDatabaseURL()

	// Verify URL structure
	if !strings.HasPrefix(url, "postgres://") {
		t.Error("buildDatabaseURL() should start with postgres://")
	}
	if !strings.Contains(url, "myuser:mypass@") {
		t.Error("buildDatabaseURL() should contain user:password@")
	}
	if !strings.Contains(url, "myhost:5432") {
		t.Error("buildDatabaseURL() should contain host:port")
	}
	if !strings.Contains(url, "/mydb?") {
		t.Error("buildDatabaseURL() should contain /database?")
	}
	if !strings.Contains(url, "sslmode=disable") {
		t.Error("buildDatabaseURL() should contain sslmode parameter")
	}
}

// ============ IsAvailable Tests ============

func TestIsAvailable_NilInstance(t *testing.T) {
	// Ensure instance is nil
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	if IsAvailable() {
		t.Error("IsAvailable() should return false when instance is nil")
	}
}

func TestIsAvailable_NonNilInstance(t *testing.T) {
	oldInstance := instance
	instance = &Client{} // Non-nil instance
	defer func() { instance = oldInstance }()

	if !IsAvailable() {
		t.Error("IsAvailable() should return true when instance is not nil")
	}
}

// ============ Start Tests ============

func TestStart_NilInstance(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	err := Start(context.Background())
	if err != nil {
		t.Errorf("Start() with nil instance = %v, want nil", err)
	}
}

// ============ Stop Tests ============

func TestStop_NilInstance(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	err := Stop(context.Background())
	if err != nil {
		t.Errorf("Stop() with nil instance = %v, want nil", err)
	}
}

// ============ GetClient Tests ============

func TestGetClient_NilInstance(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	client := GetClient()
	if client != nil {
		t.Error("GetClient() should return nil when instance is nil")
	}
}

// ============ Insert Tests ============

// Note: Insert requires type parameters, so we test with a concrete type
type testJobArgs struct {
	Data string
}

func (t testJobArgs) Kind() string {
	return "test_job"
}

func TestInsert_NilInstance(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	args := testJobArgs{Data: "test"}
	err := Insert(context.Background(), args, nil)
	if err == nil {
		t.Error("Insert() with nil instance should return error")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("Insert() error = %q, should mention 'not initialized'", err.Error())
	}
}

// ============ InsertMany Tests ============

func TestInsertMany_NilInstance(t *testing.T) {
	oldInstance := instance
	instance = nil
	defer func() { instance = oldInstance }()

	err := InsertMany(context.Background(), nil)
	if err == nil {
		t.Error("InsertMany() with nil instance should return error")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("InsertMany() error = %q, should mention 'not initialized'", err.Error())
	}
}

// ============ Initialize Tests ============

func TestInitializeWithPool_Disabled(t *testing.T) {
	oldInstance := instance
	defer func() { instance = oldInstance }()

	config := &Config{Enabled: false}
	err := InitializeWithPool(config, nil)

	if err != nil {
		t.Errorf("InitializeWithPool() with disabled config = %v, want nil", err)
	}
	if instance != nil {
		t.Error("instance should be nil when jobs are disabled")
	}
}

func TestInitialize_Disabled(t *testing.T) {
	oldInstance := instance
	defer func() { instance = oldInstance }()

	config := &Config{Enabled: false}
	err := Initialize(config)

	if err != nil {
		t.Errorf("Initialize() with disabled config = %v, want nil", err)
	}
	if instance != nil {
		t.Error("instance should be nil when jobs are disabled")
	}
}

// ============ Client Structure Tests ============

func TestClient_Structure(t *testing.T) {
	client := &Client{
		river:    nil,
		pool:     nil,
		ownsPool: true,
		config:   &Config{WorkerCount: 5},
	}

	if !client.ownsPool {
		t.Error("ownsPool should be true")
	}
	if client.config.WorkerCount != 5 {
		t.Errorf("config.WorkerCount = %d, want 5", client.config.WorkerCount)
	}
}
