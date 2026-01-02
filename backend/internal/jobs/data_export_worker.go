package jobs

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/email"
	"react-golang-starter/internal/models"

	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
)

// DataExportArgs contains the arguments for a data export job
type DataExportArgs struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	ExportID uint   `json:"export_id"`
}

// Kind returns the job type identifier
func (DataExportArgs) Kind() string {
	return "generate_data_export"
}

// InsertOpts returns the default insert options for this job type
func (DataExportArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       river.QueueDefault,
		MaxAttempts: 3,
	}
}

// DataExportWorker processes data export jobs
type DataExportWorker struct {
	river.WorkerDefaults[DataExportArgs]
}

// Work processes a data export job
func (w *DataExportWorker) Work(ctx context.Context, job *river.Job[DataExportArgs]) error {
	args := job.Args

	log.Info().
		Uint("user_id", args.UserID).
		Uint("export_id", args.ExportID).
		Msg("starting data export generation")

	// Update status to processing
	if err := database.DB.Model(&models.DataExport{}).
		Where("id = ?", args.ExportID).
		Update("status", models.ExportStatusProcessing).Error; err != nil {
		return fmt.Errorf("failed to update export status: %w", err)
	}

	// Compile user data
	userData, err := compileUserData(ctx, args.UserID)
	if err != nil {
		updateExportError(args.ExportID, fmt.Sprintf("Data compilation failed: %v", err))
		return fmt.Errorf("failed to compile user data: %w", err)
	}

	// Create ZIP archive in memory
	zipBuf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuf)

	// Add JSON file with user data
	jsonData, err := json.MarshalIndent(userData, "", "  ")
	if err != nil {
		updateExportError(args.ExportID, "JSON serialization failed")
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	jsonFile, err := zipWriter.Create("user_data.json")
	if err != nil {
		zipWriter.Close()
		updateExportError(args.ExportID, "ZIP creation failed")
		return fmt.Errorf("failed to create ZIP entry: %w", err)
	}

	if _, err := jsonFile.Write(jsonData); err != nil {
		zipWriter.Close()
		updateExportError(args.ExportID, "ZIP write failed")
		return fmt.Errorf("failed to write to ZIP: %w", err)
	}

	// Add a README file
	readmeContent := fmt.Sprintf(`Data Export for User ID: %d
Generated: %s

This archive contains your personal data as stored in our system.

Files included:
- user_data.json: Your complete data in JSON format

For questions about this export, please contact support.
`, args.UserID, time.Now().Format(time.RFC3339))

	readmeFile, err := zipWriter.Create("README.txt")
	if err == nil {
		readmeFile.Write([]byte(readmeContent))
	}

	zipWriter.Close()

	// Store file
	exportsDir := getExportsDir()
	if err := os.MkdirAll(exportsDir, 0755); err != nil {
		updateExportError(args.ExportID, "File storage failed")
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	filename := fmt.Sprintf("user_data_%d_%d.zip", args.UserID, time.Now().Unix())
	filePath := filepath.Join(exportsDir, filename)

	if err := os.WriteFile(filePath, zipBuf.Bytes(), 0600); err != nil {
		updateExportError(args.ExportID, "File write failed")
		return fmt.Errorf("failed to write export file: %w", err)
	}

	// Update export record
	expiresAt := time.Now().AddDate(0, 0, 7) // 7 days
	if err := database.DB.Model(&models.DataExport{}).
		Where("id = ?", args.ExportID).
		Updates(map[string]interface{}{
			"status":       models.ExportStatusCompleted,
			"download_url": fmt.Sprintf("/api/users/me/export/download"),
			"file_path":    filePath,
			"file_size":    int64(zipBuf.Len()),
			"completed_at": time.Now().Format(time.RFC3339),
			"expires_at":   expiresAt.Format(time.RFC3339),
		}).Error; err != nil {
		return fmt.Errorf("failed to update export record: %w", err)
	}

	// Send notification email
	if email.IsAvailable() {
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}

		err = email.Send(ctx, email.SendParams{
			To:           args.Email,
			Subject:      "Your Data Export is Ready",
			TemplateName: "data_export_ready",
			Data: map[string]interface{}{
				"DownloadLink": fmt.Sprintf("%s/settings/privacy", frontendURL),
				"ExpiresIn":    "7 days",
				"FileSize":     formatBytes(int64(zipBuf.Len())),
			},
		})
		if err != nil {
			log.Warn().Err(err).Msg("failed to send export ready email")
			// Don't fail the job - email is optional
		}
	}

	log.Info().
		Uint("user_id", args.UserID).
		Uint("export_id", args.ExportID).
		Int("size_bytes", zipBuf.Len()).
		Msg("data export completed successfully")

	return nil
}

