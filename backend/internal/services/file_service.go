package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/storage"
	"time"

	"gorm.io/gorm"
)

// ErrAccessDenied is returned when a user attempts to access a file they don't own
var ErrAccessDenied = errors.New("access denied")

// FileService handles file operations using the appropriate storage backend
type FileService struct {
	db            *gorm.DB
	s3Storage     *storage.S3Storage
	dbStorage     *storage.DatabaseStorage
	activeStorage storage.FileStorage
}

// NewFileService creates a new file service instance
func NewFileService() (*FileService, error) {
	// Initialize S3 storage (will return nil if not configured)
	s3Storage, _ := storage.NewS3Storage()

	// Initialize database storage (always available if DB is connected)
	dbStorage := storage.NewDatabaseStorage()

	var activeStorage storage.FileStorage

	// Prefer S3 if available, otherwise use database
	if s3Storage != nil && s3Storage.IsAvailable() {
		activeStorage = s3Storage
	} else {
		activeStorage = dbStorage
	}

	return &FileService{
		db:            database.DB,
		s3Storage:     s3Storage,
		dbStorage:     dbStorage,
		activeStorage: activeStorage,
	}, nil
}

// UploadFile uploads a file using the active storage backend
func (fs *FileService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*models.File, error) {
	// Upload using active storage
	fileModel, err := fs.activeStorage.UploadFile(ctx, file, header)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Set UserID from context if available
	if userID, ok := auth.GetUserIDFromContext(ctx); ok {
		fileModel.UserID = userID
	}

	// For S3 storage, save the metadata to database
	if fs.activeStorage == fs.s3Storage {
		fileModel.CreatedAt = time.Now().Format(time.RFC3339)
		fileModel.UpdatedAt = time.Now().Format(time.RFC3339)

		if err := fs.db.Create(fileModel).Error; err != nil {
			// If database save fails, try to clean up S3 file
			if cleanupErr := fs.s3Storage.DeleteFileWithKey(ctx, fileModel.Location); cleanupErr != nil {
				fmt.Printf("Warning: failed to cleanup S3 file after database error: %v\n", cleanupErr)
			}
			return nil, fmt.Errorf("failed to save file metadata to database: %w", err)
		}
	}

	return fileModel, nil
}

// DownloadFile downloads a file using the appropriate storage backend
func (fs *FileService) DownloadFile(ctx context.Context, fileID uint) ([]byte, *models.File, error) {
	// Get file metadata from database
	var file models.File
	if err := fs.db.First(&file, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, fmt.Errorf("file not found")
		}
		return nil, nil, fmt.Errorf("failed to retrieve file metadata: %w", err)
	}

	var content []byte
	var err error

	switch file.StorageType {
	case "s3":
		if fs.s3Storage == nil || !fs.s3Storage.IsAvailable() {
			return nil, nil, fmt.Errorf("S3 storage not available for file retrieval")
		}
		// For S3, return the URL instead of downloading content
		return nil, &file, fmt.Errorf("use GetFileURL for S3 files")
	case "database":
		content, err = fs.dbStorage.DownloadFile(ctx, fileID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to download file from database: %w", err)
		}
	default:
		return nil, nil, fmt.Errorf("unknown storage type: %s", file.StorageType)
	}

	return content, &file, nil
}

// DeleteFile deletes a file from both storage and database
func (fs *FileService) DeleteFile(ctx context.Context, fileID uint) error {
	// Get file metadata from database
	var file models.File
	if err := fs.db.First(&file, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("file not found")
		}
		return fmt.Errorf("failed to retrieve file metadata: %w", err)
	}

	// Delete from storage backend
	switch file.StorageType {
	case "s3":
		if fs.s3Storage != nil && fs.s3Storage.IsAvailable() {
			if err := fs.s3Storage.DeleteFileWithKey(ctx, file.Location); err != nil {
				return fmt.Errorf("failed to delete file from S3: %w", err)
			}
		}
	case "database":
		if err := fs.dbStorage.DeleteFile(ctx, fileID); err != nil {
			return fmt.Errorf("failed to delete file from database: %w", err)
		}
	default:
		return fmt.Errorf("unknown storage type: %s", file.StorageType)
	}

	return nil
}

// GetFileURL returns the URL for accessing a file
func (fs *FileService) GetFileURL(ctx context.Context, fileID uint) (string, error) {
	// Get file metadata from database
	var file models.File
	if err := fs.db.First(&file, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("file not found")
		}
		return "", fmt.Errorf("failed to retrieve file metadata: %w", err)
	}

	switch file.StorageType {
	case "s3":
		if fs.s3Storage != nil && fs.s3Storage.IsAvailable() {
			return fs.s3Storage.GetFileURLWithKey(file.Location), nil
		}
		return "", fmt.Errorf("S3 storage not available")
	case "database":
		return fs.dbStorage.GetFileURL(ctx, fileID)
	default:
		return "", fmt.Errorf("unknown storage type: %s", file.StorageType)
	}
}

// GetFileByID retrieves file metadata by ID
func (fs *FileService) GetFileByID(fileID uint) (*models.File, error) {
	var file models.File
	if err := fs.db.First(&file, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to retrieve file: %w", err)
	}
	return &file, nil
}

// GetFileByIDForUser retrieves file metadata by ID with ownership check
// Returns error if file doesn't exist or user doesn't own it (unless admin)
func (fs *FileService) GetFileByIDForUser(fileID uint, userID uint, isAdmin bool) (*models.File, error) {
	file, err := fs.GetFileByID(fileID)
	if err != nil {
		return nil, err
	}

	// Admins can access any file
	if isAdmin {
		return file, nil
	}

	// Check ownership - allow access if file has no owner (legacy files) or user owns it
	if file.UserID != 0 && file.UserID != userID {
		return nil, fmt.Errorf("%w: you do not own this file", ErrAccessDenied)
	}

	return file, nil
}

// ListFiles retrieves a list of files with pagination
func (fs *FileService) ListFiles(limit, offset int) ([]models.File, error) {
	var files []models.File
	if err := fs.db.Limit(limit).Offset(offset).Find(&files).Error; err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	return files, nil
}

// ListFilesForUser retrieves a list of files for a specific user with pagination
// Admins can see all files by passing isAdmin=true
func (fs *FileService) ListFilesForUser(userID uint, isAdmin bool, limit, offset int) ([]models.File, error) {
	var files []models.File
	query := fs.db.Limit(limit).Offset(offset)

	// Admins can see all files
	if !isAdmin {
		query = query.Where("user_id = ? OR user_id IS NULL OR user_id = 0", userID)
	}

	if err := query.Find(&files).Error; err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	return files, nil
}

// GetStorageType returns the currently active storage type
func (fs *FileService) GetStorageType() string {
	if fs.activeStorage == fs.s3Storage {
		return "s3"
	}
	return "database"
}
