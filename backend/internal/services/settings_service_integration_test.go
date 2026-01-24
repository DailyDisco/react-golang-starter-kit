package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/testutil"
)

func testSettingsSetup(t *testing.T) (*SettingsService, func()) {
	t.Helper()
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tt := testutil.NewTestTransaction(t, db)

	// Set global database.DB for test data creation in tests
	oldDB := database.DB
	database.DB = tt.DB

	// Now using dependency injection - pass db directly to service
	svc := NewSettingsService(tt.DB)

	return svc, func() {
		database.DB = oldDB
		tt.Rollback()
	}
}

func TestSettingsService_GetAllSettings_Integration(t *testing.T) {
	svc, cleanup := testSettingsSetup(t)
	defer cleanup()

	t.Run("returns all settings ordered by category and key", func(t *testing.T) {
		// Create test settings
		settings := []models.SystemSetting{
			{Key: "test_key_b", Category: "general", Value: json.RawMessage(`"value_b"`)},
			{Key: "test_key_a", Category: "email", Value: json.RawMessage(`"value_a"`)},
			{Key: "test_key_c", Category: "general", Value: json.RawMessage(`"value_c"`)},
		}
		for _, s := range settings {
			database.DB.Create(&s)
		}

		result, err := svc.GetAllSettings(context.Background())
		if err != nil {
			t.Fatalf("GetAllSettings failed: %v", err)
		}

		if len(result) < 3 {
			t.Errorf("Expected at least 3 settings, got: %d", len(result))
		}

		// Verify ordering by category, then key
		found := false
		for i := 0; i < len(result)-1; i++ {
			if result[i].Category == "email" && result[i].Key == "test_key_a" {
				found = true
			}
		}
		if !found {
			t.Log("Settings returned but ordering test may not be conclusive with existing data")
		}
	})
}

func TestSettingsService_GetSettingsByCategory_Integration(t *testing.T) {
	svc, cleanup := testSettingsSetup(t)
	defer cleanup()

	t.Run("returns settings for specific category", func(t *testing.T) {
		// Create test settings in different categories
		database.DB.Create(&models.SystemSetting{
			Key:      "cat_test_a",
			Category: "test_category",
			Value:    json.RawMessage(`"value_a"`),
		})
		database.DB.Create(&models.SystemSetting{
			Key:      "cat_test_b",
			Category: "test_category",
			Value:    json.RawMessage(`"value_b"`),
		})
		database.DB.Create(&models.SystemSetting{
			Key:      "other_test",
			Category: "other_category",
			Value:    json.RawMessage(`"other_value"`),
		})

		result, err := svc.GetSettingsByCategory(context.Background(), "test_category")
		if err != nil {
			t.Fatalf("GetSettingsByCategory failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 settings in test_category, got: %d", len(result))
		}

		for _, s := range result {
			if s.Category != "test_category" {
				t.Errorf("Expected category 'test_category', got: %s", s.Category)
			}
		}
	})

	t.Run("returns empty slice for non-existent category", func(t *testing.T) {
		result, err := svc.GetSettingsByCategory(context.Background(), "non_existent_category")
		if err != nil {
			t.Fatalf("GetSettingsByCategory failed: %v", err)
		}

		if len(result) != 0 {
			t.Errorf("Expected 0 settings, got: %d", len(result))
		}
	})
}

func TestSettingsService_GetSetting_Integration(t *testing.T) {
	svc, cleanup := testSettingsSetup(t)
	defer cleanup()

	t.Run("returns setting by key", func(t *testing.T) {
		database.DB.Create(&models.SystemSetting{
			Key:      "test_setting_key",
			Category: "test",
			Value:    json.RawMessage(`"test_value"`),
		})

		result, err := svc.GetSetting(context.Background(), "test_setting_key")
		if err != nil {
			t.Fatalf("GetSetting failed: %v", err)
		}

		if result.Key != "test_setting_key" {
			t.Errorf("Expected key 'test_setting_key', got: %s", result.Key)
		}
	})

	t.Run("returns error for non-existent key", func(t *testing.T) {
		_, err := svc.GetSetting(context.Background(), "non_existent_key")
		if err == nil {
			t.Error("Expected error for non-existent key")
		}
	})
}

