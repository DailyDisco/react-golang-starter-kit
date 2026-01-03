package services

import (
	"encoding/json"
	"testing"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/testutil"
)

func testPreferencesSetup(t *testing.T) (*UserPreferencesService, func()) {
	t.Helper()
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tt := testutil.NewTestTransaction(t, db)

	// Set global database.DB for the preferences service
	oldDB := database.DB
	database.DB = tt.DB

	svc := NewUserPreferencesService()

	return svc, func() {
		database.DB = oldDB
		tt.Rollback()
	}
}

func createTestUserForPreferences(t *testing.T, email string) *models.User {
	t.Helper()
	user := &models.User{
		Email:    email,
		Name:     "Preferences Test User",
		Password: "hashedpassword",
		Role:     models.RoleUser,
	}
	if err := database.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func TestUserPreferencesService_GetPreferences_Integration(t *testing.T) {
	svc, cleanup := testPreferencesSetup(t)
	defer cleanup()

	t.Run("creates default preferences if not exists", func(t *testing.T) {
		user := createTestUserForPreferences(t, "newuser@example.com")

		prefs, err := svc.GetPreferences(user.ID)
		if err != nil {
			t.Fatalf("GetPreferences failed: %v", err)
		}

		if prefs.ID == 0 {
			t.Error("Expected preferences to have ID")
		}
		if prefs.UserID != user.ID {
			t.Errorf("Expected UserID %d, got: %d", user.ID, prefs.UserID)
		}
		if prefs.Theme != "system" {
			t.Errorf("Expected default theme 'system', got: %s", prefs.Theme)
		}
		if prefs.Timezone != "UTC" {
			t.Errorf("Expected default timezone 'UTC', got: %s", prefs.Timezone)
		}
		if prefs.Language != "en" {
			t.Errorf("Expected default language 'en', got: %s", prefs.Language)
		}
		if prefs.DateFormat != "MM/DD/YYYY" {
			t.Errorf("Expected default date format 'MM/DD/YYYY', got: %s", prefs.DateFormat)
		}
		if prefs.TimeFormat != "12h" {
			t.Errorf("Expected default time format '12h', got: %s", prefs.TimeFormat)
		}

		// Verify default email notifications
		var emailNotifs models.EmailNotificationSettings
		if err := json.Unmarshal(prefs.EmailNotifications, &emailNotifs); err != nil {
			t.Fatalf("Failed to unmarshal email notifications: %v", err)
		}
		if emailNotifs.Marketing != false {
			t.Error("Expected Marketing to be false by default")
		}
		if emailNotifs.Security != true {
			t.Error("Expected Security to be true by default")
		}
		if emailNotifs.Updates != true {
			t.Error("Expected Updates to be true by default")
		}
		if emailNotifs.WeeklyDigest != false {
			t.Error("Expected WeeklyDigest to be false by default")
		}
	})

	t.Run("returns existing preferences", func(t *testing.T) {
		user := createTestUserForPreferences(t, "existing@example.com")

		// Create preferences first time
		prefs1, err := svc.GetPreferences(user.ID)
		if err != nil {
			t.Fatalf("First GetPreferences failed: %v", err)
		}

		// Get again - should return same record
		prefs2, err := svc.GetPreferences(user.ID)
		if err != nil {
			t.Fatalf("Second GetPreferences failed: %v", err)
		}

		if prefs1.ID != prefs2.ID {
			t.Errorf("Expected same preferences record, got IDs %d and %d", prefs1.ID, prefs2.ID)
		}
	})
}

func TestUserPreferencesService_UpdatePreferences_Integration(t *testing.T) {
	svc, cleanup := testPreferencesSetup(t)
	defer cleanup()

	t.Run("updates all fields", func(t *testing.T) {
		user := createTestUserForPreferences(t, "updateall@example.com")

		theme := "dark"
		timezone := "America/New_York"
		language := "es"
		dateFormat := "DD/MM/YYYY"
		timeFormat := "24h"
		emailNotifs := &models.EmailNotificationSettings{
			Marketing:    true,
			Security:     true,
			Updates:      false,
			WeeklyDigest: true,
		}

		req := &models.UpdateUserPreferencesRequest{
			Theme:              &theme,
			Timezone:           &timezone,
			Language:           &language,
			DateFormat:         &dateFormat,
			TimeFormat:         &timeFormat,
			EmailNotifications: emailNotifs,
		}

		updated, err := svc.UpdatePreferences(user.ID, req)
		if err != nil {
			t.Fatalf("UpdatePreferences failed: %v", err)
		}

		if updated.Theme != "dark" {
			t.Errorf("Expected theme 'dark', got: %s", updated.Theme)
		}
		if updated.Timezone != "America/New_York" {
			t.Errorf("Expected timezone 'America/New_York', got: %s", updated.Timezone)
		}
		if updated.Language != "es" {
			t.Errorf("Expected language 'es', got: %s", updated.Language)
		}
		if updated.DateFormat != "DD/MM/YYYY" {
			t.Errorf("Expected date format 'DD/MM/YYYY', got: %s", updated.DateFormat)
		}
		if updated.TimeFormat != "24h" {
			t.Errorf("Expected time format '24h', got: %s", updated.TimeFormat)
		}

		var storedNotifs models.EmailNotificationSettings
		json.Unmarshal(updated.EmailNotifications, &storedNotifs)
		if storedNotifs.Marketing != true {
			t.Error("Expected Marketing to be true")
		}
		if storedNotifs.WeeklyDigest != true {
			t.Error("Expected WeeklyDigest to be true")
		}
	})

	t.Run("partial update only changes specified fields", func(t *testing.T) {
		user := createTestUserForPreferences(t, "partial@example.com")

		// Get defaults first
		original, _ := svc.GetPreferences(user.ID)

		// Update only theme
		theme := "light"
		req := &models.UpdateUserPreferencesRequest{
			Theme: &theme,
		}

		updated, err := svc.UpdatePreferences(user.ID, req)
		if err != nil {
			t.Fatalf("UpdatePreferences failed: %v", err)
		}

		if updated.Theme != "light" {
			t.Errorf("Expected theme 'light', got: %s", updated.Theme)
		}
		if updated.Language != original.Language {
			t.Errorf("Language should not have changed, got: %s", updated.Language)
		}
		if updated.Timezone != original.Timezone {
			t.Errorf("Timezone should not have changed, got: %s", updated.Timezone)
		}
	})

	t.Run("rejects invalid theme", func(t *testing.T) {
		user := createTestUserForPreferences(t, "invalidtheme@example.com")

		theme := "invalid-theme"
		req := &models.UpdateUserPreferencesRequest{
			Theme: &theme,
		}

		_, err := svc.UpdatePreferences(user.ID, req)
		if err == nil {
			t.Error("Expected error for invalid theme")
		}
		if err.Error() != "invalid theme: must be 'light', 'dark', or 'system'" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("rejects invalid language", func(t *testing.T) {
		user := createTestUserForPreferences(t, "invalidlang@example.com")

		language := "xx" // Invalid language code
		req := &models.UpdateUserPreferencesRequest{
			Language: &language,
		}

		_, err := svc.UpdatePreferences(user.ID, req)
		if err == nil {
			t.Error("Expected error for invalid language")
		}
		if err.Error() != "invalid language: must be a supported language code" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("rejects invalid time format", func(t *testing.T) {
		user := createTestUserForPreferences(t, "invalidtime@example.com")

		timeFormat := "invalid"
		req := &models.UpdateUserPreferencesRequest{
			TimeFormat: &timeFormat,
		}

		_, err := svc.UpdatePreferences(user.ID, req)
		if err == nil {
			t.Error("Expected error for invalid time format")
		}
		if err.Error() != "invalid time format: must be '12h' or '24h'" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}

func TestUserPreferencesService_UpdateTheme_Integration(t *testing.T) {
	svc, cleanup := testPreferencesSetup(t)
	defer cleanup()

	t.Run("sets theme to light", func(t *testing.T) {
		user := createTestUserForPreferences(t, "lighttheme@example.com")

		err := svc.UpdateTheme(user.ID, "light")
		if err != nil {
			t.Fatalf("UpdateTheme failed: %v", err)
		}

		prefs, _ := svc.GetPreferences(user.ID)
		if prefs.Theme != "light" {
			t.Errorf("Expected theme 'light', got: %s", prefs.Theme)
		}
	})

	t.Run("sets theme to dark", func(t *testing.T) {
		user := createTestUserForPreferences(t, "darktheme@example.com")

		err := svc.UpdateTheme(user.ID, "dark")
		if err != nil {
			t.Fatalf("UpdateTheme failed: %v", err)
		}

		prefs, _ := svc.GetPreferences(user.ID)
		if prefs.Theme != "dark" {
			t.Errorf("Expected theme 'dark', got: %s", prefs.Theme)
		}
	})

	t.Run("sets theme to system", func(t *testing.T) {
		user := createTestUserForPreferences(t, "systemtheme@example.com")
		// First change to something else
		svc.UpdateTheme(user.ID, "dark")

		err := svc.UpdateTheme(user.ID, "system")
		if err != nil {
			t.Fatalf("UpdateTheme failed: %v", err)
		}

		prefs, _ := svc.GetPreferences(user.ID)
		if prefs.Theme != "system" {
			t.Errorf("Expected theme 'system', got: %s", prefs.Theme)
		}
	})

	t.Run("rejects invalid theme", func(t *testing.T) {
		user := createTestUserForPreferences(t, "invalidtheme2@example.com")

		err := svc.UpdateTheme(user.ID, "invalid")
		if err == nil {
			t.Error("Expected error for invalid theme")
		}
	})
}

func TestUserPreferencesService_UpdateEmailNotifications_Integration(t *testing.T) {
	svc, cleanup := testPreferencesSetup(t)
	defer cleanup()

	t.Run("updates email notification settings", func(t *testing.T) {
		user := createTestUserForPreferences(t, "emailnotif@example.com")

		notifs := &models.EmailNotificationSettings{
			Marketing:    true,
			Security:     false,
			Updates:      true,
			WeeklyDigest: true,
		}

		err := svc.UpdateEmailNotifications(user.ID, notifs)
		if err != nil {
			t.Fatalf("UpdateEmailNotifications failed: %v", err)
		}

		prefs, _ := svc.GetPreferences(user.ID)
		var stored models.EmailNotificationSettings
		json.Unmarshal(prefs.EmailNotifications, &stored)

		if stored.Marketing != true {
			t.Error("Expected Marketing to be true")
		}
		if stored.Security != false {
			t.Error("Expected Security to be false")
		}
		if stored.Updates != true {
			t.Error("Expected Updates to be true")
		}
		if stored.WeeklyDigest != true {
			t.Error("Expected WeeklyDigest to be true")
		}
	})

	t.Run("verifies default notification settings", func(t *testing.T) {
		user := createTestUserForPreferences(t, "defaultnotif@example.com")

		prefs, _ := svc.GetPreferences(user.ID)
		var defaults models.EmailNotificationSettings
		json.Unmarshal(prefs.EmailNotifications, &defaults)

		// Verify defaults: Marketing=false, Security=true, Updates=true, WeeklyDigest=false
		if defaults.Marketing != false {
			t.Error("Default Marketing should be false")
		}
		if defaults.Security != true {
			t.Error("Default Security should be true")
		}
		if defaults.Updates != true {
			t.Error("Default Updates should be true")
		}
		if defaults.WeeklyDigest != false {
			t.Error("Default WeeklyDigest should be false")
		}
	})
}

func TestUserPreferencesService_DeletePreferences_Integration(t *testing.T) {
	svc, cleanup := testPreferencesSetup(t)
	defer cleanup()

	t.Run("deletes preferences record", func(t *testing.T) {
		user := createTestUserForPreferences(t, "delete@example.com")

		// Create preferences
		_, err := svc.GetPreferences(user.ID)
		if err != nil {
			t.Fatalf("GetPreferences failed: %v", err)
		}

		// Verify preferences exist
		var count int64
		database.DB.Model(&models.UserPreferences{}).Where("user_id = ?", user.ID).Count(&count)
		if count != 1 {
			t.Errorf("Expected 1 preference record, got: %d", count)
		}

		// Delete
		err = svc.DeletePreferences(user.ID)
		if err != nil {
			t.Fatalf("DeletePreferences failed: %v", err)
		}

		// Verify deleted
		database.DB.Model(&models.UserPreferences{}).Where("user_id = ?", user.ID).Count(&count)
		if count != 0 {
			t.Error("Expected preferences to be deleted")
		}
	})

	t.Run("deleting non-existent preferences does not error", func(t *testing.T) {
		// Try to delete preferences for a non-existent user
		err := svc.DeletePreferences(99999)
		if err != nil {
			t.Errorf("Expected no error deleting non-existent preferences, got: %v", err)
		}
	})
}

func TestUserPreferencesService_LanguageValidation_Integration(t *testing.T) {
	svc, cleanup := testPreferencesSetup(t)
	defer cleanup()

	t.Run("accepts all supported languages", func(t *testing.T) {
		supportedLangs := []string{"en", "es", "fr", "de", "it", "pt", "ja", "ko", "zh"}

		for i, lang := range supportedLangs {
			user := createTestUserForPreferences(t, "lang"+lang+"@example.com")

			langCopy := lang
			req := &models.UpdateUserPreferencesRequest{
				Language: &langCopy,
			}

			updated, err := svc.UpdatePreferences(user.ID, req)
			if err != nil {
				t.Errorf("Language %s should be accepted, got error: %v", lang, err)
			}
			if updated.Language != lang {
				t.Errorf("Expected language %s, got: %s", lang, updated.Language)
			}
			_ = i // Suppress unused variable warning
		}
	})

	t.Run("rejects unsupported language codes", func(t *testing.T) {
		invalidLangs := []string{"xx", "xyz", "english", "123", ""}

		for i, lang := range invalidLangs {
			user := createTestUserForPreferences(t, "invalidlang"+string(rune('0'+i))+"@example.com")

			langCopy := lang
			req := &models.UpdateUserPreferencesRequest{
				Language: &langCopy,
			}

			_, err := svc.UpdatePreferences(user.ID, req)
			if err == nil {
				t.Errorf("Language %q should be rejected", lang)
			}
		}
	})
}

func TestUserPreferencesService_TimezoneUpdate_Integration(t *testing.T) {
	svc, cleanup := testPreferencesSetup(t)
	defer cleanup()

	t.Run("accepts common timezones", func(t *testing.T) {
		commonTimezones := []string{
			"UTC",
			"America/New_York",
			"Europe/London",
			"Asia/Tokyo",
			"Australia/Sydney",
		}

		for i, tz := range commonTimezones {
			user := createTestUserForPreferences(t, "tz"+string(rune('0'+i))+"@example.com")

			tzCopy := tz
			req := &models.UpdateUserPreferencesRequest{
				Timezone: &tzCopy,
			}

			updated, err := svc.UpdatePreferences(user.ID, req)
			if err != nil {
				t.Errorf("Timezone %s update failed: %v", tz, err)
			}
			if updated.Timezone != tz {
				t.Errorf("Expected timezone %s, got: %s", tz, updated.Timezone)
			}
		}
	})
}