// updateExportError updates the export record with an error status
func updateExportError(exportID uint, errorMsg string) {
	database.DB.Model(&models.DataExport{}).
		Where("id = ?", exportID).
		Updates(map[string]interface{}{
			"status":        models.ExportStatusFailed,
			"error_message": errorMsg,
		})
}

// getExportsDir returns the directory for storing exports
func getExportsDir() string {
	dir := os.Getenv("DATA_EXPORTS_DIR")
	if dir == "" {
		dir = "exports"
	}
	return dir
}

// formatBytes formats bytes as human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// derefString safely dereferences a string pointer, returning empty string if nil
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// EnqueueDataExport queues a data export job
func EnqueueDataExport(ctx context.Context, userID uint, email string, exportID uint) error {
	if !IsAvailable() {
		return fmt.Errorf("job system not available")
	}

	return Insert(ctx, DataExportArgs{
		UserID:   userID,
		Email:    email,
		ExportID: exportID,
	}, nil)
}

// UserDataExport contains all user data for GDPR export
type UserDataExport struct {
	ExportedAt     string                  `json:"exported_at"`
	User           userExportData          `json:"user"`
	Preferences    *models.UserPreferences `json:"preferences,omitempty"`
	Sessions       []models.UserSession    `json:"sessions,omitempty"`
	LoginHistory   []models.LoginHistory   `json:"login_history,omitempty"`
	TwoFactor      *twoFactorExportData    `json:"two_factor,omitempty"`
	APIKeys        []apiKeyExportData      `json:"api_keys,omitempty"`
	OAuthProviders []oauthProviderExport   `json:"oauth_providers,omitempty"`
	Files          []fileExportData        `json:"files,omitempty"`
	Organizations  []orgMembershipExport   `json:"organizations,omitempty"`
	AuditLogs      []auditLogExport        `json:"audit_logs,omitempty"`
}

type userExportData struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	IsActive      bool   `json:"is_active"`
	Role          string `json:"role"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	Bio           string `json:"bio,omitempty"`
	Location      string `json:"location,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type twoFactorExportData struct {
	Enabled              bool   `json:"enabled"`
	VerifiedAt           string `json:"verified_at,omitempty"`
	LastUsedAt           string `json:"last_used_at,omitempty"`
	BackupCodesRemaining int    `json:"backup_codes_remaining"`
}

type apiKeyExportData struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	KeyPreview string `json:"key_preview"`
	IsActive   bool   `json:"is_active"`
	UsageCount int    `json:"usage_count"`
	LastUsedAt string `json:"last_used_at,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type oauthProviderExport struct {
	Provider string `json:"provider"`
	Email    string `json:"email"`
	LinkedAt string `json:"linked_at"`
}

type fileExportData struct {
	ID          uint   `json:"id"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
	CreatedAt   string `json:"created_at"`
}

type orgMembershipExport struct {
	OrganizationName string `json:"organization_name"`
	OrganizationSlug string `json:"organization_slug"`
	Role             string `json:"role"`
	JoinedAt         string `json:"joined_at"`
}

