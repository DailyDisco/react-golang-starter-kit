package storage

import (
	"context"
	"testing"
)

// ============ S3Storage Tests ============

func TestNewS3Storage_MissingConfig(t *testing.T) {
	// Clear any existing env vars
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_S3_BUCKET", "")

	_, err := NewS3Storage()
	if err == nil {
		t.Error("NewS3Storage() should return error when AWS config is missing")
	}
}

func TestNewS3Storage_PartialConfig(t *testing.T) {
	// Set only some env vars
	t.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_S3_BUCKET", "")

	_, err := NewS3Storage()
	if err == nil {
		t.Error("NewS3Storage() should return error when AWS config is partial")
	}
}

func TestS3Storage_IsAvailable_NilClient(t *testing.T) {
	s := &S3Storage{
		client: nil,
	}

	if s.IsAvailable() {
		t.Error("IsAvailable() should return false when client is nil")
	}
}

func TestS3Storage_GetFileURLWithKey(t *testing.T) {
	s := &S3Storage{
		baseURL: "https://test-bucket.s3.us-east-1.amazonaws.com",
	}

	tests := []struct {
		name     string
		s3Key    string
		expected string
	}{
		{
			name:     "simple key",
			s3Key:    "uploads/file.pdf",
			expected: "https://test-bucket.s3.us-east-1.amazonaws.com/uploads/file.pdf",
		},
		{
			name:     "nested key",
			s3Key:    "uploads/2025/01/document.pdf",
			expected: "https://test-bucket.s3.us-east-1.amazonaws.com/uploads/2025/01/document.pdf",
		},
		{
			name:     "key with special chars",
			s3Key:    "exports/user_123/data_export.zip",
			expected: "https://test-bucket.s3.us-east-1.amazonaws.com/exports/user_123/data_export.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.GetFileURLWithKey(tt.s3Key)
			if result != tt.expected {
				t.Errorf("GetFileURLWithKey(%q) = %q, want %q", tt.s3Key, result, tt.expected)
			}
		})
	}
}

func TestS3Storage_MethodsReturnErrorWhenNotAvailable(t *testing.T) {
	s := &S3Storage{
		client: nil,
	}
	ctx := context.Background()

	// Test UploadFile
	_, err := s.UploadFile(ctx, nil, nil)
	if err == nil {
		t.Error("UploadFile() should return error when S3 not available")
	}

	// Test DownloadFile
	_, err = s.DownloadFile(ctx, 1)
	if err == nil {
		t.Error("DownloadFile() should return error when S3 not available")
	}

	// Test DeleteFile
	err = s.DeleteFile(ctx, 1)
	if err == nil {
		t.Error("DeleteFile() should return error when S3 not available")
	}

	// Test GetFileURL
	_, err = s.GetFileURL(ctx, 1)
	if err == nil {
		t.Error("GetFileURL() should return error when S3 not available")
	}

	// Test DeleteFileWithKey
	err = s.DeleteFileWithKey(ctx, "test-key")
	if err == nil {
		t.Error("DeleteFileWithKey() should return error when S3 not available")
	}

	// Test UploadBytes
	err = s.UploadBytes(ctx, "test-key", []byte("data"), "text/plain")
	if err == nil {
		t.Error("UploadBytes() should return error when S3 not available")
	}

	// Test GeneratePresignedURL
	_, err = s.GeneratePresignedURL(ctx, "test-key", 0)
	if err == nil {
		t.Error("GeneratePresignedURL() should return error when S3 not available")
	}
}

// ============ DatabaseStorage Tests ============

func TestNewDatabaseStorage(t *testing.T) {
	storage := NewDatabaseStorage()

	if storage == nil {
		t.Error("NewDatabaseStorage() should not return nil")
	}
}

func TestDatabaseStorage_IsAvailable_NilDB(t *testing.T) {
	storage := &DatabaseStorage{
		db: nil,
	}

	if storage.IsAvailable() {
		t.Error("IsAvailable() should return false when db is nil")
	}
}

func TestDatabaseStorage_MethodsReturnErrorWhenNotAvailable(t *testing.T) {
	storage := &DatabaseStorage{
		db: nil,
	}
	ctx := context.Background()

	// Test UploadFile
	_, err := storage.UploadFile(ctx, nil, nil)
	if err == nil {
		t.Error("UploadFile() should return error when db not available")
	}

	// Test DownloadFile
	_, err = storage.DownloadFile(ctx, 1)
	if err == nil {
		t.Error("DownloadFile() should return error when db not available")
	}

	// Test DeleteFile
	err = storage.DeleteFile(ctx, 1)
	if err == nil {
		t.Error("DeleteFile() should return error when db not available")
	}
}

func TestDatabaseStorage_GetFileURL(t *testing.T) {
	storage := &DatabaseStorage{
		db: nil, // db not needed for this method
	}
	ctx := context.Background()

	tests := []struct {
		name     string
		fileID   uint
		expected string
	}{
		{"file ID 1", 1, "/api/files/1/download"},
		{"file ID 42", 42, "/api/files/42/download"},
		{"file ID 999", 999, "/api/files/999/download"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := storage.GetFileURL(ctx, tt.fileID)
			if err != nil {
				t.Fatalf("GetFileURL(%d) error = %v", tt.fileID, err)
			}
			if result != tt.expected {
				t.Errorf("GetFileURL(%d) = %q, want %q", tt.fileID, result, tt.expected)
			}
		})
	}
}

// ============ FileStorage Interface Tests ============

func TestS3Storage_ImplementsFileStorageInterface(t *testing.T) {
	var _ FileStorage = (*S3Storage)(nil)
}

func TestDatabaseStorage_ImplementsFileStorageInterface(t *testing.T) {
	var _ FileStorage = (*DatabaseStorage)(nil)
}

// ============ S3Storage Structure Tests ============

func TestS3Storage_Structure(t *testing.T) {
	s := &S3Storage{
		bucketName: "test-bucket",
		region:     "us-west-2",
		baseURL:    "https://test-bucket.s3.us-west-2.amazonaws.com",
	}

	if s.bucketName != "test-bucket" {
		t.Errorf("bucketName = %q, want 'test-bucket'", s.bucketName)
	}

	if s.region != "us-west-2" {
		t.Errorf("region = %q, want 'us-west-2'", s.region)
	}

	if s.baseURL != "https://test-bucket.s3.us-west-2.amazonaws.com" {
		t.Errorf("baseURL = %q, want 'https://test-bucket.s3.us-west-2.amazonaws.com'", s.baseURL)
	}
}

// ============ Error Message Tests ============

func TestS3Storage_ErrorMessages(t *testing.T) {
	s := &S3Storage{client: nil}
	ctx := context.Background()

	// Test that error messages are descriptive
	_, err := s.UploadFile(ctx, nil, nil)
	if err == nil || err.Error() != "S3 storage not available" {
		t.Error("UploadFile error should mention S3 storage not available")
	}

	_, err = s.DownloadFile(ctx, 1)
	if err == nil || err.Error() != "S3 storage not available" {
		t.Error("DownloadFile error should mention S3 storage not available")
	}
}

func TestDatabaseStorage_ErrorMessages(t *testing.T) {
	storage := &DatabaseStorage{db: nil}
	ctx := context.Background()

	// Test that error messages are descriptive
	_, err := storage.UploadFile(ctx, nil, nil)
	if err == nil || err.Error() != "database storage not available" {
		t.Error("UploadFile error should mention database storage not available")
	}

	_, err = storage.DownloadFile(ctx, 1)
	if err == nil || err.Error() != "database storage not available" {
		t.Error("DownloadFile error should mention database storage not available")
	}
}
