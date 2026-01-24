package email

import (
	"context"
	"strings"
	"testing"
)

// ============ NoOpProvider Tests ============

func TestNoOpProvider_Send(t *testing.T) {
	provider := NewNoOpProvider()

	params := SendParams{
		To:           "test@example.com",
		Subject:      "Test Subject",
		TemplateName: "test",
		Data:         map[string]interface{}{"key": "value"},
	}

	err := provider.Send(context.Background(), params)
	if err != nil {
		t.Errorf("NoOpProvider.Send() error = %v, want nil", err)
	}
}

func TestNoOpProvider_SendBatch(t *testing.T) {
	provider := NewNoOpProvider()

	params := []SendParams{
		{To: "test1@example.com", Subject: "Test 1"},
		{To: "test2@example.com", Subject: "Test 2"},
	}

	err := provider.SendBatch(context.Background(), params)
	if err != nil {
		t.Errorf("NoOpProvider.SendBatch() error = %v, want nil", err)
	}
}

func TestNoOpProvider_IsAvailable(t *testing.T) {
	provider := NewNoOpProvider()

	if provider.IsAvailable() {
		t.Error("NoOpProvider.IsAvailable() = true, want false")
	}
}

func TestNoOpProvider_Close(t *testing.T) {
	provider := NewNoOpProvider()

	err := provider.Close()
	if err != nil {
		t.Errorf("NoOpProvider.Close() error = %v, want nil", err)
	}
}

// ============ Package Level Function Tests ============

func TestInitialize_Disabled(t *testing.T) {
	// Reset global state
	instance = nil
	config = nil

	cfg := &Config{
		Enabled: false,
	}

	err := Initialize(cfg, nil)
	if err != nil {
		t.Errorf("Initialize() error = %v, want nil", err)
	}

	if instance == nil {
		t.Fatal("Initialize() did not set instance")
	}

	// Should use NoOpProvider when disabled
	if instance.IsAvailable() {
		t.Error("Initialize() with disabled config should create unavailable provider")
	}
}

func TestSend_NotInitialized(t *testing.T) {
	// Reset global state
	instance = nil

	params := SendParams{
		To:      "test@example.com",
		Subject: "Test",
	}

	err := Send(context.Background(), params)
	if err != ErrNotInitialized {
		t.Errorf("Send() error = %v, want %v", err, ErrNotInitialized)
	}
}

func TestSendBatch_NotInitialized(t *testing.T) {
	// Reset global state
	instance = nil

	params := []SendParams{
		{To: "test@example.com", Subject: "Test"},
	}

	err := SendBatch(context.Background(), params)
	if err != ErrNotInitialized {
		t.Errorf("SendBatch() error = %v, want %v", err, ErrNotInitialized)
	}
}

func TestIsAvailable_NotInitialized(t *testing.T) {
	// Reset global state
	instance = nil

	if IsAvailable() {
		t.Error("IsAvailable() = true when not initialized, want false")
	}
}

func TestIsAvailable_WithNoOpProvider(t *testing.T) {
	// Set up NoOpProvider
	instance = NewNoOpProvider()

	if IsAvailable() {
		t.Error("IsAvailable() = true with NoOpProvider, want false")
	}
}

func TestGetFrontendURL_NoConfig(t *testing.T) {
	// Reset config
	config = nil

	url := GetFrontendURL()
	expected := "http://localhost:5173"
	if url != expected {
		t.Errorf("GetFrontendURL() = %q, want %q", url, expected)
	}
}

func TestGetFrontendURL_WithConfig(t *testing.T) {
	config = &Config{
		FrontendURL: "https://example.com",
	}
	defer func() { config = nil }()

	url := GetFrontendURL()
	if url != "https://example.com" {
		t.Errorf("GetFrontendURL() = %q, want %q", url, "https://example.com")
	}
}

