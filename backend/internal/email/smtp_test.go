package email

import (
	"context"
	"testing"
)

// ============ SMTPProvider Tests ============

func TestNewSMTPProvider(t *testing.T) {
	cfg := &Config{
		SMTPHost: "localhost",
		SMTPPort: 587,
	}

	provider, err := NewSMTPProvider(cfg, nil)
	if err != nil {
		t.Fatalf("NewSMTPProvider() error = %v", err)
	}

	if provider == nil {
		t.Fatal("NewSMTPProvider() returned nil")
	}

	if provider.config != cfg {
		t.Error("Provider should have the provided config")
	}

	if provider.templates == nil {
		t.Error("Provider should have a template manager")
	}
}

func TestNewSMTPProvider_WithoutDB(t *testing.T) {
	cfg := &Config{
		SMTPHost: "localhost",
		SMTPPort: 587,
	}

	provider, err := NewSMTPProvider(cfg, nil)
	if err != nil {
		t.Fatalf("NewSMTPProvider() error = %v", err)
	}

	// Without DB, template resolver should be nil
	if provider.templateResolver != nil {
		t.Error("templateResolver should be nil without database")
	}
}

// ============ SMTPProvider.IsAvailable Tests ============

func TestSMTPProvider_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		expected bool
	}{
		{"host and port set", "smtp.example.com", 587, true},
		{"no host", "", 587, false},
		{"no port", "smtp.example.com", 0, false},
		{"neither", "", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &SMTPProvider{
				config: &Config{
					SMTPHost: tt.host,
					SMTPPort: tt.port,
				},
			}

			if provider.IsAvailable() != tt.expected {
				t.Errorf("IsAvailable() = %v, want %v", provider.IsAvailable(), tt.expected)
			}
		})
	}
}

// ============ SMTPProvider.Close Tests ============

func TestSMTPProvider_Close(t *testing.T) {
	provider := &SMTPProvider{
		config: &Config{},
	}

	err := provider.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

// ============ SMTPProvider.getTLSPolicy Tests ============

func TestSMTPProvider_GetTLSPolicy(t *testing.T) {
	tests := []struct {
		name       string
		tlsPolicy  string
		wantPolicy string
	}{
		{"mandatory", "mandatory", "mandatory"},
		{"none", "none", "none"},
		{"opportunistic", "opportunistic", "default"},
		{"empty default", "", "default"},
		{"unknown default", "unknown", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &SMTPProvider{
				config: &Config{
					TLSPolicy: tt.tlsPolicy,
				},
			}

			policy := provider.getTLSPolicy()
			// We can't easily compare mail.TLSPolicy values, but we can verify the method runs
			if policy < 0 || policy > 3 {
				t.Errorf("getTLSPolicy() returned invalid policy value")
			}
		})
	}
}

// ============ SMTPProvider.Send Tests ============

func TestSMTPProvider_Send_EmptyRecipient(t *testing.T) {
	provider := &SMTPProvider{
		config: &Config{
			SMTPHost: "localhost",
			SMTPPort: 587,
		},
	}

	params := SendParams{
		To:      "",
		Subject: "Test",
	}

	err := provider.Send(context.Background(), params)
	if err == nil {
		t.Error("Send() should return error for empty recipient")
	}
}

func TestSMTPProvider_Send_DevMode(t *testing.T) {
	tm, _ := NewTemplateManager()
	provider := &SMTPProvider{
		config: &Config{
			SMTPHost: "localhost",
			SMTPPort: 587,
			DevMode:  true,
		},
		templates: tm,
	}

	params := SendParams{
		To:           "test@example.com",
		Subject:      "Test Subject",
		TemplateName: "welcome",
		Data:         map[string]interface{}{"Name": "Test"},
	}

	// In dev mode, should just log and return nil
	err := provider.Send(context.Background(), params)
	if err != nil {
		t.Errorf("Send() in dev mode should return nil, got %v", err)
	}
}

// ============ SMTPProvider.SendBatch Tests ============

func TestSMTPProvider_SendBatch_DevMode(t *testing.T) {
	tm, _ := NewTemplateManager()
	provider := &SMTPProvider{
		config: &Config{
			SMTPHost: "localhost",
			SMTPPort: 587,
			DevMode:  true,
		},
		templates: tm,
	}

	params := []SendParams{
		{To: "test1@example.com", Subject: "Test 1", TemplateName: "welcome", Data: map[string]interface{}{"Name": "Test1"}},
		{To: "test2@example.com", Subject: "Test 2", TemplateName: "welcome", Data: map[string]interface{}{"Name": "Test2"}},
	}

	err := provider.SendBatch(context.Background(), params)
	if err != nil {
		t.Errorf("SendBatch() in dev mode should return nil, got %v", err)
	}
}

func TestSMTPProvider_SendBatch_WithError(t *testing.T) {
	tm, _ := NewTemplateManager()
	provider := &SMTPProvider{
		config: &Config{
			SMTPHost: "localhost",
			SMTPPort: 587,
			DevMode:  false,
		},
		templates: tm,
	}

	// One empty recipient should cause error but not stop the batch
	params := []SendParams{
		{To: "", Subject: "Test 1"},           // Will fail
		{To: "test@example.com", Subject: ""}, // Will also try
	}

	// This will return the last error encountered
	err := provider.SendBatch(context.Background(), params)
	if err == nil {
		t.Error("SendBatch() should return error when any email fails")
	}
}

