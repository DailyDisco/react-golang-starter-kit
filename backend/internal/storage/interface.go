package storage

import (
	"context"
	"mime/multipart"
	"react-golang-starter/internal/models"
)

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	// UploadFile uploads a file to the storage backend
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*models.File, error)

	// DownloadFile retrieves a file from the storage backend
	DownloadFile(ctx context.Context, fileID uint) ([]byte, error)

	// DeleteFile removes a file from the storage backend
	DeleteFile(ctx context.Context, fileID uint) error

	// GetFileURL returns the URL for accessing the file
	GetFileURL(ctx context.Context, fileID uint) (string, error)

	// IsAvailable checks if the storage backend is available
	IsAvailable() bool
}
