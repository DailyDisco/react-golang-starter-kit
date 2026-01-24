package email

import (
	"testing"
)

// ============ toSnakeCase Tests ============

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Name", "name"},
		{"ResetURL", "reset_u_r_l"},
		{"resetURL", "reset_u_r_l"},
		{"UserName", "user_name"},
		{"userEmail", "user_email"},
		{"HTTPServer", "h_t_t_p_server"},
		{"simple", "simple"},
		{"ALLCAPS", "a_l_l_c_a_p_s"},
		{"", ""},
		{"A", "a"},
		{"ABC", "a_b_c"},
		{"FirstName", "first_name"},
		{"lastName", "last_name"},
		{"EmailAddress", "email_address"},
		{"ID", "i_d"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ============ RenderResult Tests ============

func TestRenderResult_Structure(t *testing.T) {
	result := RenderResult{
		Subject:  "Test Subject",
		BodyHTML: "<p>HTML Body</p>",
		BodyText: "Text Body",
		Source:   "database",
	}

	if result.Subject != "Test Subject" {
		t.Errorf("Subject = %q, want %q", result.Subject, "Test Subject")
	}

	if result.BodyHTML != "<p>HTML Body</p>" {
		t.Errorf("BodyHTML = %q, want %q", result.BodyHTML, "<p>HTML Body</p>")
	}

	if result.BodyText != "Text Body" {
		t.Errorf("BodyText = %q, want %q", result.BodyText, "Text Body")
	}

	if result.Source != "database" {
		t.Errorf("Source = %q, want %q", result.Source, "database")
	}
}

func TestRenderResult_Sources(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"database source", "database"},
		{"file source", "file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderResult{Source: tt.source}
			if result.Source != tt.source {
				t.Errorf("Source = %q, want %q", result.Source, tt.source)
			}
		})
	}
}

// ============ NewTemplateResolver Tests ============

func TestNewTemplateResolver(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	tr := NewTemplateResolver(nil, tm)

	if tr == nil {
		t.Fatal("NewTemplateResolver() returned nil")
	}

	if tr.fileTemplates != tm {
		t.Error("NewTemplateResolver() did not set fileTemplates correctly")
	}

	if tr.db != nil {
		t.Error("NewTemplateResolver() should have nil db")
	}
}

// ============ substituteVariables Tests ============

func TestTemplateResolver_SubstituteVariables(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	tr := NewTemplateResolver(nil, tm)

	tests := []struct {
		name    string
		content string
		data    map[string]interface{}
		want    string
	}{
		{
			name:    "simple substitution",
			content: "Hello, {{Name}}!",
			data:    map[string]interface{}{"Name": "John"},
			want:    "Hello, John!",
		},
		{
			name:    "multiple substitutions",
			content: "{{Name}} has email {{Email}}",
			data:    map[string]interface{}{"Name": "Jane", "Email": "jane@example.com"},
			want:    "Jane has email jane@example.com",
		},
		{
			name:    "snake_case key",
			content: "Hello, {{user_name}}!",
			data:    map[string]interface{}{"user_name": "Bob"},
			want:    "Hello, Bob!",
		},
		{
			name:    "no substitutions",
			content: "Plain text",
			data:    map[string]interface{}{},
			want:    "Plain text",
		},
		{
			name:    "unknown placeholder",
			content: "Hello, {{Unknown}}!",
			data:    map[string]interface{}{},
			want:    "Hello, {{Unknown}}!",
		},
		{
			name:    "number value",
			content: "Count: {{Count}}",
			data:    map[string]interface{}{"Count": 42},
			want:    "Count: 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tr.substituteVariables(tt.content, tt.data)
			if got != tt.want {
				t.Errorf("substituteVariables() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTemplateResolver_SubstituteVariables_CommonVars(t *testing.T) {
	// Set up config for common variables
	oldConfig := config
	config = &Config{
		SMTPFromName: "Test App",
		SMTPFrom:     "test@example.com",
		FrontendURL:  "https://test.com",
	}
	defer func() { config = oldConfig }()

	tm, _ := NewTemplateManager()
	tr := NewTemplateResolver(nil, tm)

	content := "Welcome to {{app_name}}! Support: {{support_email}}"
	result := tr.substituteVariables(content, nil)

	if result == content {
		t.Error("substituteVariables() should substitute common variables")
	}
}

func TestTemplateResolver_SubstituteVariables_PascalToSnake(t *testing.T) {
	tm, _ := NewTemplateManager()
	tr := NewTemplateResolver(nil, tm)

	// When passing PascalCase data, it should also create snake_case versions
	content := "Hello {{user_name}}!"
	data := map[string]interface{}{"UserName": "Alice"}
	result := tr.substituteVariables(content, data)

	if result != "Hello Alice!" {
		t.Errorf("substituteVariables() = %q, want %q", result, "Hello Alice!")
	}
}

// ============ renderFileTemplate Tests ============

func TestTemplateResolver_RenderFileTemplate(t *testing.T) {
	tm, err := NewTemplateManager()
	if err != nil {
		t.Fatalf("NewTemplateManager() error = %v", err)
	}

	tr := NewTemplateResolver(nil, tm)

	result, err := tr.renderFileTemplate("welcome", map[string]interface{}{
		"Name": "Test User",
	})

	if err != nil {
		t.Fatalf("renderFileTemplate() error = %v", err)
	}

	if result == nil {
		t.Fatal("renderFileTemplate() returned nil result")
	}

	if result.Source != "file" {
		t.Errorf("Source = %q, want %q", result.Source, "file")
	}

	if result.Subject == "" {
		t.Error("Subject should not be empty")
	}

	if result.BodyHTML == "" {
		t.Error("BodyHTML should not be empty")
	}

	if result.BodyText == "" {
		t.Error("BodyText should not be empty")
	}
}

func TestTemplateResolver_RenderFileTemplate_NotFound(t *testing.T) {
	tm, _ := NewTemplateManager()
	tr := NewTemplateResolver(nil, tm)

	_, err := tr.renderFileTemplate("nonexistent", nil)

	if err == nil {
		t.Error("renderFileTemplate() should return error for nonexistent template")
	}
}