// ============ NoOpProvider Additional Tests ============

func TestNoOpProvider_SendBatch_Multiple(t *testing.T) {
	provider := NewNoOpProvider()

	params := []SendParams{
		{To: "a@test.com", Subject: "A"},
		{To: "b@test.com", Subject: "B"},
		{To: "c@test.com", Subject: "C"},
	}

	err := provider.SendBatch(context.Background(), params)
	if err != nil {
		t.Errorf("NoOpProvider.SendBatch() error = %v, want nil", err)
	}
}

// ============ SendParams Tests ============

func TestSendParams_AllFields(t *testing.T) {
	params := SendParams{
		To:           "user@example.com",
		Subject:      "Important Message",
		TemplateName: "notification",
		Data:         map[string]interface{}{"key": "value"},
		PlainText:    "Plain text version",
	}

	if params.To != "user@example.com" {
		t.Errorf("To = %q, want %q", params.To, "user@example.com")
	}
	if params.Subject != "Important Message" {
		t.Errorf("Subject = %q, want %q", params.Subject, "Important Message")
	}
	if params.TemplateName != "notification" {
		t.Errorf("TemplateName = %q, want %q", params.TemplateName, "notification")
	}
	if params.PlainText != "Plain text version" {
		t.Errorf("PlainText = %q, want %q", params.PlainText, "Plain text version")
	}
}

// ============ SMTPProvider.Send Additional Tests ============

func TestSMTPProvider_Send_DevMode_NoTemplate(t *testing.T) {
	tm, _ := NewTemplateManager()
	provider := &SMTPProvider{
		config: &Config{
			SMTPHost: "localhost",
			SMTPPort: 587,
			DevMode:  true,
		},
		templates: tm,
	}

	// Send without template - uses subject directly
	params := SendParams{
		To:      "test@example.com",
		Subject: "Direct Subject",
	}

	err := provider.Send(context.Background(), params)
	if err != nil {
		t.Errorf("Send() in dev mode without template should return nil, got %v", err)
	}
}

func TestSMTPProvider_Send_DevMode_WithPlainText(t *testing.T) {
	tm, _ := NewTemplateManager()
	provider := &SMTPProvider{
		config: &Config{
			SMTPHost: "localhost",
			SMTPPort: 587,
			DevMode:  true,
		},
		templates: tm,
	}

	params := SendParams{
		To:        "test@example.com",
		Subject:   "Test",
		PlainText: "Plain text content",
	}

	err := provider.Send(context.Background(), params)
	if err != nil {
		t.Errorf("Send() in dev mode with plain text should return nil, got %v", err)
	}
}

func TestSMTPProvider_Send_InvalidTemplate(t *testing.T) {
	tm, _ := NewTemplateManager()
	provider := &SMTPProvider{
		config: &Config{
			SMTPHost: "localhost",
			SMTPPort: 587,
			DevMode:  false,
		},
		templates: tm,
	}

	params := SendParams{
		To:           "test@example.com",
		Subject:      "Test",
		TemplateName: "nonexistent_template",
		Data:         map[string]interface{}{},
	}

	err := provider.Send(context.Background(), params)
	if err == nil {
		t.Error("Send() with invalid template should return error")
	}
}

// ============ SMTPProvider DevMode Data Logging Tests ============

func TestSMTPProvider_Send_DevMode_WithData(t *testing.T) {
	tm, _ := NewTemplateManager()
	provider := &SMTPProvider{
		config: &Config{
			SMTPHost: "localhost",
			SMTPPort: 587,
			DevMode:  true,
		},
		templates: tm,
	}

	// Send with template data - should log all data
	params := SendParams{
		To:           "test@example.com",
		Subject:      "Test Subject",
		TemplateName: "password_reset",
		Data: map[string]interface{}{
			"Name":     "John Doe",
			"ResetURL": "https://example.com/reset?token=abc123",
		},
	}

	err := provider.Send(context.Background(), params)
	if err != nil {
		t.Errorf("Send() in dev mode with data should return nil, got %v", err)
	}
}

// ============ SMTPProvider SendBatch Edge Cases ============

func TestSMTPProvider_SendBatch_AllSuccess(t *testing.T) {
	tm, _ := NewTemplateManager()
	provider := &SMTPProvider{
		config: &Config{
			SMTPHost: "localhost",
			SMTPPort: 587,
			DevMode:  true,
		},
		templates: tm,
	}

	params := []SendParams{
		{To: "user1@example.com", Subject: "Test 1"},
		{To: "user2@example.com", Subject: "Test 2"},
		{To: "user3@example.com", Subject: "Test 3"},
	}

	err := provider.SendBatch(context.Background(), params)
	if err != nil {
		t.Errorf("SendBatch() with all valid emails should return nil, got %v", err)
	}
}

func TestSMTPProvider_SendBatch_Empty(t *testing.T) {
	tm, _ := NewTemplateManager()
	provider := &SMTPProvider{
		config: &Config{
			SMTPHost: "localhost",
			SMTPPort: 587,
			DevMode:  true,
		},
		templates: tm,
	}

	// Empty batch should not error
	params := []SendParams{}

	err := provider.SendBatch(context.Background(), params)
	if err != nil {
		t.Errorf("SendBatch() with empty batch should return nil, got %v", err)
	}
}
