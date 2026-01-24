package services

import (
	"testing"
)

// ============ File Service Error Tests ============

func TestErrAccessDenied(t *testing.T) {
	if ErrAccessDenied.Error() != "access denied" {
		t.Errorf("ErrAccessDenied.Error() = %q, want %q", ErrAccessDenied.Error(), "access denied")
	}
}

func TestErrAccessDenied_NotNil(t *testing.T) {
	if ErrAccessDenied == nil {
		t.Error("ErrAccessDenied should not be nil")
	}
}

// ============ FileService Structure Tests ============

func TestFileService_GetStorageType_NilStorage(t *testing.T) {
	// Test GetStorageType with nil s3Storage
	fs := &FileService{
		s3Storage: nil,
	}

	storageType := fs.GetStorageType()
	if storageType != "database" {
		t.Errorf("GetStorageType() = %q, want 'database' when s3Storage is nil", storageType)
	}
}

func TestFileService_Structure(t *testing.T) {
	// Test that FileService can be instantiated with nil values
	fs := &FileService{
		db:            nil,
		s3Storage:     nil,
		dbStorage:     nil,
		activeStorage: nil,
	}

	// Should not panic when accessing fields
	if fs.db != nil {
		t.Error("db should be nil")
	}
	if fs.s3Storage != nil {
		t.Error("s3Storage should be nil")
	}
	if fs.dbStorage != nil {
		t.Error("dbStorage should be nil")
	}
	if fs.activeStorage != nil {
		t.Error("activeStorage should be nil")
	}
}

// ============ FileService Storage Type Tests ============

func TestFileService_GetStorageType_ReturnsDatabase(t *testing.T) {
	// When s3Storage is nil and activeStorage is different from s3Storage
	fs := &FileService{
		s3Storage:     nil,
		activeStorage: nil,
	}

	// activeStorage != s3Storage (both nil, but different paths)
	storageType := fs.GetStorageType()
	if storageType != "database" {
		t.Errorf("GetStorageType() = %q, want 'database'", storageType)
	}
}
