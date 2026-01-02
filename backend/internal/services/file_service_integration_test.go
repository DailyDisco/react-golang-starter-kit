package services

import (
	"context"
	"testing"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/storage"
	"react-golang-starter/internal/testutil"

	"gorm.io/gorm"
)

func testFileServiceSetup(t *testing.T) (*FileService, *gorm.DB, func()) {
	t.Helper()
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tt := testutil.NewTestTransaction(t, db)

	// Set global database.DB
	oldDB := database.DB
	database.DB = tt.DB

	// Create file service with database storage only (no S3 in tests)
	dbStorage := storage.NewDatabaseStorage()

	svc := &FileService{
		db:            tt.DB,
		dbStorage:     dbStorage,
		activeStorage: dbStorage,
	}

	return svc, tt.DB, func() {
		database.DB = oldDB
		tt.Rollback()
	}
}

func createTestUserForFiles(t *testing.T, db *gorm.DB, email string) *models.User {
	t.Helper()
	user := &models.User{
		Email:    email,
		Name:     "File Test User",
		Password: "hashedpassword",
		Role:     models.RoleUser,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func createTestFile(t *testing.T, db *gorm.DB, userID uint, fileName string, content []byte) *models.File {
	t.Helper()
	file := &models.File{
		UserID:      userID,
		FileName:    fileName,
		ContentType: "text/plain",
		FileSize:    int64(len(content)),
		Content:     content,
		StorageType: "database",
		Location:    "database://files",
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}
	if err := db.Create(file).Error; err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return file
}

func TestFileService_GetFileByID_Integration(t *testing.T) {
	svc, db, cleanup := testFileServiceSetup(t)
	defer cleanup()

	t.Run("returns file metadata", func(t *testing.T) {
		user := createTestUserForFiles(t, db, "getfile@example.com")
		file := createTestFile(t, db, user.ID, "test.txt", []byte("hello world"))

		result, err := svc.GetFileByID(file.ID)
		if err != nil {
			t.Fatalf("GetFileByID failed: %v", err)
		}

		if result.ID != file.ID {
			t.Errorf("Expected ID %d, got: %d", file.ID, result.ID)
		}
		if result.FileName != "test.txt" {
			t.Errorf("Expected filename 'test.txt', got: %s", result.FileName)
		}
		if result.UserID != user.ID {
			t.Errorf("Expected UserID %d, got: %d", user.ID, result.UserID)
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := svc.GetFileByID(99999)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
		if err.Error() != "file not found" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}

func TestFileService_GetFileByIDForUser_Integration(t *testing.T) {
	svc, db, cleanup := testFileServiceSetup(t)
	defer cleanup()

	t.Run("owner can access own file", func(t *testing.T) {
		user := createTestUserForFiles(t, db, "owner@example.com")
		file := createTestFile(t, db, user.ID, "myfile.txt", []byte("content"))

		result, err := svc.GetFileByIDForUser(file.ID, user.ID, false)
		if err != nil {
			t.Fatalf("GetFileByIDForUser failed: %v", err)
		}

		if result.ID != file.ID {
			t.Errorf("Expected file ID %d, got: %d", file.ID, result.ID)
		}
	})

	t.Run("non-owner denied access", func(t *testing.T) {
		owner := createTestUserForFiles(t, db, "owner2@example.com")
		otherUser := createTestUserForFiles(t, db, "other@example.com")
		file := createTestFile(t, db, owner.ID, "private.txt", []byte("secret"))

		_, err := svc.GetFileByIDForUser(file.ID, otherUser.ID, false)
		if err == nil {
			t.Error("Expected access denied error")
		}
		if err.Error() != "access denied: you do not own this file" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("admin can access any file", func(t *testing.T) {
		owner := createTestUserForFiles(t, db, "owner3@example.com")
		admin := createTestUserForFiles(t, db, "admin@example.com")
		admin.Role = models.RoleAdmin
		db.Save(admin)

		file := createTestFile(t, db, owner.ID, "admintest.txt", []byte("content"))

		result, err := svc.GetFileByIDForUser(file.ID, admin.ID, true)
		if err != nil {
			t.Fatalf("Admin should be able to access any file: %v", err)
		}

		if result.ID != file.ID {
			t.Error("Expected to get the file")
		}
	})

	t.Run("legacy files with no owner are accessible", func(t *testing.T) {
		// Create a file with UserID = 0 (legacy/no owner)
		file := &models.File{
			UserID:      0, // No owner
			FileName:    "legacy.txt",
			ContentType: "text/plain",
			FileSize:    5,
			Content:     []byte("hello"),
			StorageType: "database",
			Location:    "database://files",
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		}
		db.Create(file)

		randomUser := createTestUserForFiles(t, db, "random@example.com")

		result, err := svc.GetFileByIDForUser(file.ID, randomUser.ID, false)
		if err != nil {
			t.Fatalf("Legacy files should be accessible: %v", err)
		}

		if result.ID != file.ID {
			t.Error("Expected to get the legacy file")
		}
	})
}

func TestFileService_ListFiles_Integration(t *testing.T) {
	svc, db, cleanup := testFileServiceSetup(t)
	defer cleanup()

	t.Run("returns paginated results", func(t *testing.T) {
		user := createTestUserForFiles(t, db, "list@example.com")

		// Create 5 files
		for i := 0; i < 5; i++ {
			createTestFile(t, db, user.ID, "file"+string(rune('0'+i))+".txt", []byte("content"))
		}

		// Get first page
		files, err := svc.ListFiles(2, 0)
		if err != nil {
			t.Fatalf("ListFiles failed: %v", err)
		}

		if len(files) != 2 {
			t.Errorf("Expected 2 files, got: %d", len(files))
		}

		// Get second page
		files, err = svc.ListFiles(2, 2)
		if err != nil {
			t.Fatalf("ListFiles page 2 failed: %v", err)
		}

		if len(files) != 2 {
			t.Errorf("Expected 2 files on page 2, got: %d", len(files))
		}

		// Get third page (should have 1)
		files, err = svc.ListFiles(2, 4)
		if err != nil {
			t.Fatalf("ListFiles page 3 failed: %v", err)
		}

		if len(files) != 1 {
			t.Errorf("Expected 1 file on page 3, got: %d", len(files))
		}
	})
}

func TestFileService_ListFilesForUser_Integration(t *testing.T) {
	svc, db, cleanup := testFileServiceSetup(t)
	defer cleanup()

	t.Run("returns only users files", func(t *testing.T) {
		user1 := createTestUserForFiles(t, db, "user1@example.com")
		user2 := createTestUserForFiles(t, db, "user2@example.com")

		// Create files for user1
		createTestFile(t, db, user1.ID, "user1-file1.txt", []byte("a"))
		createTestFile(t, db, user1.ID, "user1-file2.txt", []byte("b"))

		// Create files for user2
		createTestFile(t, db, user2.ID, "user2-file1.txt", []byte("c"))

		files, err := svc.ListFilesForUser(user1.ID, false, 10, 0)
		if err != nil {
			t.Fatalf("ListFilesForUser failed: %v", err)
		}

		if len(files) != 2 {
			t.Errorf("Expected 2 files for user1, got: %d", len(files))
		}

		for _, f := range files {
			if f.UserID != user1.ID && f.UserID != 0 {
				t.Errorf("Got file belonging to wrong user: %d", f.UserID)
			}
		}
	})

	t.Run("admin sees all files", func(t *testing.T) {
		user1 := createTestUserForFiles(t, db, "admintest1@example.com")
		user2 := createTestUserForFiles(t, db, "admintest2@example.com")
		admin := createTestUserForFiles(t, db, "adminlist@example.com")
		admin.Role = models.RoleAdmin
		db.Save(admin)

		createTestFile(t, db, user1.ID, "u1.txt", []byte("a"))
		createTestFile(t, db, user2.ID, "u2.txt", []byte("b"))

		files, err := svc.ListFilesForUser(admin.ID, true, 100, 0)
		if err != nil {
			t.Fatalf("ListFilesForUser (admin) failed: %v", err)
		}

		// Admin should see files from multiple users
		userIDs := make(map[uint]bool)
		for _, f := range files {
			userIDs[f.UserID] = true
		}

		if len(userIDs) < 2 {
			t.Error("Admin should see files from multiple users")
		}
	})
}

func TestFileService_DeleteFile_Integration(t *testing.T) {
	svc, db, cleanup := testFileServiceSetup(t)
	defer cleanup()

	t.Run("deletes file from database", func(t *testing.T) {
		user := createTestUserForFiles(t, db, "delete@example.com")
		file := createTestFile(t, db, user.ID, "todelete.txt", []byte("content"))

		err := svc.DeleteFile(context.Background(), file.ID)
		if err != nil {
			t.Fatalf("DeleteFile failed: %v", err)
		}

		// Verify file is deleted (soft delete)
		var count int64
		db.Unscoped().Model(&models.File{}).Where("id = ? AND deleted_at IS NOT NULL", file.ID).Count(&count)
		if count != 1 {
			t.Error("Expected file to be soft deleted")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		err := svc.DeleteFile(context.Background(), 99999)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestFileService_DownloadFile_Integration(t *testing.T) {
	svc, db, cleanup := testFileServiceSetup(t)
	defer cleanup()

	t.Run("downloads file content from database storage", func(t *testing.T) {
		user := createTestUserForFiles(t, db, "download@example.com")
		expectedContent := []byte("hello download test")
		file := createTestFile(t, db, user.ID, "download.txt", expectedContent)

		content, metadata, err := svc.DownloadFile(context.Background(), file.ID)
		if err != nil {
			t.Fatalf("DownloadFile failed: %v", err)
		}

		if string(content) != string(expectedContent) {
			t.Errorf("Expected content %q, got: %q", expectedContent, content)
		}

		if metadata.FileName != "download.txt" {
			t.Errorf("Expected filename 'download.txt', got: %s", metadata.FileName)
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, _, err := svc.DownloadFile(context.Background(), 99999)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestFileService_GetFileURL_Integration(t *testing.T) {
	svc, db, cleanup := testFileServiceSetup(t)
	defer cleanup()

	t.Run("returns URL for database file", func(t *testing.T) {
		user := createTestUserForFiles(t, db, "url@example.com")
		file := createTestFile(t, db, user.ID, "urltest.txt", []byte("content"))

		url, err := svc.GetFileURL(context.Background(), file.ID)
		if err != nil {
			t.Fatalf("GetFileURL failed: %v", err)
		}

		if url == "" {
			t.Error("Expected non-empty URL")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := svc.GetFileURL(context.Background(), 99999)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestFileService_GetStorageType_Integration(t *testing.T) {
	svc, _, cleanup := testFileServiceSetup(t)
	defer cleanup()

	t.Run("returns active storage type", func(t *testing.T) {
		storageType := svc.GetStorageType()

		// In test setup, we only use database storage
		if storageType != "database" {
			t.Errorf("Expected 'database', got: %s", storageType)
		}
	})
}

func TestFileService_ErrAccessDenied_Integration(t *testing.T) {
	_, db, cleanup := testFileServiceSetup(t)
	defer cleanup()

	t.Run("ErrAccessDenied is properly defined", func(t *testing.T) {
		if ErrAccessDenied == nil {
			t.Error("ErrAccessDenied should not be nil")
		}
		if ErrAccessDenied.Error() != "access denied" {
			t.Errorf("Unexpected error message: %s", ErrAccessDenied.Error())
		}
	})
	_ = db // Suppress unused variable
}
