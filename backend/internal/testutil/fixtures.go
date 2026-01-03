// Package testutil provides testing utilities, fixtures, and helpers.
package testutil

import (
	"fmt"
	"sync/atomic"
	"time"

	"react-golang-starter/internal/models"
)

// Global sequence counter for unique IDs
var sequence uint64

// nextID returns the next unique ID for test fixtures
func nextID() uint {
	return uint(atomic.AddUint64(&sequence, 1))
}

// UserFactory creates test User fixtures with sensible defaults.
type UserFactory struct {
	id            uint
	name          string
	email         string
	password      string
	role          string
	isActive      bool
	emailVerified bool
	avatarURL     string
}

// NewUserFactory creates a new UserFactory with defaults.
func NewUserFactory() *UserFactory {
	id := nextID()
	return &UserFactory{
		id:            id,
		name:          fmt.Sprintf("Test User %d", id),
		email:         fmt.Sprintf("user%d@example.com", id),
		password:      "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // "Password123!"
		role:          "user",
		isActive:      true,
		emailVerified: true,
	}
}

// WithID sets a specific ID.
func (f *UserFactory) WithID(id uint) *UserFactory {
	f.id = id
	return f
}

// WithName sets the user's name.
func (f *UserFactory) WithName(name string) *UserFactory {
	f.name = name
	return f
}

// WithEmail sets the user's email.
func (f *UserFactory) WithEmail(email string) *UserFactory {
	f.email = email
	return f
}

// WithPassword sets the user's hashed password.
func (f *UserFactory) WithPassword(hashedPassword string) *UserFactory {
	f.password = hashedPassword
	return f
}

// WithRole sets the user's role.
func (f *UserFactory) WithRole(role string) *UserFactory {
	f.role = role
	return f
}

// AsAdmin sets the user as an admin.
func (f *UserFactory) AsAdmin() *UserFactory {
	f.role = "admin"
	return f
}

// AsSuperAdmin sets the user as a super admin.
func (f *UserFactory) AsSuperAdmin() *UserFactory {
	f.role = "super_admin"
	return f
}

// Inactive sets the user as inactive.
func (f *UserFactory) Inactive() *UserFactory {
	f.isActive = false
	return f
}

// UnverifiedEmail sets the user's email as unverified.
func (f *UserFactory) UnverifiedEmail() *UserFactory {
	f.emailVerified = false
	return f
}

// WithAvatarURL sets the user's avatar URL.
func (f *UserFactory) WithAvatarURL(url string) *UserFactory {
	f.avatarURL = url
	return f
}

// Build creates a User model from the factory configuration.
func (f *UserFactory) Build() *models.User {
	now := time.Now()
	return &models.User{
		ID:            f.id,
		Name:          f.name,
		Email:         f.email,
		Password:      f.password,
		Role:          f.role,
		IsActive:      f.isActive,
		EmailVerified: f.emailVerified,
		AvatarURL:     f.avatarURL,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// BuildUserResponse creates a UserResponse from the factory configuration.
func (f *UserFactory) BuildUserResponse() models.UserResponse {
	return models.UserResponse{
		ID:            f.id,
		Name:          f.name,
		Email:         f.email,
		Role:          f.role,
		IsActive:      f.isActive,
		EmailVerified: f.emailVerified,
		AvatarURL:     f.avatarURL,
	}
}

// FileFactory creates test File fixtures with sensible defaults.
type FileFactory struct {
	id          uint
	userID      uint
	fileName    string
	contentType string
	fileSize    int64
	storageType string
	location    string
}

// NewFileFactory creates a new FileFactory with defaults.
func NewFileFactory() *FileFactory {
	id := nextID()
	return &FileFactory{
		id:          id,
		userID:      1,
		fileName:    fmt.Sprintf("test-file-%d.txt", id),
		contentType: "text/plain",
		fileSize:    1024,
		storageType: "database",
		location:    fmt.Sprintf("files/%d/test-file-%d.txt", 1, id),
	}
}

// WithID sets a specific ID.
func (f *FileFactory) WithID(id uint) *FileFactory {
	f.id = id
	return f
}

// WithUserID sets the owner's user ID.
func (f *FileFactory) WithUserID(userID uint) *FileFactory {
	f.userID = userID
	return f
}

// WithFileName sets the filename.
func (f *FileFactory) WithFileName(fileName string) *FileFactory {
	f.fileName = fileName
	return f
}

// WithContentType sets the content type.
func (f *FileFactory) WithContentType(contentType string) *FileFactory {
	f.contentType = contentType
	return f
}

// WithFileSize sets the file size.
func (f *FileFactory) WithFileSize(size int64) *FileFactory {
	f.fileSize = size
	return f
}

// AsImage sets the file as an image.
func (f *FileFactory) AsImage() *FileFactory {
	f.contentType = "image/png"
	f.fileName = fmt.Sprintf("test-image-%d.png", f.id)
	return f
}

// AsPDF sets the file as a PDF.
func (f *FileFactory) AsPDF() *FileFactory {
	f.contentType = "application/pdf"
	f.fileName = fmt.Sprintf("test-document-%d.pdf", f.id)
	return f
}

// Build creates a File model from the factory configuration.
func (f *FileFactory) Build() *models.File {
	now := time.Now().Format(time.RFC3339)
	return &models.File{
		ID:          f.id,
		UserID:      f.userID,
		FileName:    f.fileName,
		ContentType: f.contentType,
		FileSize:    f.fileSize,
		StorageType: f.storageType,
		Location:    f.location,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AuditLogFactory creates test AuditLog fixtures.
type AuditLogFactory struct {
	id         uint
	userID     *uint
	action     string
	targetType string
	targetID   *uint
	changes    string
	ipAddr     string
	userAgent  string
}

// NewAuditLogFactory creates a new AuditLogFactory with defaults.
func NewAuditLogFactory() *AuditLogFactory {
	id := nextID()
	userID := uint(1)
	targetID := uint(1)
	return &AuditLogFactory{
		id:         id,
		userID:     &userID,
		action:     "test_action",
		targetType: "user",
		targetID:   &targetID,
		changes:    `{"field": "value"}`,
		ipAddr:     "127.0.0.1",
		userAgent:  "TestAgent/1.0",
	}
}

// WithUserID sets the user ID.
func (f *AuditLogFactory) WithUserID(userID uint) *AuditLogFactory {
	f.userID = &userID
	return f
}

// WithoutUser creates an audit log without a user.
func (f *AuditLogFactory) WithoutUser() *AuditLogFactory {
	f.userID = nil
	return f
}

// WithAction sets the action.
func (f *AuditLogFactory) WithAction(action string) *AuditLogFactory {
	f.action = action
	return f
}

// WithTargetType sets the target type.
func (f *AuditLogFactory) WithTargetType(targetType string) *AuditLogFactory {
	f.targetType = targetType
	return f
}

// WithTargetID sets the target ID.
func (f *AuditLogFactory) WithTargetID(targetID uint) *AuditLogFactory {
	f.targetID = &targetID
	return f
}

// Build creates an AuditLog model from the factory configuration.
func (f *AuditLogFactory) Build() *models.AuditLog {
	return &models.AuditLog{
		ID:         f.id,
		UserID:     f.userID,
		Action:     f.action,
		TargetType: f.targetType,
		TargetID:   f.targetID,
		Changes:    f.changes,
		IPAddress:  f.ipAddr,
		UserAgent:  f.userAgent,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}
}

// ResetSequence resets the global sequence counter (useful between tests).
func ResetSequence() {
	atomic.StoreUint64(&sequence, 0)
}