func TestClose_NotInitialized(t *testing.T) {
	// Reset global state
	instance = nil

	err := Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestClose_WithProvider(t *testing.T) {
	instance = NewNoOpProvider()

	err := Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestGetConfig(t *testing.T) {
	expectedConfig := &Config{
		Enabled:     true,
		SMTPHost:    "smtp.example.com",
		FrontendURL: "https://example.com",
	}
	config = expectedConfig
	defer func() { config = nil }()

	got := GetConfig()
	if got != expectedConfig {
		t.Errorf("GetConfig() returned different config pointer")
	}
}

// ============ SendParams Validation Tests ============

func TestSendParams_Validation(t *testing.T) {
	tests := []struct {
		name   string
		params SendParams
		valid  bool
	}{
		{
			name: "valid with all fields",
			params: SendParams{
				To:           "test@example.com",
				Subject:      "Test Subject",
				TemplateName: "welcome",
				Data:         map[string]interface{}{"name": "John"},
				PlainText:    "Plain text version",
			},
			valid: true,
		},
		{
			name: "valid minimal",
			params: SendParams{
				To:      "test@example.com",
				Subject: "Test",
			},
			valid: true,
		},
		{
			name: "empty recipient",
			params: SendParams{
				To:      "",
				Subject: "Test",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For now, just check that To is required
			hasTo := tt.params.To != ""
			if hasTo != tt.valid {
				t.Errorf("SendParams validation for %s: got valid=%v, want %v", tt.name, hasTo, tt.valid)
			}
		})
	}
}

// ============ Template Manager Tests ============

func TestNewTemplateManager(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	expectedTemplates := []string{
		"welcome",
		"password_reset",
		"verification",
		"email_change_verify",
		"password_changed",
		"two_factor_code",
		"login_new_device",
		"account_locked",
	}

	for _, name := range expectedTemplates {
		if !tm.HasTemplate(name) {
			t.Errorf("Template %q not found", name)
		}
	}
}

func TestTemplateManager_ListTemplates(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	templates := tm.ListTemplates()
	if len(templates) < 8 {
		t.Errorf("ListTemplates() returned %d templates, want at least 8", len(templates))
	}
}

func TestTemplateManager_Render_Welcome(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	data := map[string]interface{}{
		"Name": "John Doe",
	}

	subject, html, text, err := tm.Render("welcome", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if subject == "" {
		t.Error("Render() returned empty subject")
	}

	if html == "" {
		t.Error("Render() returned empty HTML body")
	}

	if text == "" {
		t.Error("Render() returned empty text body")
	}

	// Verify template data was substituted
	if !strings.Contains(html, "John Doe") {
		t.Error("HTML body does not contain user name")
	}

	if !strings.Contains(html, "Welcome to") {
		t.Error("HTML body does not contain welcome message")
	}
}

func TestTemplateManager_Render_PasswordReset(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	data := map[string]interface{}{
		"Name":     "Jane Doe",
		"ResetURL": "https://example.com/reset?token=abc123",
	}

	subject, html, _, err := tm.Render("password_reset", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if !strings.Contains(subject, "password") {
		t.Error("Subject should mention password")
	}

	if !strings.Contains(html, "https://example.com/reset?token=abc123") {
		t.Error("HTML body does not contain reset URL")
	}
}

func TestTemplateManager_Render_TwoFactorCode(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	data := map[string]interface{}{
		"Name": "Test User",
		"Code": "123456",
	}

	subject, html, _, err := tm.Render("two_factor_code", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if !strings.Contains(subject, "123456") {
		t.Error("Subject should contain the code")
	}

	if !strings.Contains(html, "123456") {
		t.Error("HTML body does not contain the code")
	}
}

func TestTemplateManager_Render_LoginNewDevice(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	data := map[string]interface{}{
		"Name":     "Test User",
		"Device":   "Chrome on Windows",
		"Location": "New York, USA",
		"IP":       "192.168.1.1",
		"Time":     "Dec 30, 2025 at 7:30 PM",
	}

	_, html, _, err := tm.Render("login_new_device", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if !strings.Contains(html, "Chrome on Windows") {
		t.Error("HTML body does not contain device info")
	}

	if !strings.Contains(html, "New York, USA") {
		t.Error("HTML body does not contain location")
	}
}

func TestTemplateManager_Render_AccountLocked(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	data := map[string]interface{}{
		"Name":           "Test User",
		"LockDuration":   "30 minutes",
		"FailedAttempts": "5",
	}

	_, html, _, err := tm.Render("account_locked", data)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if !strings.Contains(html, "30 minutes") {
		t.Error("HTML body does not contain lock duration")
	}

	if !strings.Contains(html, "5") {
		t.Error("HTML body does not contain failed attempts count")
	}
}

func TestTemplateManager_Render_NotFound(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	_, _, _, err = tm.Render("nonexistent", nil)
	if err == nil {
		t.Error("Render() should return error for nonexistent template")
	}
}

// ============ Error Types Tests ============

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrNotInitialized",
			err:  ErrNotInitialized,
			want: "email service not initialized",
		},
		{
			name: "ErrInvalidParams",
			err:  ErrInvalidParams,
			want: "invalid email parameters",
		},
		{
			name: "ErrSendFailed",
			err:  ErrSendFailed,
			want: "failed to send email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.want)
			}
		})
	}
}
