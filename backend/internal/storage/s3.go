package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"react-golang-starter/internal/models"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// S3Storage implements FileStorage using AWS S3
type S3Storage struct {
	client     *s3.Client
	bucketName string
	region     string
	baseURL    string
}

// NewS3Storage creates a new S3 storage instance
func NewS3Storage() (*S3Storage, error) {
	// Check if AWS credentials are available
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")
	bucketName := os.Getenv("AWS_S3_BUCKET")

	if accessKey == "" || secretKey == "" || region == "" || bucketName == "" {
		return nil, fmt.Errorf("AWS S3 configuration not available")
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	baseURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucketName, region)

	return &S3Storage{
		client:     client,
		bucketName: bucketName,
		region:     region,
		baseURL:    baseURL,
	}, nil
}

// IsAvailable checks if S3 storage is available
func (s *S3Storage) IsAvailable() bool {
	return s.client != nil
}

// UploadFile uploads a file to S3
func (s *S3Storage) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*models.File, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("S3 storage not available")
	}

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Generate unique filename
	fileExt := filepath.Ext(header.Filename)
	fileName := strings.TrimSuffix(header.Filename, fileExt)
	uniqueID := uuid.New().String()
	s3Key := fmt.Sprintf("uploads/%s/%s_%s%s", time.Now().Format("2006-01-02"), fileName, uniqueID, fileExt)

	// Upload to S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(s3Key),
		Body:          bytes.NewReader(fileContent),
		ContentType:   aws.String(header.Header.Get("Content-Type")),
		ContentLength: aws.Int64(int64(len(fileContent))),
		ACL:           "private", // Set to private for security
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Create file record
	fileModel := &models.File{
		FileName:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		FileSize:    int64(len(fileContent)),
		Location:    s3Key,
		StorageType: "s3",
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	return fileModel, nil
}

// DownloadFile downloads a file from S3
func (s *S3Storage) DownloadFile(ctx context.Context, fileID uint) ([]byte, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("S3 storage not available")
	}

	// This method requires the file record to get the S3 key
	// In practice, this would need the file model passed in
	return nil, fmt.Errorf("DownloadFile not implemented for S3 - use GetFileURL instead")
}

// DeleteFile deletes a file from S3
func (s *S3Storage) DeleteFile(ctx context.Context, fileID uint) error {
	if !s.IsAvailable() {
		return fmt.Errorf("S3 storage not available")
	}

	// This method requires the file record to get the S3 key
	// In practice, this would need the file model passed in
	return fmt.Errorf("DeleteFile not implemented for S3 - use DeleteFile with S3 key")
}

// GetFileURL returns the S3 URL for the file
func (s *S3Storage) GetFileURL(ctx context.Context, fileID uint) (string, error) {
	if !s.IsAvailable() {
		return "", fmt.Errorf("S3 storage not available")
	}

	// This method requires the file record to get the S3 key
	// In practice, this would need the file model passed in
	return "", fmt.Errorf("GetFileURL not implemented for S3 - use with file record")
}

// GetFileURLWithKey returns the S3 URL for a specific key
func (s *S3Storage) GetFileURLWithKey(s3Key string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, s3Key)
}

// DeleteFileWithKey deletes a file from S3 using its key
func (s *S3Storage) DeleteFileWithKey(ctx context.Context, s3Key string) error {
	if !s.IsAvailable() {
		return fmt.Errorf("S3 storage not available")
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}
