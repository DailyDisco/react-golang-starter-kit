package jobs

import (
	"context"
	"os"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/storage"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// StartExportCleanup starts a background goroutine that periodically cleans up expired exports
// Call this from main.go after database initialization
func StartExportCleanup(ctx context.Context, db *gorm.DB, interval time.Duration) {
	go func() {
		// Initial delay to allow system startup
		time.Sleep(30 * time.Second)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run immediately on start
		if err := RunExportCleanup(ctx, db); err != nil {
			log.Error().Err(err).Msg("initial export cleanup failed")
		}

		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("export cleanup job shutting down")
				return
			case <-ticker.C:
				if err := RunExportCleanup(ctx, db); err != nil {
					log.Error().Err(err).Msg("export cleanup failed")
				}
			}
		}
	}()

	log.Info().Dur("interval", interval).Msg("export cleanup job started")
}

// RunExportCleanup finds and cleans up expired export files
func RunExportCleanup(ctx context.Context, db *gorm.DB) error {
	// Find exports that:
	// 1. Have expired (expires_at < now - 24 hours grace period)
	// 2. Still have a file_path set (not yet cleaned up)
	gracePeriod := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

	var expiredExports []models.DataExport
	if err := db.WithContext(ctx).
		Where("expires_at < ? AND file_path IS NOT NULL AND file_path != ''", gracePeriod).
		Find(&expiredExports).Error; err != nil {
		return err
	}

	if len(expiredExports) == 0 {
		log.Debug().Msg("no expired exports to clean up")
		return nil
	}

	log.Info().Int("count", len(expiredExports)).Msg("cleaning up expired exports")

	// Initialize S3 storage (may not be available)
	s3Storage, _ := storage.NewS3Storage()

	var cleanedCount, errorCount int

	for _, export := range expiredExports {
		if export.FilePath == nil || *export.FilePath == "" {
			continue
		}

		var deleteErr error

		// Delete file based on storage type
		if export.StorageType == "s3" && s3Storage != nil && s3Storage.IsAvailable() {
			deleteErr = s3Storage.DeleteFileWithKey(ctx, *export.FilePath)
		} else if export.StorageType == "local" || export.StorageType == "" {
			deleteErr = os.Remove(*export.FilePath)
			// Ignore "file not found" errors for local files
			if os.IsNotExist(deleteErr) {
				deleteErr = nil
			}
		}

		if deleteErr != nil {
			log.Warn().
				Err(deleteErr).
				Uint("export_id", export.ID).
				Str("storage_type", export.StorageType).
				Str("file_path", *export.FilePath).
				Msg("failed to delete export file")
			errorCount++
			continue
		}

		// Clear the file_path and update status
		if err := db.Model(&export).Updates(map[string]interface{}{
			"file_path":    nil,
			"status":       models.ExportStatusExpired,
			"storage_type": "",
		}).Error; err != nil {
			log.Warn().Err(err).Uint("export_id", export.ID).Msg("failed to update export record")
			errorCount++
			continue
		}

		cleanedCount++
		log.Debug().
			Uint("export_id", export.ID).
			Str("storage_type", export.StorageType).
			Msg("cleaned up expired export")
	}

	log.Info().
		Int("cleaned", cleanedCount).
		Int("errors", errorCount).
		Int("total", len(expiredExports)).
		Msg("export cleanup completed")

	return nil
}

// CleanupExportFile manually cleans up a specific export's file
// Useful for immediate cleanup when a user deletes their account
func CleanupExportFile(ctx context.Context, export *models.DataExport) error {
	if export.FilePath == nil || *export.FilePath == "" {
		return nil
	}

	var deleteErr error

	if export.StorageType == "s3" {
		s3Storage, err := storage.NewS3Storage()
		if err == nil && s3Storage.IsAvailable() {
			deleteErr = s3Storage.DeleteFileWithKey(ctx, *export.FilePath)
		}
	} else {
		deleteErr = os.Remove(*export.FilePath)
		if os.IsNotExist(deleteErr) {
			deleteErr = nil
		}
	}

	if deleteErr != nil {
		return deleteErr
	}

	// Clear the file_path
	return database.DB.Model(export).Updates(map[string]interface{}{
		"file_path":    nil,
		"storage_type": "",
	}).Error
}
