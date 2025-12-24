package sanitize

import (
	"testing"
)

func TestText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strips script tags",
			input:    "<script>alert('xss')</script>Hello",
			expected: "Hello",
		},
		{
			name:     "strips style tags",
			input:    "<style>body{display:none}</style>Content",
			expected: "Content",
		},
		{
			name:     "strips HTML tags",
			input:    "<b>Bold</b> <i>Italic</i>",
			expected: "Bold Italic",
		},
		{
			name:     "strips event handlers",
			input:    `<img onerror="alert(1)" src="x">`,
			expected: "",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "normalizes whitespace",
			input:    "  hello   world  ",
			expected: "hello world",
		},
		{
			name:     "preserves normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "decodes HTML entities",
			input:    "&lt;script&gt;",
			expected: "<script>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Text(tt.input)
			if result != tt.expected {
				t.Errorf("Text(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "escapes HTML tags",
			input:    "<script>",
			expected: "&lt;script&gt;",
		},
		{
			name:     "escapes quotes",
			input:    `"quoted"`,
			expected: "&#34;quoted&#34;",
		},
		{
			name:     "escapes ampersand",
			input:    "a & b",
			expected: "a &amp; b",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "preserves normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HTML(tt.input)
			if result != tt.expected {
				t.Errorf("HTML(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "allows https URLs",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "allows http URLs",
			input:    "http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "allows mailto URLs",
			input:    "mailto:test@example.com",
			expected: "mailto:test@example.com",
		},
		{
			name:     "allows relative URLs",
			input:    "/path/to/page",
			expected: "/path/to/page",
		},
		{
			name:     "blocks javascript URLs",
			input:    "javascript:alert('xss')",
			expected: "",
		},
		{
			name:     "blocks data URLs",
			input:    "data:text/html,<script>alert(1)</script>",
			expected: "",
		},
		{
			name:     "blocks vbscript URLs",
			input:    "vbscript:alert('xss')",
			expected: "",
		},
		{
			name:     "handles case variations",
			input:    "JAVASCRIPT:alert('xss')",
			expected: "",
		},
		{
			name:     "trims whitespace",
			input:    "  https://example.com  ",
			expected: "https://example.com",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := URL(tt.input)
			if result != tt.expected {
				t.Errorf("URL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "accepts valid email",
			input:    "user@example.com",
			expected: "user@example.com",
		},
		{
			name:     "normalizes to lowercase",
			input:    "USER@EXAMPLE.COM",
			expected: "user@example.com",
		},
		{
			name:     "rejects invalid email - no @",
			input:    "not-an-email",
			expected: "",
		},
		{
			name:     "rejects invalid email - no domain",
			input:    "user@",
			expected: "",
		},
		{
			name:     "rejects invalid email - no local part",
			input:    "@example.com",
			expected: "",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "strips HTML from email",
			input:    "<script>alert(1)</script>user@example.com",
			expected: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Email(tt.input)
			if result != tt.expected {
				t.Errorf("Email(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "preserves valid filename",
			input:    "document.pdf",
			expected: "document.pdf",
		},
		{
			name:     "preserves hyphens and underscores",
			input:    "my-file_v2.txt",
			expected: "my-file_v2.txt",
		},
		{
			name:     "removes path traversal",
			input:    "../../../etc/passwd",
			expected: "etcpasswd",
		},
		{
			name:     "removes forward slashes",
			input:    "path/to/file.txt",
			expected: "pathtofile.txt",
		},
		{
			name:     "removes backslashes",
			input:    "path\\to\\file.txt",
			expected: "pathtofile.txt",
		},
		{
			name:     "replaces dangerous characters",
			input:    "file<script>.txt",
			expected: "file_script_.txt",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles dot only",
			input:    ".",
			expected: "unnamed",
		},
		{
			name:     "handles double dot",
			input:    "..",
			expected: "unnamed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filename(tt.input)
			if result != tt.expected {
				t.Errorf("Filename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSQLString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "accepts normal text",
			input:    "Hello World",
			expected: true,
		},
		{
			name:     "accepts empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "detects UNION SELECT",
			input:    "1 UNION SELECT * FROM users",
			expected: false,
		},
		{
			name:     "detects DROP TABLE",
			input:    "'; DROP TABLE users; --",
			expected: false,
		},
		{
			name:     "detects INSERT INTO",
			input:    "'; INSERT INTO users VALUES",
			expected: false,
		},
		{
			name:     "detects UPDATE SET",
			input:    "'; UPDATE users SET admin=1",
			expected: false,
		},
		{
			name:     "detects DELETE FROM",
			input:    "'; DELETE FROM users",
			expected: false,
		},
		{
			name:     "handles case variations",
			input:    "UNION select * from users",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SQLString(tt.input)
			if result != tt.expected {
				t.Errorf("SQLString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStripNullBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes null bytes",
			input:    "hello\x00world",
			expected: "helloworld",
		},
		{
			name:     "preserves normal text",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "handles multiple null bytes",
			input:    "\x00hello\x00\x00world\x00",
			expected: "helloworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripNullBytes(tt.input)
			if result != tt.expected {
				t.Errorf("StripNullBytes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "truncates long string",
			input:     "Hello World",
			maxLength: 5,
			expected:  "Hello",
		},
		{
			name:      "preserves short string",
			input:     "Hi",
			maxLength: 5,
			expected:  "Hi",
		},
		{
			name:      "handles exact length",
			input:     "Hello",
			maxLength: 5,
			expected:  "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLength)
			if result != tt.expected {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.input, tt.maxLength, result, tt.expected)
			}
		})
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "title cases name",
			input:    "john doe",
			expected: "John Doe",
		},
		{
			name:     "handles uppercase",
			input:    "JOHN DOE",
			expected: "John Doe",
		},
		{
			name:     "handles mixed case",
			input:    "jOHN dOE",
			expected: "John Doe",
		},
		{
			name:     "strips HTML",
			input:    "<b>John</b> Doe",
			expected: "John Doe",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPassword(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		minLength int
		maxLength int
		expected  string
	}{
		{
			name:      "accepts valid password",
			input:     "MySecurePass123!",
			minLength: 8,
			maxLength: 128,
			expected:  "",
		},
		{
			name:      "rejects short password",
			input:     "short",
			minLength: 8,
			maxLength: 128,
			expected:  "password too short",
		},
		{
			name:      "rejects long password",
			input:     "a" + string(make([]byte, 200)),
			minLength: 8,
			maxLength: 128,
			expected:  "password too long",
		},
		{
			name:      "rejects null bytes",
			input:     "password\x00injection",
			minLength: 8,
			maxLength: 128,
			expected:  "password contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Password(tt.input, tt.minLength, tt.maxLength)
			if result != tt.expected {
				t.Errorf("Password() = %q, want %q", result, tt.expected)
			}
		})
	}
}