type auditLogExport struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	IPAddress   string `json:"ip_address,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// compileUserData gathers all user data for GDPR export
func compileUserData(ctx context.Context, userID uint) (*UserDataExport, error) {
	data := &UserDataExport{
		ExportedAt: time.Now().Format(time.RFC3339),
	}

	// Get user (exclude sensitive fields like password hash)
	var user models.User
	if err := database.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}
	data.User = userExportData{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		IsActive:      user.IsActive,
		Role:          user.Role,
		AvatarURL:     user.AvatarURL,
		Bio:           user.Bio,
		Location:      user.Location,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	// Get preferences
	var prefs models.UserPreferences
	if err := database.DB.WithContext(ctx).Where("user_id = ?", userID).First(&prefs).Error; err == nil {
		data.Preferences = &prefs
	}

	// Get sessions
	database.DB.WithContext(ctx).Where("user_id = ?", userID).Find(&data.Sessions)

	// Get login history (last 100)
	database.DB.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(100).
		Find(&data.LoginHistory)

	// Get 2FA info (no secrets)
	var twoFactor models.UserTwoFactor
	if err := database.DB.WithContext(ctx).Where("user_id = ?", userID).First(&twoFactor).Error; err == nil {
		data.TwoFactor = &twoFactorExportData{
			Enabled:              twoFactor.IsEnabled,
			VerifiedAt:           derefString(twoFactor.VerifiedAt),
			LastUsedAt:           derefString(twoFactor.LastUsedAt),
			BackupCodesRemaining: twoFactor.BackupCodesRemaining,
		}
	}

	// Get API keys (no actual keys)
	var apiKeys []models.UserAPIKey
	database.DB.WithContext(ctx).Where("user_id = ?", userID).Find(&apiKeys)
	for _, key := range apiKeys {
		data.APIKeys = append(data.APIKeys, apiKeyExportData{
			ID:         key.ID,
			Name:       key.Name,
			KeyPreview: key.KeyPreview,
			IsActive:   key.IsActive,
			UsageCount: key.UsageCount,
			LastUsedAt: derefString(key.LastUsedAt),
			CreatedAt:  key.CreatedAt,
		})
	}

	// Get OAuth providers
	var providers []models.OAuthProvider
	database.DB.WithContext(ctx).Where("user_id = ?", userID).Find(&providers)
	for _, p := range providers {
		data.OAuthProviders = append(data.OAuthProviders, oauthProviderExport{
			Provider: p.Provider,
			Email:    p.Email,
			LinkedAt: p.CreatedAt,
		})
	}

	// Get files
	var files []models.File
	database.DB.WithContext(ctx).Where("user_id = ?", userID).Find(&files)
	for _, f := range files {
		data.Files = append(data.Files, fileExportData{
			ID:          f.ID,
			FileName:    f.FileName,
			FileSize:    f.FileSize,
			ContentType: f.ContentType,
			CreatedAt:   f.CreatedAt,
		})
	}

	// Get organization memberships
	var members []models.OrganizationMember
	database.DB.WithContext(ctx).Where("user_id = ?", userID).Preload("Organization").Find(&members)
	for _, m := range members {
		if m.Organization != nil {
			data.Organizations = append(data.Organizations, orgMembershipExport{
				OrganizationName: m.Organization.Name,
				OrganizationSlug: m.Organization.Slug,
				Role:             string(m.Role),
				JoinedAt:         m.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	// Get audit logs (last 500)
	var auditLogs []models.AuditLog
	database.DB.WithContext(ctx).Where("user_id = ? OR target_id = ?", userID, userID).
		Order("created_at DESC").
		Limit(500).
		Find(&auditLogs)
	for _, l := range auditLogs {
		data.AuditLogs = append(data.AuditLogs, auditLogExport{
			Action:      l.Action,
			Description: fmt.Sprintf("%s on %s", l.Action, l.TargetType),
			IPAddress:   l.IPAddress,
			CreatedAt:   l.CreatedAt,
		})
	}

	return data, nil
}
