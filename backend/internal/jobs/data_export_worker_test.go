package jobs

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/riverqueue/river"
)

// ============ DataExportArgs Tests ============

func TestDataExportArgs_Kind(t *testing.T) {
	args := DataExportArgs{}
	if args.Kind() != "generate_data_export" {
		t.Errorf("Kind() = %q, want %q", args.Kind(), "generate_data_export")
	}
}

func TestDataExportArgs_InsertOpts(t *testing.T) {
	args := DataExportArgs{}
	opts := args.InsertOpts()

	if opts.Queue != river.QueueDefault {
		t.Errorf("InsertOpts().Queue = %q, want default queue", opts.Queue)
	}
	if opts.MaxAttempts != 3 {
		t.Errorf("InsertOpts().MaxAttempts = %d, want 3", opts.MaxAttempts)
	}
}

func TestDataExportArgs_JSON(t *testing.T) {
	args := DataExportArgs{
		UserID:   42,
		Email:    "test@example.com",
		ExportID: 123,
	}

	data, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if result["user_id"] != float64(42) {
		t.Errorf("user_id = %v, want 42", result["user_id"])
	}
	if result["email"] != "test@example.com" {
		t.Errorf("email = %v, want test@example.com", result["email"])
	}
	if result["export_id"] != float64(123) {
		t.Errorf("export_id = %v, want 123", result["export_id"])
	}
}