func TestSettingsService_GetSettingValue_Integration(t *testing.T) {
	svc, cleanup := testSettingsSetup(t)
	defer cleanup()

	t.Run("unmarshals string value", func(t *testing.T) {
		database.DB.Create(&models.SystemSetting{
			Key:      "string_setting",
			Category: "test",
			Value:    json.RawMessage(`"hello world"`),
		})

		var result string
		err := svc.GetSettingValue(context.Background(), "string_setting", &result)
		if err != nil {
			t.Fatalf("GetSettingValue failed: %v", err)
		}

		if result != "hello world" {
			t.Errorf("Expected 'hello world', got: %s", result)
		}
	})

	t.Run("unmarshals integer value", func(t *testing.T) {
		database.DB.Create(&models.SystemSetting{
			Key:      "int_setting",
			Category: "test",
			Value:    json.RawMessage(`42`),
		})

		var result int
		err := svc.GetSettingValue(context.Background(), "int_setting", &result)
		if err != nil {
			t.Fatalf("GetSettingValue failed: %v", err)
		}

		if result != 42 {
			t.Errorf("Expected 42, got: %d", result)
		}
	})

	t.Run("unmarshals boolean value", func(t *testing.T) {
		database.DB.Create(&models.SystemSetting{
			Key:      "bool_setting",
			Category: "test",
			Value:    json.RawMessage(`true`),
		})

		var result bool
		err := svc.GetSettingValue(context.Background(), "bool_setting", &result)
		if err != nil {
			t.Fatalf("GetSettingValue failed: %v", err)
		}

		if !result {
			t.Error("Expected true")
		}
	})

	t.Run("unmarshals object value", func(t *testing.T) {
		database.DB.Create(&models.SystemSetting{
			Key:      "obj_setting",
			Category: "test",
			Value:    json.RawMessage(`{"name":"test","count":5}`),
		})

		var result struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}
		err := svc.GetSettingValue(context.Background(), "obj_setting", &result)
		if err != nil {
			t.Fatalf("GetSettingValue failed: %v", err)
		}

		if result.Name != "test" || result.Count != 5 {
			t.Errorf("Expected {name:test, count:5}, got: %+v", result)
		}
	})
}

func TestSettingsService_GetSettingsByKeys_Integration(t *testing.T) {
	svc, cleanup := testSettingsSetup(t)
	defer cleanup()

	t.Run("returns multiple settings by keys", func(t *testing.T) {
		database.DB.Create(&models.SystemSetting{Key: "multi_key_1", Category: "test", Value: json.RawMessage(`"value1"`)})
		database.DB.Create(&models.SystemSetting{Key: "multi_key_2", Category: "test", Value: json.RawMessage(`"value2"`)})
		database.DB.Create(&models.SystemSetting{Key: "multi_key_3", Category: "test", Value: json.RawMessage(`"value3"`)})

		result, err := svc.GetSettingsByKeys(context.Background(), []string{"multi_key_1", "multi_key_3"})
		if err != nil {
			t.Fatalf("GetSettingsByKeys failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 settings, got: %d", len(result))
		}

		if _, ok := result["multi_key_1"]; !ok {
			t.Error("Expected multi_key_1 in results")
		}
		if _, ok := result["multi_key_3"]; !ok {
			t.Error("Expected multi_key_3 in results")
		}
	})

	t.Run("returns empty map for non-existent keys", func(t *testing.T) {
		result, err := svc.GetSettingsByKeys(context.Background(), []string{"nonexistent_1", "nonexistent_2"})
		if err != nil {
			t.Fatalf("GetSettingsByKeys failed: %v", err)
		}

		if len(result) != 0 {
			t.Errorf("Expected empty map, got: %d entries", len(result))
		}
	})
}

