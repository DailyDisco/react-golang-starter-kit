package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- GetAllowedOrigins Tests ---

func TestGetAllowedOrigins_FromEnvironment(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com,https://api.example.com")

	origins := GetAllowedOrigins()

	assert.Equal(t, []string{"https://example.com", "https://api.example.com"}, origins)
}

func TestGetAllowedOrigins_Default(t *testing.T) {
	// Ensure env var is not set
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	origins := GetAllowedOrigins()

	assert.Equal(t, DefaultAllowedOrigins, origins)
	assert.Contains(t, origins, "http://localhost:3000")
	assert.Contains(t, origins, "http://localhost:5173")
}

// --- Config.Validate Tests ---

func TestConfigValidate_MissingJWTSecret(t *testing.T) {
	cfg := &Config{
		JWT: JWTConfig{
			Secret: "",
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Name: "testdb",
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET is required")
}

func TestConfigValidate_MissingDatabaseHost(t *testing.T) {
	cfg := &Config{
		JWT: JWTConfig{
			Secret: "test-secret",
		},
		Database: DatabaseConfig{
			Host: "",
			Name: "testdb",
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database host is required")
}

func TestConfigValidate_MissingDatabaseName(t *testing.T) {
	cfg := &Config{
		JWT: JWTConfig{
			Secret: "test-secret",
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Name: "",
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database name is required")
}

func TestConfigValidate_InvalidLogLevel(t *testing.T) {
	cfg := &Config{
		JWT: JWTConfig{
			Secret: "test-secret",
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Name: "testdb",
		},
		Logging: LoggingConfig{
			Level: "invalid",
		},
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestConfigValidate_ValidLogLevels(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warn level", "warn"},
		{"error level", "error"},
		{"fatal level", "fatal"},
		{"uppercase DEBUG", "DEBUG"},
		{"mixed case Info", "Info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				JWT: JWTConfig{
					Secret: "test-secret",
				},
				Database: DatabaseConfig{
					Host: "localhost",
					Name: "testdb",
				},
				Logging: LoggingConfig{
					Level: tt.level,
				},
			}

			err := cfg.Validate()

			assert.NoError(t, err)
		})
	}
}

func TestConfigValidate_AllFieldsValid(t *testing.T) {
	cfg := &Config{
		JWT: JWTConfig{
			Secret: "test-secret-key",
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Name: "testdb",
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}

	err := cfg.Validate()

	assert.NoError(t, err)
}

func TestConfigValidate_MultipleErrors(t *testing.T) {
	cfg := &Config{
		JWT: JWTConfig{
			Secret: "",
		},
		Database: DatabaseConfig{
			Host: "",
			Name: "",
		},
		Logging: LoggingConfig{
			Level: "invalid",
		},
	}

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET is required")
	assert.Contains(t, err.Error(), "database host is required")
	assert.Contains(t, err.Error(), "database name is required")
	assert.Contains(t, err.Error(), "invalid log level")
}

// --- Helper Function Tests ---

func TestGetEnv_WithValue(t *testing.T) {
	t.Setenv("TEST_KEY", "test_value")

	result := getEnv("TEST_KEY", "fallback")

	assert.Equal(t, "test_value", result)
}

func TestGetEnv_Fallback(t *testing.T) {
	os.Unsetenv("NONEXISTENT_KEY")

	result := getEnv("NONEXISTENT_KEY", "fallback_value")

	assert.Equal(t, "fallback_value", result)
}

func TestGetEnv_EmptyStringUsesEnv(t *testing.T) {
	// Empty string is still a value, but getEnv treats it as not set
	t.Setenv("EMPTY_KEY", "")

	result := getEnv("EMPTY_KEY", "fallback")

	// Empty string is treated as unset
	assert.Equal(t, "fallback", result)
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name     string
		envVal   string
		fallback bool
		want     bool
	}{
		{"true string", "true", false, true},
		{"false string", "false", true, false},
		{"1 as true", "1", false, true},
		{"0 as false", "0", true, false},
		{"TRUE uppercase", "TRUE", false, true},
		{"FALSE uppercase", "FALSE", true, false},
		{"invalid uses fallback true", "notabool", true, true},
		{"invalid uses fallback false", "notabool", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TEST_BOOL_KEY", tt.envVal)

			got := getEnvBool("TEST_BOOL_KEY", tt.fallback)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetEnvBool_Unset(t *testing.T) {
	os.Unsetenv("UNSET_BOOL_KEY")

	result := getEnvBool("UNSET_BOOL_KEY", true)

	assert.True(t, result)
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name     string
		envVal   string
		fallback int
		want     int
	}{
		{"positive number", "42", 0, 42},
		{"negative number", "-10", 0, -10},
		{"zero", "0", 5, 0},
		{"invalid uses fallback", "notanumber", 99, 99},
		{"float truncates uses fallback", "3.14", 0, 0}, // strconv.Atoi fails on floats
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TEST_INT_KEY", tt.envVal)

			got := getEnvInt("TEST_INT_KEY", tt.fallback)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetEnvInt_Unset(t *testing.T) {
	os.Unsetenv("UNSET_INT_KEY")

	result := getEnvInt("UNSET_INT_KEY", 123)

	assert.Equal(t, 123, result)
}

func TestGetEnvInt64(t *testing.T) {
	tests := []struct {
		name     string
		envVal   string
		fallback int64
		want     int64
	}{
		{"large positive", "9223372036854775807", 0, 9223372036854775807},
		{"large negative", "-9223372036854775808", 0, -9223372036854775808},
		{"regular number", "1000000", 0, 1000000},
		{"invalid uses fallback", "invalid", 42, 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TEST_INT64_KEY", tt.envVal)

			got := getEnvInt64("TEST_INT64_KEY", tt.fallback)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetEnvFloat64(t *testing.T) {
	tests := []struct {
		name     string
		envVal   string
		fallback float64
		want     float64
	}{
		{"decimal", "3.14159", 0, 3.14159},
		{"negative decimal", "-2.5", 0, -2.5},
		{"integer as float", "42", 0, 42.0},
		{"scientific notation", "1.5e10", 0, 1.5e10},
		{"invalid uses fallback", "notafloat", 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TEST_FLOAT_KEY", tt.envVal)

			got := getEnvFloat64("TEST_FLOAT_KEY", tt.fallback)

			assert.InDelta(t, tt.want, got, 0.0001)
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name     string
		envVal   string
		fallback time.Duration
		want     time.Duration
	}{
		{"seconds", "30s", 0, 30 * time.Second},
		{"minutes", "5m", 0, 5 * time.Minute},
		{"hours", "2h", 0, 2 * time.Hour},
		{"milliseconds", "500ms", 0, 500 * time.Millisecond},
		{"combined", "1h30m", 0, 90 * time.Minute},
		{"invalid uses fallback", "invalid", 10 * time.Second, 10 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TEST_DURATION_KEY", tt.envVal)

			got := getEnvDuration("TEST_DURATION_KEY", tt.fallback)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetEnvDuration_Unset(t *testing.T) {
	os.Unsetenv("UNSET_DURATION_KEY")

	result := getEnvDuration("UNSET_DURATION_KEY", 5*time.Minute)

	assert.Equal(t, 5*time.Minute, result)
}

func TestGetEnvSlice(t *testing.T) {
	tests := []struct {
		name     string
		envVal   string
		fallback []string
		want     []string
	}{
		{"comma separated", "a,b,c", nil, []string{"a", "b", "c"}},
		{"single value", "only", nil, []string{"only"}},
		{"with spaces", "one, two, three", nil, []string{"one", " two", " three"}}, // Note: spaces preserved
		{"empty uses fallback", "", []string{"default"}, []string{"default"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				t.Setenv("TEST_SLICE_KEY", tt.envVal)
			} else {
				os.Unsetenv("TEST_SLICE_KEY")
			}

			got := getEnvSlice("TEST_SLICE_KEY", tt.fallback)

			assert.Equal(t, tt.want, got)
		})
	}
}

// --- Config Method Tests ---

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"lowercase development", "development", true},
		{"uppercase DEVELOPMENT", "DEVELOPMENT", true},
		{"mixed case Development", "Development", true},
		{"production", "production", false},
		{"staging", "staging", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Env: tt.env,
				},
			}

			got := cfg.IsDevelopment()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"lowercase production", "production", true},
		{"uppercase PRODUCTION", "PRODUCTION", true},
		{"mixed case Production", "Production", true},
		{"development", "development", false},
		{"staging", "staging", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Env: tt.env,
				},
			}

			got := cfg.IsProduction()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfig_GetDatabaseDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "db.example.com",
			Port:     "5432",
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
			SSLMode:  "require",
		},
	}

	dsn := cfg.GetDatabaseDSN()

	assert.Equal(t, "host=db.example.com port=5432 user=testuser password=testpass dbname=testdb sslmode=require", dsn)
}

func TestConfig_GetDatabaseDSN_WithSpecialChars(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "user",
			Password: "p@ss=word!",
			Name:     "my-db",
			SSLMode:  "disable",
		},
	}

	dsn := cfg.GetDatabaseDSN()

	assert.Contains(t, dsn, "password=p@ss=word!")
	assert.Contains(t, dsn, "dbname=my-db")
}

func TestConfig_GetServerAddr(t *testing.T) {
	tests := []struct {
		name string
		host string
		port string
		want string
	}{
		{"standard", "0.0.0.0", "8080", "0.0.0.0:8080"},
		{"localhost", "localhost", "3000", "localhost:3000"},
		{"ip address", "192.168.1.1", "9000", "192.168.1.1:9000"},
		{"empty host", "", "8080", ":8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Host: tt.host,
					Port: tt.port,
				},
			}

			got := cfg.GetServerAddr()

			assert.Equal(t, tt.want, got)
		})
	}
}

// --- Load Function Tests ---

func TestLoad_WithValidEnv(t *testing.T) {
	// Set minimum required env vars
	t.Setenv("JWT_SECRET", "test-jwt-secret-key")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_NAME", "testdb")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test-jwt-secret-key", cfg.JWT.Secret)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "debug", cfg.Logging.Level)
}