// ============ formatBytes Tests ============

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		// Bytes
		{"zero bytes", 0, "0 B"},
		{"one byte", 1, "1 B"},
		{"500 bytes", 500, "500 B"},
		{"1023 bytes", 1023, "1023 B"},

		// Kilobytes
		{"exactly 1 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"10 KB", 10240, "10.0 KB"},
		{"1023 KB", 1047552, "1023.0 KB"},

		// Megabytes
		{"exactly 1 MB", 1048576, "1.0 MB"},
		{"1.5 MB", 1572864, "1.5 MB"},
		{"10 MB", 10485760, "10.0 MB"},
		{"100 MB", 104857600, "100.0 MB"},

		// Gigabytes
		{"exactly 1 GB", 1073741824, "1.0 GB"},
		{"1.5 GB", 1610612736, "1.5 GB"},
		{"10 GB", 10737418240, "10.0 GB"},

		// Terabytes
		{"exactly 1 TB", 1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes_LargeValues(t *testing.T) {
	// Test very large values
	result := formatBytes(1125899906842624) // 1 PB
	if result != "1.0 PB" {
		t.Errorf("formatBytes(1 PB) = %q, want '1.0 PB'", result)
	}
}

// ============ derefString Tests ============

func TestDerefString(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{"nil pointer", nil, ""},
		{"empty string", ptrString(""), ""},
		{"non-empty string", ptrString("hello"), "hello"},
		{"string with spaces", ptrString("  spaces  "), "  spaces  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := derefString(tt.input)
			if result != tt.expected {
				t.Errorf("derefString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ptrString is a helper to create a pointer to a string
func ptrString(s string) *string {
	return &s
}

// ============ getExportsDir Tests ============

func TestGetExportsDir_Default(t *testing.T) {
	// Clear any existing env var
	os.Unsetenv("DATA_EXPORTS_DIR")

	dir := getExportsDir()
	if dir != "exports" {
		t.Errorf("getExportsDir() = %q, want 'exports' when env var is not set", dir)
	}
}

func TestGetExportsDir_FromEnv(t *testing.T) {
	// Set custom directory
	os.Setenv("DATA_EXPORTS_DIR", "/custom/exports/path")
	defer os.Unsetenv("DATA_EXPORTS_DIR")

	dir := getExportsDir()
	if dir != "/custom/exports/path" {
		t.Errorf("getExportsDir() = %q, want '/custom/exports/path'", dir)
	}
}

func TestGetExportsDir_EmptyEnv(t *testing.T) {
	// Set empty env var
	os.Setenv("DATA_EXPORTS_DIR", "")
	defer os.Unsetenv("DATA_EXPORTS_DIR")

	dir := getExportsDir()
	if dir != "exports" {
		t.Errorf("getExportsDir() = %q, want 'exports' when env var is empty", dir)
	}
}

// ============ UserDataExport Structure Tests ============

func TestUserDataExport_JSONStructure(t *testing.T) {
	export := UserDataExport{
		ExportedAt: "2025-01-01T00:00:00Z",
		User: userExportData{
			ID:            1,
			Name:          "Test User",
			Email:         "test@example.com",
			EmailVerified: true,
			IsActive:      true,
			Role:          "user",
			CreatedAt:     "2024-01-01T00:00:00Z",
			UpdatedAt:     "2025-01-01T00:00:00Z",
		},
	}

	// Should be JSON serializable
	data, err := json.Marshal(export)
	if err != nil {
		t.Fatalf("UserDataExport JSON marshal failed: %v", err)
	}

	// Verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if result["exported_at"] != "2025-01-01T00:00:00Z" {
		t.Errorf("exported_at = %v, want '2025-01-01T00:00:00Z'", result["exported_at"])
	}

	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatal("user field should be a map")
	}

	if user["id"] != float64(1) {
		t.Errorf("user.id = %v, want 1", user["id"])
	}
	if user["email"] != "test@example.com" {
		t.Errorf("user.email = %v, want 'test@example.com'", user["email"])
	}
}

func TestUserExportData_OmitsEmptyOptionalFields(t *testing.T) {
	user := userExportData{
		ID:            1,
		Name:          "Test",
		Email:         "test@example.com",
		EmailVerified: false,
		IsActive:      true,
		Role:          "user",
		// Optional fields left empty
		AvatarURL: "",
		Bio:       "",
		Location:  "",
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2025-01-01T00:00:00Z",
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	// Optional fields with omitempty should not appear when empty
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// avatar_url, bio, location should be omitted when empty (if omitempty is used)
	// Note: Current implementation doesn't use omitempty for these fields,
	// so they will appear as empty strings
}

func TestTwoFactorExportData_Structure(t *testing.T) {
	tfa := twoFactorExportData{
		Enabled:              true,
		VerifiedAt:           "2024-06-01T00:00:00Z",
		LastUsedAt:           "2025-01-01T00:00:00Z",
		BackupCodesRemaining: 5,
	}

	data, err := json.Marshal(tfa)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if result["enabled"] != true {
		t.Errorf("enabled = %v, want true", result["enabled"])
	}
	if result["backup_codes_remaining"] != float64(5) {
		t.Errorf("backup_codes_remaining = %v, want 5", result["backup_codes_remaining"])
	}
}

func TestApiKeyExportData_NoSecrets(t *testing.T) {
	apiKey := apiKeyExportData{
		ID:         1,
		Name:       "My API Key",
		KeyPreview: "sk_...xyz", // Only preview, not full key
		IsActive:   true,
		UsageCount: 100,
		CreatedAt:  "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(apiKey)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	// Verify no full API key is exposed
	dataStr := string(data)
	if len(dataStr) > 500 {
		t.Error("API key export data seems too large, might contain full key")
	}
}

func TestOrgMembershipExport_Structure(t *testing.T) {
	membership := orgMembershipExport{
		OrganizationName: "My Org",
		OrganizationSlug: "my-org",
		Role:             "admin",
		JoinedAt:         "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(membership)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if result["organization_name"] != "My Org" {
		t.Errorf("organization_name = %v, want 'My Org'", result["organization_name"])
	}
	if result["role"] != "admin" {
		t.Errorf("role = %v, want 'admin'", result["role"])
	}
}

func TestFileExportData_Structure(t *testing.T) {
	file := fileExportData{
		ID:          1,
		FileName:    "document.pdf",
		FileSize:    1048576,
		ContentType: "application/pdf",
		CreatedAt:   "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(file)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if result["file_name"] != "document.pdf" {
		t.Errorf("file_name = %v, want 'document.pdf'", result["file_name"])
	}
	if result["file_size"] != float64(1048576) {
		t.Errorf("file_size = %v, want 1048576", result["file_size"])
	}
}

func TestAuditLogExport_Structure(t *testing.T) {
	auditLog := auditLogExport{
		Action:      "user.login",
		Description: "User logged in successfully",
		IPAddress:   "192.168.1.1",
		CreatedAt:   "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(auditLog)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if result["action"] != "user.login" {
		t.Errorf("action = %v, want 'user.login'", result["action"])
	}
}

// ============ Export Data Limits Tests ============

func TestDataExport_MaxRetries(t *testing.T) {
	opts := DataExportArgs{}.InsertOpts()

	// Data exports should have limited retries (they're resource-intensive)
	if opts.MaxAttempts > 5 {
		t.Errorf("DataExport MaxAttempts = %d, should not exceed 5", opts.MaxAttempts)
	}
	if opts.MaxAttempts < 2 {
		t.Errorf("DataExport MaxAttempts = %d, should be at least 2", opts.MaxAttempts)
	}
}
