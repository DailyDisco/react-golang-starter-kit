package email

import (
	"testing"
)

// ============ extractTemplateName Tests ============

func TestExtractTemplateName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"templates/emails/welcome.html", "welcome"},
		{"templates/emails/password_reset.html", "password_reset"},
		{"templates/emails/email_change_verify.html", "email_change_verify"},
		{"templates/emails/two_factor_code.html", "two_factor_code"},
		{"welcome.html", "welcome"}, // Edge case
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := extractTemplateName(tt.path)
			if got != tt.want {
				t.Errorf("extractTemplateName(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// ============ generatePlainText Tests ============

func TestGeneratePlainText_BasicHTML(t *testing.T) {
	html := "<p>Hello, World!</p>"
	text := generatePlainText(html)

	if text != "Hello, World!" {
		t.Errorf("generatePlainText() = %q, want %q", text, "Hello, World!")
	}
}

func TestGeneratePlainText_LineBreaks(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{"br tag", "Line 1<br>Line 2", "Line 1\nLine 2"},
		{"br self-closing", "Line 1<br/>Line 2", "Line 1\nLine 2"},
		{"br with space", "Line 1<br />Line 2", "Line 1\nLine 2"},
		{"paragraph end", "<p>Para 1</p><p>Para 2</p>", "Para 1\n\nPara 2"},
		{"div end", "<div>Div 1</div><div>Div 2</div>", "Div 1\nDiv 2"},
		{"li end", "<ul><li>Item 1</li><li>Item 2</li></ul>", "Item 1\nItem 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generatePlainText(tt.html)
			if got != tt.want {
				t.Errorf("generatePlainText(%q) = %q, want %q", tt.html, got, tt.want)
			}
		})
	}
}

func TestGeneratePlainText_Headers(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{"h1", "<h1>Title</h1>", "Title"},
		{"h2", "<h2>Subtitle</h2>", "Subtitle"},
		{"h3", "<h3>Section</h3>", "Section"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generatePlainText(tt.html)
			if got != tt.want {
				t.Errorf("generatePlainText(%q) = %q, want %q", tt.html, got, tt.want)
			}
		})
	}
}

func TestGeneratePlainText_StripsTags(t *testing.T) {
	html := "<div class=\"container\"><span style=\"color: red;\">Text</span></div>"
	text := generatePlainText(html)

	if text != "Text" {
		t.Errorf("generatePlainText() = %q, want %q", text, "Text")
	}
}

func TestGeneratePlainText_RemovesComments(t *testing.T) {
	html := "<p>Before<!-- This is a comment -->After</p>"
	text := generatePlainText(html)

	if text != "BeforeAfter" {
		t.Errorf("generatePlainText() = %q, want %q", text, "BeforeAfter")
	}
}

func TestGeneratePlainText_MultipleComments(t *testing.T) {
	html := "<!-- Comment 1 -->Text<!-- Comment 2 -->"
	text := generatePlainText(html)

	if text != "Text" {
		t.Errorf("generatePlainText() = %q, want %q", text, "Text")
	}
}

func TestGeneratePlainText_CollapsesNewlines(t *testing.T) {
	html := "<p>Para 1</p>\n\n\n<p>Para 2</p>"
	text := generatePlainText(html)

	// Should not have more than 2 consecutive newlines
	if text == "Para 1\n\n\n\nPara 2" {
		t.Error("generatePlainText() should collapse multiple newlines")
	}
}

func TestGeneratePlainText_TrimWhitespace(t *testing.T) {
	html := "   <p>Text</p>   "
	text := generatePlainText(html)

	if text != "Text" {
		t.Errorf("generatePlainText() = %q, want %q", text, "Text")
	}
}

func TestGeneratePlainText_EmptyInput(t *testing.T) {
	text := generatePlainText("")

	if text != "" {
		t.Errorf("generatePlainText(\"\") = %q, want empty string", text)
	}
}

func TestGeneratePlainText_ComplexHTML(t *testing.T) {
	html := `
		<!DOCTYPE html>
		<html>
		<head><title>Test</title></head>
		<body>
			<h1>Welcome</h1>
			<p>Hello, <strong>John</strong>!</p>
			<ul>
				<li>Item 1</li>
				<li>Item 2</li>
			</ul>
		</body>
		</html>
	`
	text := generatePlainText(html)

	// Should contain the text content
	if len(text) == 0 {
		t.Error("generatePlainText() returned empty for complex HTML")
	}

	// Should not contain HTML tags
	if contains(text, "<") || contains(text, ">") {
		t.Error("generatePlainText() should strip all HTML tags")
	}
}

// ============ TemplateData Tests ============

func TestTemplateData_Structure(t *testing.T) {
	data := TemplateData{
		AppName:      "My App",
		SupportEmail: "support@myapp.com",
		CurrentYear:  2025,
		FrontendURL:  "https://myapp.com",
		Data: map[string]interface{}{
			"Name": "John",
		},
	}

	if data.AppName != "My App" {
		t.Errorf("AppName = %q, want %q", data.AppName, "My App")
	}

	if data.SupportEmail != "support@myapp.com" {
		t.Errorf("SupportEmail = %q, want %q", data.SupportEmail, "support@myapp.com")
	}

	if data.CurrentYear != 2025 {
		t.Errorf("CurrentYear = %d, want %d", data.CurrentYear, 2025)
	}

	if data.FrontendURL != "https://myapp.com" {
		t.Errorf("FrontendURL = %q, want %q", data.FrontendURL, "https://myapp.com")
	}

	if data.Data["Name"] != "John" {
		t.Errorf("Data[Name] = %v, want %q", data.Data["Name"], "John")
	}
}

