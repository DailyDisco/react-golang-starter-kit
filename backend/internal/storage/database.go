package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"time"

	"gorm.io/gorm"
)

// DatabaseStorage implements FileStorage using database BLOB storage
type DatabaseStorage struct {
	db *gorm.DB
}

// NewDatabaseStorage creates a new database storage instance
func NewDatabaseStorage() *DatabaseStorage {
	return &DatabaseStorage{
		db: database.DB,
	}
}

// IsAvailable checks if database storage is available
func (d *DatabaseStorage) IsAvailable() bool {
	return d.db != nil
}

// UploadFile uploads a file to database storage
func (d *DatabaseStorage) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*models.File, error) {
	if !d.IsAvailable() {
		return nil, fmt.Errorf("database storage not available")
	}

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Create file record with content
	fileModel := &models.File{
		FileName:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		FileSize:    int64(len(fileContent)),
		Location:    fmt.Sprintf("db_file_%d", time.Now().UnixNano()), // Unique identifier for database storage
		Content:     fileContent,
		StorageType: "database",
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// Save to database
	if err := d.db.Create(fileModel).Error; err != nil {
		return nil, fmt.Errorf("failed to save file to database: %w", err)
	}

	return fileModel, nil
}

// DownloadFile downloads a file from database storage
func (d *DatabaseStorage) DownloadFile(ctx context.Context, fileID uint) ([]byte, error) {
	if !d.IsAvailable() {
		return nil, fmt.Errorf("database storage not available")
	}

	var file models.File
	if err := d.db.First(&file, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to retrieve file from database: %w", err)
	}

	if file.StorageType != "database" {
		return nil, fmt.Errorf("file is not stored in database")
	}

	return file.Content, nil
}

// DeleteFile deletes a file from database storage
func (d *DatabaseStorage) DeleteFile(ctx context.Context, fileID uint) error {
	if !d.IsAvailable() {
		return fmt.Errorf("database storage not available")
	}

	var file models.File
	if err := d.db.First(&file, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("file not found")
		}
		return fmt.Errorf("failed to find file: %w", err)
	}

	if file.StorageType != "database" {
		return fmt.Errorf("file is not stored in database")
	}

	if err := d.db.Delete(&file).Error; err != nil {
		return fmt.Errorf("failed to delete file from database: %w", err)
	}

	return nil
}

// GetFileURL returns a URL for accessing the file (for database storage, this would be an API endpoint)
func (d *DatabaseStorage) GetFileURL(ctx context.Context, fileID uint) (string, error) {
	// For database storage, return an API endpoint URL
	return fmt.Sprintf("/api/files/%d/download", fileID), nil
}