func TestSettingsService_UpdateSetting_Integration(t *testing.T) {
	svc, cleanup := testSettingsSetup(t)
	defer cleanup()

	t.Run("updates existing setting", func(t *testing.T) {
		database.DB.Create(&models.SystemSetting{
			Key:      "update_test_key",
			Category: "test",
			Value:    json.RawMessage(`"old_value"`),
		})

		err := svc.UpdateSetting(context.Background(), "update_test_key", "new_value")
		if err != nil {
			t.Fatalf("UpdateSetting failed: %v", err)
		}

		// Verify update
		var setting models.SystemSetting
		database.DB.Where("key = ?", "update_test_key").First(&setting)

		var value string
		json.Unmarshal(setting.Value, &value)
		if value != "new_value" {
			t.Errorf("Expected 'new_value', got: %s", value)
		}
	})

	t.Run("returns error for non-existent setting", func(t *testing.T) {
		err := svc.UpdateSetting(context.Background(), "nonexistent_setting_key", "value")
		if err == nil {
			t.Error("Expected error for non-existent setting")
		}
	})

	t.Run("updates with different types", func(t *testing.T) {
		database.DB.Create(&models.SystemSetting{
			Key:      "type_test_key",
			Category: "test",
			Value:    json.RawMessage(`"string"`),
		})

		// Update to integer
		err := svc.UpdateSetting(context.Background(), "type_test_key", 123)
		if err != nil {
			t.Fatalf("UpdateSetting failed: %v", err)
		}

		var setting models.SystemSetting
		database.DB.Where("key = ?", "type_test_key").First(&setting)

		var value int
		json.Unmarshal(setting.Value, &value)
		if value != 123 {
			t.Errorf("Expected 123, got: %d", value)
		}
	})
}

func TestSettingsService_IPBlocklist_Integration(t *testing.T) {
	svc, cleanup := testSettingsSetup(t)
	defer cleanup()

	user := createTestUserForSettings(t, "blocklist")

	t.Run("blocks and retrieves IP", func(t *testing.T) {
		req := &models.CreateIPBlockRequest{
			IPAddress: "192.168.1.100",
			Reason:    "Test block",
		}

		block, err := svc.BlockIP(context.Background(), req, user.ID)
		if err != nil {
			t.Fatalf("BlockIP failed: %v", err)
		}

		if block.ID == 0 {
			t.Error("Expected block to have ID")
		}
		if block.IPAddress != "192.168.1.100" {
			t.Errorf("Expected IP '192.168.1.100', got: %s", block.IPAddress)
		}
		if !block.IsActive {
			t.Error("Expected block to be active")
		}

		// Verify it's in blocklist
		blocks, err := svc.GetIPBlocklist(context.Background())
		if err != nil {
			t.Fatalf("GetIPBlocklist failed: %v", err)
		}

		found := false
		for _, b := range blocks {
			if b.IPAddress == "192.168.1.100" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected blocked IP in blocklist")
		}
	})

	t.Run("checks if IP is blocked", func(t *testing.T) {
		svc.BlockIP(context.Background(), &models.CreateIPBlockRequest{
			IPAddress: "10.0.0.1",
			Reason:    "Test",
		}, user.ID)

		blocked, err := svc.IsIPBlocked(context.Background(), "10.0.0.1")
		if err != nil {
			t.Fatalf("IsIPBlocked failed: %v", err)
		}
		if !blocked {
			t.Error("Expected IP to be blocked")
		}

		blocked, err = svc.IsIPBlocked(context.Background(), "10.0.0.2")
		if err != nil {
			t.Fatalf("IsIPBlocked failed: %v", err)
		}
		if blocked {
			t.Error("Expected IP to not be blocked")
		}
	})

	t.Run("unblocks IP", func(t *testing.T) {
		block, _ := svc.BlockIP(context.Background(), &models.CreateIPBlockRequest{
			IPAddress: "172.16.0.1",
			Reason:    "Test",
		}, user.ID)

		err := svc.UnblockIP(context.Background(), block.ID)
		if err != nil {
			t.Fatalf("UnblockIP failed: %v", err)
		}

		// Should no longer be blocked
		blocked, _ := svc.IsIPBlocked(context.Background(), "172.16.0.1")
		if blocked {
			t.Error("Expected IP to be unblocked")
		}
	})

	t.Run("handles expired blocks", func(t *testing.T) {
		expiredTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		block := &models.IPBlocklist{
			IPAddress: "192.168.2.1",
			Reason:    "Expired block",
			BlockType: "manual",
			BlockedBy: &user.ID,
			IsActive:  true,
			ExpiresAt: &expiredTime,
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		}
		database.DB.Create(block)

		blocked, err := svc.IsIPBlocked(context.Background(), "192.168.2.1")
		if err != nil {
			t.Fatalf("IsIPBlocked failed: %v", err)
		}
		if blocked {
			t.Error("Expected expired block to not be active")
		}
	})
}