func TestLoad_ValidationError(t *testing.T) {
	// Clear all env vars that might have secrets
	os.Unsetenv("JWT_SECRET")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_NAME", "testdb")

	cfg, err := Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "JWT_SECRET is required")
}

func TestLoad_DefaultValues(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	// Don't set DB_HOST or DB_NAME, let them use defaults
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("PGHOST")
	os.Unsetenv("PGDATABASE")

	cfg, err := Load()

	require.NoError(t, err)
	// Check defaults are applied
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "development", cfg.Server.Env)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "starter_kit_db", cfg.Database.Name)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, int64(10), cfg.FileUpload.MaxFileSizeMB)
}

func TestLoad_PostgresEnvVarPrecedence(t *testing.T) {
	// PGHOST/PGPORT should take precedence over DB_HOST/DB_PORT
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("DB_HOST", "db-host")
	t.Setenv("PGHOST", "pg-host")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("PGPORT", "5434")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "pg-host", cfg.Database.Host)
	assert.Equal(t, "5434", cfg.Database.Port)
}

func TestLoad_RateLimitDefaults(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	cfg, err := Load()

	require.NoError(t, err)
	assert.True(t, cfg.RateLimit.Enabled)
	assert.Equal(t, 60, cfg.RateLimit.IPRequestsPerMinute)
	assert.Equal(t, 1000, cfg.RateLimit.IPRequestsPerHour)
	assert.Equal(t, 5, cfg.RateLimit.AuthRequestsPerMinute)
}

func TestLoad_CustomRateLimits(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("RATE_LIMIT_ENABLED", "false")
	t.Setenv("RATE_LIMIT_IP_PER_MINUTE", "120")
	t.Setenv("RATE_LIMIT_AUTH_PER_MINUTE", "10")

	cfg, err := Load()

	require.NoError(t, err)
	assert.False(t, cfg.RateLimit.Enabled)
	assert.Equal(t, 120, cfg.RateLimit.IPRequestsPerMinute)
	assert.Equal(t, 10, cfg.RateLimit.AuthRequestsPerMinute)
}