// ============ Helper Tests ============

func TestGetAppName_NoConfig(t *testing.T) {
	// Save and restore config
	oldConfig := config
	config = nil
	defer func() { config = oldConfig }()

	name := getAppName()
	if name != "React Go Starter" {
		t.Errorf("getAppName() = %q, want %q", name, "React Go Starter")
	}
}

func TestGetAppName_WithConfig(t *testing.T) {
	oldConfig := config
	config = &Config{SMTPFromName: "Custom App"}
	defer func() { config = oldConfig }()

	name := getAppName()
	if name != "Custom App" {
		t.Errorf("getAppName() = %q, want %q", name, "Custom App")
	}
}

func TestGetSupportEmail_NoConfig(t *testing.T) {
	oldConfig := config
	config = nil
	defer func() { config = oldConfig }()

	email := getSupportEmail()
	if email != "support@example.com" {
		t.Errorf("getSupportEmail() = %q, want %q", email, "support@example.com")
	}
}

func TestGetSupportEmail_WithConfig(t *testing.T) {
	oldConfig := config
	config = &Config{SMTPFrom: "help@myapp.com"}
	defer func() { config = oldConfig }()

	email := getSupportEmail()
	if email != "help@myapp.com" {
		t.Errorf("getSupportEmail() = %q, want %q", email, "help@myapp.com")
	}
}

// ============ TemplateManager Tests ============

func TestTemplateManager_HasTemplate(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	tests := []struct {
		name     string
		template string
		want     bool
	}{
		{"existing template", "welcome", true},
		{"nonexistent template", "nonexistent", false},
		{"password reset", "password_reset", true},
		{"verification", "verification", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tm.HasTemplate(tt.template); got != tt.want {
				t.Errorf("HasTemplate(%q) = %v, want %v", tt.template, got, tt.want)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============ Additional TemplateManager Tests ============

func TestTemplateManager_ListTemplates_NotEmpty(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	templates := tm.ListTemplates()
	if len(templates) == 0 {
		t.Error("ListTemplates() should return non-empty list")
	}

	// Check that known templates are in the list
	found := make(map[string]bool)
	for _, name := range templates {
		found[name] = true
	}

	expectedTemplates := []string{"welcome", "password_reset", "verification"}
	for _, expected := range expectedTemplates {
		if !found[expected] {
			t.Errorf("ListTemplates() missing expected template %q", expected)
		}
	}
}

// ============ Additional Helper Tests ============

func TestGetAppName_EmptyFromName(t *testing.T) {
	oldConfig := config
	config = &Config{SMTPFromName: ""} // Empty but config exists
	defer func() { config = oldConfig }()

	name := getAppName()
	// Should fall back to default when SMTPFromName is empty
	if name != "React Go Starter" {
		t.Errorf("getAppName() = %q, want %q", name, "React Go Starter")
	}
}

func TestGetSupportEmail_EmptyFrom(t *testing.T) {
	oldConfig := config
	config = &Config{SMTPFrom: ""} // Empty but config exists
	defer func() { config = oldConfig }()

	email := getSupportEmail()
	// Should fall back to default when SMTPFrom is empty
	if email != "support@example.com" {
		t.Errorf("getSupportEmail() = %q, want %q", email, "support@example.com")
	}
}

// ============ Additional generatePlainText Tests ============

func TestGeneratePlainText_UnclosedComment(t *testing.T) {
	// Test with comment that never closes
	html := "<p>Text<!-- unclosed comment"
	text := generatePlainText(html)

	// Should handle gracefully without infinite loop
	if text == "" {
		t.Error("generatePlainText() should handle unclosed comments")
	}
}

func TestGeneratePlainText_NestedTags(t *testing.T) {
	html := "<div><p><strong><em>Nested</em></strong></p></div>"
	text := generatePlainText(html)

	if text != "Nested" {
		t.Errorf("generatePlainText() = %q, want %q", text, "Nested")
	}
}

func TestGeneratePlainText_IncompleteTags(t *testing.T) {
	// Test with incomplete HTML
	html := "<p>Start<div>Middle"
	text := generatePlainText(html)

	// Should extract text even with broken HTML
	if text != "StartMiddle" {
		t.Errorf("generatePlainText() = %q, want %q", text, "StartMiddle")
	}
}

// ============ TemplateManager Render Tests ============

func TestTemplateManager_Render_AllTemplates(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	// Test rendering all templates to ensure they don't error
	templates := tm.ListTemplates()
	for _, name := range templates {
		t.Run(name, func(t *testing.T) {
			// Provide common data that templates might need
			data := map[string]interface{}{
				"Name":           "Test User",
				"ResetURL":       "https://example.com/reset",
				"VerifyURL":      "https://example.com/verify",
				"Code":           "123456",
				"Device":         "Chrome on Windows",
				"Location":       "New York, USA",
				"IP":             "192.168.1.1",
				"Time":           "Jan 1, 2025 at 10:00 AM",
				"LockDuration":   "30 minutes",
				"FailedAttempts": "5",
			}

			subject, html, text, err := tm.Render(name, data)
			if err != nil {
				t.Errorf("Render(%q) error = %v", name, err)
				return
			}

			// All templates should produce some output
			if subject == "" {
				t.Errorf("Render(%q) returned empty subject", name)
			}
			if html == "" {
				t.Errorf("Render(%q) returned empty HTML body", name)
			}
			if text == "" {
				t.Errorf("Render(%q) returned empty text body", name)
			}
		})
	}
}