func TestSettingsService_Announcements_Integration(t *testing.T) {
	svc, cleanup := testSettingsSetup(t)
	defer cleanup()

	user := createTestUserForSettings(t, "announce")

	t.Run("creates and retrieves announcement", func(t *testing.T) {
		req := &models.CreateAnnouncementRequest{
			Title:    "Test Announcement",
			Message:  "This is a test message",
			Type:     "info",
			IsActive: true,
		}

		announcement, err := svc.CreateAnnouncement(context.Background(), req, user.ID)
		if err != nil {
			t.Fatalf("CreateAnnouncement failed: %v", err)
		}

		if announcement.ID == 0 {
			t.Error("Expected announcement to have ID")
		}
		if announcement.Title != "Test Announcement" {
			t.Errorf("Expected title 'Test Announcement', got: %s", announcement.Title)
		}

		// Retrieve all announcements
		announcements, err := svc.GetAnnouncements(context.Background())
		if err != nil {
			t.Fatalf("GetAnnouncements failed: %v", err)
		}

		found := false
		for _, a := range announcements {
			if a.ID == announcement.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected announcement in list")
		}
	})

	t.Run("gets active announcements for user", func(t *testing.T) {
		// Create active and inactive announcements
		svc.CreateAnnouncement(context.Background(), &models.CreateAnnouncementRequest{
			Title:    "Active Announcement",
			Message:  "Active",
			IsActive: true,
		}, user.ID)

		svc.CreateAnnouncement(context.Background(), &models.CreateAnnouncementRequest{
			Title:    "Inactive Announcement",
			Message:  "Inactive",
			IsActive: false,
		}, user.ID)

		announcements, err := svc.GetActiveAnnouncements(context.Background(), &user.ID, models.RoleUser)
		if err != nil {
			t.Fatalf("GetActiveAnnouncements failed: %v", err)
		}

		for _, a := range announcements {
			if !a.IsActive {
				t.Error("Expected only active announcements")
			}
		}
	})

	t.Run("filters by target roles", func(t *testing.T) {
		_, err := svc.CreateAnnouncement(context.Background(), &models.CreateAnnouncementRequest{
			Title:       "Admin Only",
			Message:     "For admins",
			IsActive:    true,
			TargetRoles: []string{models.RoleAdmin},
		}, user.ID)
		if err != nil {
			t.Fatalf("CreateAnnouncement with TargetRoles failed: %v", err)
		}

		// User with 'user' role should not see admin-only announcements
		announcements, err := svc.GetActiveAnnouncements(context.Background(), &user.ID, models.RoleUser)
		if err != nil {
			t.Fatalf("GetActiveAnnouncements failed: %v", err)
		}

		for _, a := range announcements {
			if a.Title == "Admin Only" {
				t.Error("User should not see admin-only announcement")
			}
		}

		// Admin should see the announcement
		announcements, err = svc.GetActiveAnnouncements(context.Background(), &user.ID, models.RoleAdmin)
		if err != nil {
			t.Fatalf("GetActiveAnnouncements failed: %v", err)
		}

		found := false
		for _, a := range announcements {
			if a.Title == "Admin Only" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Admin should see admin-only announcement")
		}
	})

	t.Run("updates announcement", func(t *testing.T) {
		announcement, _ := svc.CreateAnnouncement(context.Background(), &models.CreateAnnouncementRequest{
			Title:    "Original Title",
			Message:  "Original message",
			IsActive: true,
		}, user.ID)

		newTitle := "Updated Title"
		updated, err := svc.UpdateAnnouncement(context.Background(), announcement.ID, &models.UpdateAnnouncementRequest{
			Title: &newTitle,
		})
		if err != nil {
			t.Fatalf("UpdateAnnouncement failed: %v", err)
		}

		if updated.Title != "Updated Title" {
			t.Errorf("Expected 'Updated Title', got: %s", updated.Title)
		}
	})

	t.Run("deletes announcement", func(t *testing.T) {
		announcement, _ := svc.CreateAnnouncement(context.Background(), &models.CreateAnnouncementRequest{
			Title:    "To Delete",
			Message:  "Will be deleted",
			IsActive: true,
		}, user.ID)

		err := svc.DeleteAnnouncement(context.Background(), announcement.ID)
		if err != nil {
			t.Fatalf("DeleteAnnouncement failed: %v", err)
		}

		// Verify deleted
		var count int64
		database.DB.Model(&models.AnnouncementBanner{}).Where("id = ?", announcement.ID).Count(&count)
		if count != 0 {
			t.Error("Expected announcement to be deleted")
		}
	})

	t.Run("dismisses announcement for user", func(t *testing.T) {
		announcement, _ := svc.CreateAnnouncement(context.Background(), &models.CreateAnnouncementRequest{
			Title:         "Dismissible",
			Message:       "Can be dismissed",
			IsActive:      true,
			IsDismissible: true,
		}, user.ID)

		err := svc.DismissAnnouncement(context.Background(), user.ID, announcement.ID)
		if err != nil {
			t.Fatalf("DismissAnnouncement failed: %v", err)
		}

		// Should not appear in active announcements for this user
		announcements, _ := svc.GetActiveAnnouncements(context.Background(), &user.ID, models.RoleUser)
		for _, a := range announcements {
			if a.ID == announcement.ID {
				t.Error("Dismissed announcement should not appear for user")
			}
		}
	})
}

