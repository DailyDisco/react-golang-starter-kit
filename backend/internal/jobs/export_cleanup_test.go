package jobs

import (
	"context"
	"testing"

	"react-golang-starter/internal/models"
)

// ============ CleanupExportFile Tests ============

func TestCleanupExportFile_NilFilePath(t *testing.T) {
	export := &models.DataExport{
		FilePath: nil,
	}

	err := CleanupExportFile(context.Background(), export)
	if err != nil {
		t.Errorf("CleanupExportFile() with nil FilePath = %v, want nil", err)
	}
}

func TestCleanupExportFile_EmptyFilePath(t *testing.T) {
	empty := ""
	export := &models.DataExport{
		FilePath: &empty,
	}

	err := CleanupExportFile(context.Background(), export)
	if err != nil {
		t.Errorf("CleanupExportFile() with empty FilePath = %v, want nil", err)
	}
}

// ============ DataExport Model Tests ============

func TestDataExport_StorageTypes(t *testing.T) {
	tests := []struct {
		name        string
		storageType string
	}{
		{"local storage", "local"},
		{"s3 storage", "s3"},
		{"empty storage type", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			export := &models.DataExport{
				StorageType: tt.storageType,
			}
			if export.StorageType != tt.storageType {
				t.Errorf("StorageType = %q, want %q", export.StorageType, tt.storageType)
			}
		})
	}
}

func TestDataExport_ExportStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{"pending status", models.ExportStatusPending},
		{"processing status", models.ExportStatusProcessing},
		{"completed status", models.ExportStatusCompleted},
		{"failed status", models.ExportStatusFailed},
		{"expired status", models.ExportStatusExpired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			export := &models.DataExport{
				Status: tt.status,
			}
			if export.Status != tt.status {
				t.Errorf("Status = %v, want %v", export.Status, tt.status)
			}
		})
	}
}

// ============ Export Status Constants Tests ============

func TestExportStatusConstants(t *testing.T) {
	// Verify status constants are distinct and non-empty
	statuses := []string{
		models.ExportStatusPending,
		models.ExportStatusProcessing,
		models.ExportStatusCompleted,
		models.ExportStatusFailed,
		models.ExportStatusExpired,
	}

	seen := make(map[string]bool)
	for _, status := range statuses {
		if status == "" {
			t.Error("Export status should not be empty")
		}
		if seen[status] {
			t.Errorf("Duplicate export status: %v", status)
		}
		seen[status] = true
	}
}
