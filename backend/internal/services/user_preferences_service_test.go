package services

import (
	"testing"
)

// ============ Theme Validation Tests ============

func TestIsValidTheme(t *testing.T) {
	tests := []struct {
		name  string
		theme string
		want  bool
	}{
		{"light theme", "light", true},
		{"dark theme", "dark", true},
		{"system theme", "system", true},
		{"empty theme", "", false},
		{"invalid theme", "auto", false},
		{"uppercase light", "Light", false},
		{"uppercase dark", "DARK", false},
		{"mixed case", "System", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidTheme(tt.theme); got != tt.want {
				t.Errorf("isValidTheme(%q) = %v, want %v", tt.theme, got, tt.want)
			}
		})
	}
}

// ============ Language Validation Tests ============

func TestIsValidLanguage(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want bool
	}{
		{"English", "en", true},
		{"Spanish", "es", true},
		{"French", "fr", true},
		{"German", "de", true},
		{"Italian", "it", true},
		{"Portuguese", "pt", true},
		{"Japanese", "ja", true},
		{"Korean", "ko", true},
		{"Chinese", "zh", true},
		{"invalid code", "xx", false},
		{"empty", "", false},
		{"too long", "eng", false},
		{"uppercase", "EN", false},
		{"mixed case", "En", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidLanguage(tt.lang); got != tt.want {
				t.Errorf("isValidLanguage(%q) = %v, want %v", tt.lang, got, tt.want)
			}
		})
	}
}

// ============ Supported Languages Map Tests ============

func TestSupportedLanguages(t *testing.T) {
	expectedLanguages := map[string]string{
		"en": "English",
		"es": "Spanish",
		"fr": "French",
		"de": "German",
		"it": "Italian",
		"pt": "Portuguese",
		"ja": "Japanese",
		"ko": "Korean",
		"zh": "Chinese",
	}

	if len(SupportedLanguages) != len(expectedLanguages) {
		t.Errorf("SupportedLanguages has %d entries, want %d", len(SupportedLanguages), len(expectedLanguages))
	}

	for code, name := range expectedLanguages {
		if SupportedLanguages[code] != name {
			t.Errorf("SupportedLanguages[%q] = %q, want %q", code, SupportedLanguages[code], name)
		}
	}
}

// ============ Common Timezones Tests ============

func TestCommonTimezones(t *testing.T) {
	// Should include major timezones
	expectedTimezones := []string{
		"UTC",
		"America/New_York",
		"America/Los_Angeles",
		"Europe/London",
		"Asia/Tokyo",
		"Australia/Sydney",
	}

	timezoneSet := make(map[string]bool)
	for _, tz := range CommonTimezones {
		timezoneSet[tz] = true
	}

	for _, tz := range expectedTimezones {
		if !timezoneSet[tz] {
			t.Errorf("CommonTimezones should include %q", tz)
		}
	}
}

func TestCommonTimezones_NotEmpty(t *testing.T) {
	if len(CommonTimezones) == 0 {
		t.Error("CommonTimezones should not be empty")
	}

	if len(CommonTimezones) < 10 {
		t.Errorf("CommonTimezones should have at least 10 entries, got %d", len(CommonTimezones))
	}
}

// ============ Supported Date Formats Tests ============

func TestSupportedDateFormats(t *testing.T) {
	expectedFormats := []string{
		"MM/DD/YYYY",
		"DD/MM/YYYY",
		"YYYY-MM-DD",
		"DD.MM.YYYY",
		"YYYY/MM/DD",
	}

	if len(SupportedDateFormats) != len(expectedFormats) {
		t.Errorf("SupportedDateFormats has %d entries, want %d", len(SupportedDateFormats), len(expectedFormats))
	}

	formatSet := make(map[string]bool)
	for _, f := range SupportedDateFormats {
		formatSet[f] = true
	}

	for _, f := range expectedFormats {
		if !formatSet[f] {
			t.Errorf("SupportedDateFormats should include %q", f)
		}
	}
}