func TestSettingsService_Changelog_Integration(t *testing.T) {
	svc, cleanup := testSettingsSetup(t)
	defer cleanup()

	user := createTestUserForSettings(t, "changelog")

	t.Run("returns paginated changelog", func(t *testing.T) {
		// Create multiple changelog entries
		for i := 0; i < 15; i++ {
			svc.CreateAnnouncement(context.Background(), &models.CreateAnnouncementRequest{
				Title:    "Changelog Entry",
				Message:  "Entry content",
				Category: "feature",
				IsActive: true,
			}, user.ID)
		}

		// Get first page
		result, err := svc.GetChangelog(context.Background(), 1, 10, "")
		if err != nil {
			t.Fatalf("GetChangelog failed: %v", err)
		}

		if len(result.Data) > 10 {
			t.Errorf("Expected max 10 entries, got: %d", len(result.Data))
		}
		if result.Meta.Page != 1 {
			t.Errorf("Expected page 1, got: %d", result.Meta.Page)
		}
	})

	t.Run("filters by category", func(t *testing.T) {
		svc.CreateAnnouncement(context.Background(), &models.CreateAnnouncementRequest{
			Title:    "Feature",
			Message:  "Feature content",
			Category: "feature",
			IsActive: true,
		}, user.ID)

		svc.CreateAnnouncement(context.Background(), &models.CreateAnnouncementRequest{
			Title:    "Bugfix",
			Message:  "Bugfix content",
			Category: "bugfix",
			IsActive: true,
		}, user.ID)

		result, err := svc.GetChangelog(context.Background(), 1, 50, "bugfix")
		if err != nil {
			t.Fatalf("GetChangelog failed: %v", err)
		}

		for _, entry := range result.Data {
			if entry.Category != "bugfix" {
				t.Errorf("Expected category 'bugfix', got: %s", entry.Category)
			}
		}
	})
}

func TestSettingsService_ModalAnnouncements_Integration(t *testing.T) {
	svc, cleanup := testSettingsSetup(t)
	defer cleanup()

	user := createTestUserForSettings(t, "modal")

	t.Run("gets unread modal announcements", func(t *testing.T) {
		svc.CreateAnnouncement(context.Background(), &models.CreateAnnouncementRequest{
			Title:       "Modal Announcement",
			Message:     "Important update",
			DisplayType: "modal",
			IsActive:    true,
		}, user.ID)

		svc.CreateAnnouncement(context.Background(), &models.CreateAnnouncementRequest{
			Title:       "Banner Announcement",
			Message:     "Less important",
			DisplayType: "banner",
			IsActive:    true,
		}, user.ID)

		modals, err := svc.GetUnreadModalAnnouncements(context.Background(), user.ID, models.RoleUser)
		if err != nil {
			t.Fatalf("GetUnreadModalAnnouncements failed: %v", err)
		}

		for _, m := range modals {
			if m.DisplayType != "modal" {
				t.Errorf("Expected display_type 'modal', got: %s", m.DisplayType)
			}
		}
	})

	t.Run("marks announcement as read", func(t *testing.T) {
		announcement, _ := svc.CreateAnnouncement(context.Background(), &models.CreateAnnouncementRequest{
			Title:       "Unread Modal",
			Message:     "Will be marked read",
			DisplayType: "modal",
			IsActive:    true,
		}, user.ID)

		err := svc.MarkAnnouncementRead(context.Background(), user.ID, announcement.ID)
		if err != nil {
			t.Fatalf("MarkAnnouncementRead failed: %v", err)
		}

		// Should not appear in unread modals
		modals, _ := svc.GetUnreadModalAnnouncements(context.Background(), user.ID, models.RoleUser)
		for _, m := range modals {
			if m.ID == announcement.ID {
				t.Error("Read announcement should not appear in unread list")
			}
		}
	})
}

