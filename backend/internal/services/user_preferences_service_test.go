package services

import (
	"encoding/json"
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

// ============ Constructor Tests ============

func TestNewUserPreferencesService(t *testing.T) {
	svc := NewUserPreferencesService()
	if svc == nil {
		t.Fatal("NewUserPreferencesService() returned nil")
	}
}

// ============ Extended Theme Validation Tests ============

func TestIsValidTheme_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		theme string
		want  bool
	}{
		{"leading space", " light", false},
		{"trailing space", "light ", false},
		{"with number", "dark1", false},
		{"partial match", "lig", false},
		{"longer match", "lighter", false},
		{"null string", "\x00", false},
		{"unicode", "dark☀", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidTheme(tt.theme); got != tt.want {
				t.Errorf("isValidTheme(%q) = %v, want %v", tt.theme, got, tt.want)
			}
		})
	}
}

// ============ Extended Language Validation Tests ============

func TestIsValidLanguage_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want bool
	}{
		{"leading space", " en", false},
		{"trailing space", "en ", false},
		{"three letter", "eng", false},
		{"one letter", "e", false},
		{"with number", "e1", false},
		{"null string", "\x00", false},
		{"russian code", "ru", false}, // Not in supported languages
		{"arabic code", "ar", false},  // Not in supported languages
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidLanguage(tt.lang); got != tt.want {
				t.Errorf("isValidLanguage(%q) = %v, want %v", tt.lang, got, tt.want)
			}
		})
	}
}

// ============ Supported Languages Completeness ============

func TestSupportedLanguages_HasNames(t *testing.T) {
	for code, name := range SupportedLanguages {
		if code == "" {
			t.Error("SupportedLanguages has empty code")
		}
		if name == "" {
			t.Errorf("SupportedLanguages[%q] has empty name", code)
		}
		if len(code) != 2 {
			t.Errorf("SupportedLanguages code %q should be 2 characters", code)
		}
	}
}

// ============ Timezones Validation ============

func TestCommonTimezones_AllValid(t *testing.T) {
	for _, tz := range CommonTimezones {
		if tz == "" {
			t.Error("CommonTimezones contains empty string")
		}
		// Check basic format (should contain a slash for most, or be UTC)
		if tz != "UTC" && !containsSlash(tz) {
			t.Errorf("Timezone %q doesn't look like a valid timezone format", tz)
		}
	}
}

func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}

func TestCommonTimezones_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, tz := range CommonTimezones {
		if seen[tz] {
			t.Errorf("CommonTimezones contains duplicate: %q", tz)
		}
		seen[tz] = true
	}
}

// ============ Date Formats Validation ============

func TestSupportedDateFormats_AllValid(t *testing.T) {
	for _, format := range SupportedDateFormats {
		if format == "" {
			t.Error("SupportedDateFormats contains empty string")
		}
		// Each format should contain Y, M, D
		hasY := false
		hasM := false
		hasD := false
		for _, c := range format {
			if c == 'Y' {
				hasY = true
			}
			if c == 'M' {
				hasM = true
			}
			if c == 'D' {
				hasD = true
			}
		}
		if !hasY || !hasM || !hasD {
			t.Errorf("Date format %q should contain Y, M, and D", format)
		}
	}
}

func TestSupportedDateFormats_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, format := range SupportedDateFormats {
		if seen[format] {
			t.Errorf("SupportedDateFormats contains duplicate: %q", format)
		}
		seen[format] = true
	}
}

// ============ createDefaultPreferences Tests ============

func TestCreateDefaultPreferences(t *testing.T) {
	svc := NewUserPreferencesService()

	tests := []struct {
		name   string
		userID uint
	}{
		{"user 1", 1},
		{"user 999", 999},
		{"user 0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefs := svc.createDefaultPreferences(tt.userID)

			// Verify user ID is set
			if prefs.UserID != tt.userID {
				t.Errorf("UserID = %d, want %d", prefs.UserID, tt.userID)
			}

			// Verify defaults
			if prefs.Theme != "system" {
				t.Errorf("Theme = %q, want %q", prefs.Theme, "system")
			}
			if prefs.Timezone != "UTC" {
				t.Errorf("Timezone = %q, want %q", prefs.Timezone, "UTC")
			}
			if prefs.Language != "en" {
				t.Errorf("Language = %q, want %q", prefs.Language, "en")
			}
			if prefs.DateFormat != "MM/DD/YYYY" {
				t.Errorf("DateFormat = %q, want %q", prefs.DateFormat, "MM/DD/YYYY")
			}
			if prefs.TimeFormat != "12h" {
				t.Errorf("TimeFormat = %q, want %q", prefs.TimeFormat, "12h")
			}

			// Verify email notifications JSON is set
			if len(prefs.EmailNotifications) == 0 {
				t.Error("EmailNotifications should not be empty")
			}

			// Verify timestamps are set
			if prefs.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set")
			}
			if prefs.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be set")
			}
		})
	}
}

func TestCreateDefaultPreferences_EmailNotificationsContent(t *testing.T) {
	svc := NewUserPreferencesService()
	prefs := svc.createDefaultPreferences(1)

	// Parse the email notifications JSON
	var notif map[string]bool
	if err := json.Unmarshal(prefs.EmailNotifications, &notif); err != nil {
		t.Fatalf("Failed to unmarshal EmailNotifications: %v", err)
	}

	// Verify default values
	if notif["marketing"] != false {
		t.Error("Marketing should default to false")
	}
	if notif["security"] != true {
		t.Error("Security should default to true")
	}
	if notif["updates"] != true {
		t.Error("Updates should default to true")
	}
	if notif["weekly_digest"] != false {
		t.Error("WeeklyDigest should default to false")
	}
}
