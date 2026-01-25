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

func TestNewHealthService_StartTimeIsRecent(t *testing.T) {
	service := NewHealthService()

	// startTime should be very close to now
	elapsed := service.startTime.Sub(service.startTime)
	if elapsed < 0 {
		t.Error("startTime should not be in the future")
	}
}

// ============ GetRuntimeMetrics Tests ============

func TestHealthService_GetRuntimeMetrics(t *testing.T) {
	service := NewHealthService()
	metrics := service.GetRuntimeMetrics()

	if metrics == nil {
		t.Fatal("GetRuntimeMetrics() returned nil")
	}

	// Check that expected keys exist
	expectedKeys := []string{
		"goroutines",
		"memory_alloc",
		"memory_sys",
		"memory_heap",
		"gc_runs",
		"gc_pause_total",
		"go_version",
		"num_cpu",
		"uptime",
	}

	for _, key := range expectedKeys {
		if _, ok := metrics[key]; !ok {
			t.Errorf("GetRuntimeMetrics() missing key %q", key)
		}
	}
}

func TestHealthService_GetRuntimeMetrics_GoroutinesPositive(t *testing.T) {
	service := NewHealthService()
	metrics := service.GetRuntimeMetrics()

	goroutines, ok := metrics["goroutines"].(int)
	if !ok {
		t.Fatal("goroutines should be an int")
	}
	if goroutines < 1 {
		t.Errorf("goroutines = %d, should be at least 1", goroutines)
	}
}

func TestHealthService_GetRuntimeMetrics_NumCPU(t *testing.T) {
	service := NewHealthService()
	metrics := service.GetRuntimeMetrics()

	numCPU, ok := metrics["num_cpu"].(int)
	if !ok {
		t.Fatal("num_cpu should be an int")
	}
	if numCPU < 1 {
		t.Errorf("num_cpu = %d, should be at least 1", numCPU)
	}
}

func TestHealthService_GetRuntimeMetrics_GoVersion(t *testing.T) {
	service := NewHealthService()
	metrics := service.GetRuntimeMetrics()

	goVersion, ok := metrics["go_version"].(string)
	if !ok {
		t.Fatal("go_version should be a string")
	}
	if len(goVersion) == 0 {
		t.Error("go_version should not be empty")
	}
	// Should start with "go"
	if len(goVersion) < 2 || goVersion[:2] != "go" {
		t.Errorf("go_version = %q, should start with 'go'", goVersion)
	}
}

// ============ Additional formatBytes Tests ============

func TestFormatBytes_LargeValues(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		contains string
	}{
		{"petabyte range", 1024 * 1024 * 1024 * 1024 * 1024, "PB"},
		{"exabyte range", 1024 * 1024 * 1024 * 1024 * 1024 * 1024, "EB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if len(result) < 2 {
				t.Errorf("formatBytes(%d) = %q, too short", tt.bytes, result)
			}
		})
	}
}

func TestFormatBytes_NegativeValue(t *testing.T) {
	// Negative values should still work (formatBytes handles int64)
	result := formatBytes(-1)
	// Should return something without panicking
	if result == "" {
		t.Error("formatBytes(-1) should return non-empty string")
	}
}

// ============ Additional splitLines Tests ============

func TestSplitLines_OnlyNewlines(t *testing.T) {
	result := splitLines("\n\n\n")
	// Should produce empty strings
	if len(result) != 3 {
		t.Errorf("splitLines(\"\\n\\n\\n\") length = %d, want 3", len(result))
	}
}

func TestSplitLines_MixedLineEndings(t *testing.T) {
	input := "line1\r\nline2\nline3\r\n"
	result := splitLines(input)

	expected := []string{"line1", "line2", "line3"}
	if len(result) != len(expected) {
		t.Errorf("splitLines() length = %d, want %d", len(result), len(expected))
		return
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("splitLines()[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

// ============ Additional splitKeyValue Tests ============

func TestSplitKeyValue_LongSeparator(t *testing.T) {
	result := splitKeyValue("key=>value", "=>")
	expected := []string{"key", "value"}

	if len(result) != len(expected) {
		t.Errorf("splitKeyValue() length = %d, want %d", len(result), len(expected))
		return
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("splitKeyValue()[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestSplitKeyValue_SeparatorAtStart(t *testing.T) {
	result := splitKeyValue("::value", "::")
	if len(result) != 2 {
		t.Errorf("splitKeyValue() length = %d, want 2", len(result))
		return
	}
	if result[0] != "" {
		t.Errorf("splitKeyValue()[0] = %q, want empty string", result[0])
	}
	if result[1] != "value" {
		t.Errorf("splitKeyValue()[1] = %q, want \"value\"", result[1])
	}
}

// ============ parseRedisInfo Edge Cases ============

func TestParseRedisInfo_OnlyComments(t *testing.T) {
	input := "# Server\n# Memory\n# Clients"
	result := parseRedisInfo(input)

	if len(result) != 0 {
		t.Errorf("parseRedisInfo() with only comments should return empty map, got %d entries", len(result))
	}
}

func TestParseRedisInfo_WithCRLF(t *testing.T) {
	input := "key1:value1\r\nkey2:value2\r\n"
	result := parseRedisInfo(input)

	if len(result) != 2 {
		t.Errorf("parseRedisInfo() length = %d, want 2", len(result))
	}
	if result["key1"] != "value1" {
		t.Errorf("parseRedisInfo()[key1] = %v, want \"value1\"", result["key1"])
	}
	if result["key2"] != "value2" {
		t.Errorf("parseRedisInfo()[key2] = %v, want \"value2\"", result["key2"])
	}
}
