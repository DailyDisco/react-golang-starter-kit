package services

import (
	"testing"
)

// ============ Helper Function Tests ============

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"megabytes", 1024 * 1024, "1.0 MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.0 GB"},
		{"terabytes", 1024 * 1024 * 1024 * 1024, "1.0 TB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"2.5 MB", 2621440, "2.5 MB"},
		{"10 GB", 10737418240, "10.0 GB"},
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

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty string", "", []string{}},
		{"single line", "hello", []string{"hello"}},
		{"two lines", "hello\nworld", []string{"hello", "world"}},
		{"with carriage return", "hello\r\nworld", []string{"hello", "world"}},
		{"trailing newline", "hello\n", []string{"hello"}},
		{"multiple lines", "a\nb\nc", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("splitLines(%q) length = %d, want %d", tt.input, len(result), len(tt.expected))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("splitLines(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestSplitKeyValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      string
		expected []string
	}{
		{"simple key value", "key:value", ":", []string{"key", "value"}},
		{"no separator", "keyvalue", ":", []string{"keyvalue"}},
		{"multiple separators", "key:value:extra", ":", []string{"key", "value:extra"}},
		{"empty value", "key:", ":", []string{"key", ""}},
		{"empty key", ":value", ":", []string{"", "value"}},
		{"equals separator", "key=value", "=", []string{"key", "value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitKeyValue(tt.input, tt.sep)

			if len(result) != len(tt.expected) {
				t.Errorf("splitKeyValue(%q, %q) length = %d, want %d", tt.input, tt.sep, len(result), len(tt.expected))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("splitKeyValue(%q, %q)[%d] = %q, want %q", tt.input, tt.sep, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestParseRedisInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]interface{}{},
		},
		{
			name:     "single key-value",
			input:    "redis_version:7.0.0",
			expected: map[string]interface{}{"redis_version": "7.0.0"},
		},
		{
			name:     "multiple key-values",
			input:    "redis_version:7.0.0\nused_memory:1000000",
			expected: map[string]interface{}{"redis_version": "7.0.0", "used_memory": "1000000"},
		},
		{
			name:     "with comment lines",
			input:    "# Server\nredis_version:7.0.0\n# Memory\nused_memory:1000000",
			expected: map[string]interface{}{"redis_version": "7.0.0", "used_memory": "1000000"},
		},
		{
			name:     "with empty lines",
			input:    "key1:value1\n\nkey2:value2",
			expected: map[string]interface{}{"key1": "value1", "key2": "value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRedisInfo(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("parseRedisInfo() length = %d, want %d", len(result), len(tt.expected))
			}

			for key, expected := range tt.expected {
				if result[key] != expected {
					t.Errorf("parseRedisInfo()[%q] = %v, want %v", key, result[key], expected)
				}
			}
		})
	}
}

// ============ Health Service Constructor Tests ============

func TestNewHealthService(t *testing.T) {
	service := NewHealthService()

	if service == nil {
		t.Fatal("NewHealthService() returned nil")
	}

	// startTime should be set
	if service.startTime.IsZero() {
		t.Error("NewHealthService() should set startTime")
	}
}