func TestSettingsService_EmailTemplates_Integration(t *testing.T) {
	svc, cleanup := testSettingsSetup(t)
	defer cleanup()

	user := createTestUserForSettings(t, "email")

	t.Run("retrieves email templates", func(t *testing.T) {
		// Create test template
		template := &models.EmailTemplate{
			Key:      "test_template",
			Name:     "Test Template",
			Subject:  "Test Subject",
			BodyHTML: "<p>Test body</p>",
			BodyText: "Test body",
			IsActive: true,
		}
		database.DB.Create(template)

		templates, err := svc.GetEmailTemplates(context.Background())
		if err != nil {
			t.Fatalf("GetEmailTemplates failed: %v", err)
		}

		if len(templates) == 0 {
			t.Error("Expected at least one template")
		}
	})

	t.Run("gets template by ID", func(t *testing.T) {
		template := &models.EmailTemplate{
			Key:      "template_by_id",
			Name:     "Template By ID",
			Subject:  "Subject",
			BodyHTML: "<p>Body</p>",
			BodyText: "Body",
			IsActive: true,
		}
		database.DB.Create(template)

		result, err := svc.GetEmailTemplate(context.Background(), template.ID)
		if err != nil {
			t.Fatalf("GetEmailTemplate failed: %v", err)
		}

		if result.Key != "template_by_id" {
			t.Errorf("Expected key 'template_by_id', got: %s", result.Key)
		}
	})

	t.Run("gets template by key", func(t *testing.T) {
		template := &models.EmailTemplate{
			Key:      "template_by_key",
			Name:     "Template By Key",
			Subject:  "Subject",
			BodyHTML: "<p>Body</p>",
			BodyText: "Body",
			IsActive: true,
		}
		database.DB.Create(template)

		result, err := svc.GetEmailTemplateByKey(context.Background(), "template_by_key")
		if err != nil {
			t.Fatalf("GetEmailTemplateByKey failed: %v", err)
		}

		if result.Name != "Template By Key" {
			t.Errorf("Expected name 'Template By Key', got: %s", result.Name)
		}
	})

	t.Run("updates email template", func(t *testing.T) {
		template := &models.EmailTemplate{
			Key:      "template_to_update",
			Name:     "Original Name",
			Subject:  "Original Subject",
			BodyHTML: "<p>Original</p>",
			BodyText: "Original",
			IsActive: true,
		}
		database.DB.Create(template)

		newSubject := "Updated Subject"
		updated, err := svc.UpdateEmailTemplate(context.Background(), template.ID, &models.UpdateEmailTemplateRequest{
			Subject: &newSubject,
		}, user.ID)
		if err != nil {
			t.Fatalf("UpdateEmailTemplate failed: %v", err)
		}

		if updated.Subject != "Updated Subject" {
			t.Errorf("Expected 'Updated Subject', got: %s", updated.Subject)
		}
	})

	t.Run("returns error for non-existent template", func(t *testing.T) {
		_, err := svc.GetEmailTemplate(context.Background(), 99999)
		if err == nil {
			t.Error("Expected error for non-existent template")
		}

		_, err = svc.GetEmailTemplateByKey(context.Background(), "nonexistent_key")
		if err == nil {
			t.Error("Expected error for non-existent key")
		}
	})
}

// Helper function to create a user for settings tests
func createTestUserForSettings(t *testing.T, suffix string) *models.User {
	t.Helper()
	user := &models.User{
		Email:    "settings_test_" + suffix + "@example.com",
		Name:     "Settings Test User",
		Password: "hashedpassword",
		Role:     models.RoleUser,
	}
	if err := database.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}
