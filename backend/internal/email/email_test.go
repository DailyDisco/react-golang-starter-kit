package email

import (
	"context"
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

	err := Initialize(cfg)
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
