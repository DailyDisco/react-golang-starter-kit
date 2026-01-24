package models

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// ============ System Settings Models ============

// SystemSetting represents an admin-configurable system setting
// swagger:model SystemSetting
type SystemSetting struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	Key         string          `json:"key" gorm:"type:varchar(100);uniqueIndex;not null"`
	Value       json.RawMessage `json:"value" gorm:"type:jsonb;not null"`
	Category    string          `json:"category" gorm:"type:varchar(50);not null;index"`
	Description string          `json:"description,omitempty" gorm:"type:text"`
	IsSensitive bool            `json:"is_sensitive" gorm:"default:false"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// SystemSettingResponse represents setting data returned to frontend
// swagger:model SystemSettingResponse
type SystemSettingResponse struct {
	ID          uint        `json:"id"`
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Category    string      `json:"category"`
	Description string      `json:"description,omitempty"`
	IsSensitive bool        `json:"is_sensitive"`
	UpdatedAt   string      `json:"updated_at"`
}

// ToResponse converts SystemSetting to SystemSettingResponse (hides sensitive values)
func (s *SystemSetting) ToResponse() SystemSettingResponse {
	var value interface{}
	if s.IsSensitive {
		value = "********" // Hide sensitive values
	} else {
		json.Unmarshal(s.Value, &value)
	}
	return SystemSettingResponse{
		ID:          s.ID,
		Key:         s.Key,
		Value:       value,
		Category:    s.Category,
		Description: s.Description,
		IsSensitive: s.IsSensitive,
		UpdatedAt:   s.UpdatedAt,
	}
}

// UpdateSystemSettingRequest represents a request to update a setting
// swagger:model UpdateSystemSettingRequest
type UpdateSystemSettingRequest struct {
	Value interface{} `json:"value" binding:"required"`
}

// EmailSettings represents SMTP configuration
// swagger:model EmailSettings
type EmailSettings struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password,omitempty"` // Only set on update, never returned
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
	Enabled      bool   `json:"enabled"`
}

// SecuritySettings represents security configuration
// swagger:model SecuritySettings
type SecuritySettings struct {
	PasswordMinLength        int  `json:"password_min_length"`
	PasswordRequireUppercase bool `json:"password_require_uppercase"`
	PasswordRequireLowercase bool `json:"password_require_lowercase"`
	PasswordRequireNumber    bool `json:"password_require_number"`
	PasswordRequireSpecial   bool `json:"password_require_special"`
	SessionTimeoutMinutes    int  `json:"session_timeout_minutes"`
	MaxLoginAttempts         int  `json:"max_login_attempts"`
	LockoutDurationMinutes   int  `json:"lockout_duration_minutes"`
	Require2FAForAdmins      bool `json:"require_2fa_for_admins"`
}

// SiteSettings represents site configuration
// swagger:model SiteSettings
type SiteSettings struct {
	SiteName           string `json:"site_name"`
	SiteLogoURL        string `json:"site_logo_url,omitempty"`
	MaintenanceMode    bool   `json:"maintenance_mode"`
	MaintenanceMessage string `json:"maintenance_message,omitempty"`
}

// ============ User Preferences Models ============

// UserPreferences represents user-specific settings
// swagger:model UserPreferences
type UserPreferences struct {
	ID                 uint            `json:"id" gorm:"primaryKey"`
	UserID             uint            `json:"user_id" gorm:"uniqueIndex;not null"`
	Theme              string          `json:"theme" gorm:"type:varchar(20);default:'system'"`
	Timezone           string          `json:"timezone" gorm:"type:varchar(50);default:'UTC'"`
	Language           string          `json:"language" gorm:"type:varchar(10);default:'en'"`
	DateFormat         string          `json:"date_format" gorm:"type:varchar(20);default:'MM/DD/YYYY'"`
	TimeFormat         string          `json:"time_format" gorm:"type:varchar(10);default:'12h'"`
	EmailNotifications json.RawMessage `json:"email_notifications" gorm:"type:jsonb"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// EmailNotificationSettings represents email notification preferences
type EmailNotificationSettings struct {
	Marketing    bool `json:"marketing"`
	Security     bool `json:"security"`
	Updates      bool `json:"updates"`
	WeeklyDigest bool `json:"weekly_digest"`
}

// UserPreferencesResponse represents preferences returned to frontend
// swagger:model UserPreferencesResponse
type UserPreferencesResponse struct {
	Theme              string                    `json:"theme"`
	Timezone           string                    `json:"timezone"`
	Language           string                    `json:"language"`
	DateFormat         string                    `json:"date_format"`
	TimeFormat         string                    `json:"time_format"`
	EmailNotifications EmailNotificationSettings `json:"email_notifications"`
}

// ToResponse converts UserPreferences to UserPreferencesResponse
func (p *UserPreferences) ToResponse() UserPreferencesResponse {
	var emailNotifs EmailNotificationSettings
	if p.EmailNotifications != nil {
		json.Unmarshal(p.EmailNotifications, &emailNotifs)
	}
	return UserPreferencesResponse{
		Theme:              p.Theme,
		Timezone:           p.Timezone,
		Language:           p.Language,
		DateFormat:         p.DateFormat,
		TimeFormat:         p.TimeFormat,
		EmailNotifications: emailNotifs,
	}
}

// UpdateUserPreferencesRequest represents a request to update preferences
// swagger:model UpdateUserPreferencesRequest
type UpdateUserPreferencesRequest struct {
	Theme              *string                    `json:"theme,omitempty"`
	Timezone           *string                    `json:"timezone,omitempty"`
	Language           *string                    `json:"language,omitempty"`
	DateFormat         *string                    `json:"date_format,omitempty"`
	TimeFormat         *string                    `json:"time_format,omitempty"`
	EmailNotifications *EmailNotificationSettings `json:"email_notifications,omitempty"`
}

// ============ User Sessions Models ============

// UserSession represents an active user session
// swagger:model UserSession
type UserSession struct {
	ID               uint            `json:"id" gorm:"primaryKey"`
	UserID           uint            `json:"user_id" gorm:"not null;index"`
	SessionTokenHash string          `json:"-" gorm:"type:varchar(64);uniqueIndex;not null"`
	DeviceInfo       json.RawMessage `json:"device_info" gorm:"type:jsonb"`
	IPAddress        string          `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent        string          `json:"-" gorm:"type:text"`
	Location         json.RawMessage `json:"location" gorm:"type:jsonb"`
	IsCurrent        bool            `json:"is_current" gorm:"default:false"`
	LastActiveAt     time.Time       `json:"last_active_at"`
	ExpiresAt        time.Time       `json:"expires_at"`
	CreatedAt        time.Time       `json:"created_at"`
}

// DeviceInfo represents parsed device information
type DeviceInfo struct {
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browser_version"`
	OS             string `json:"os"`
	OSVersion      string `json:"os_version"`
	DeviceType     string `json:"device_type"` // desktop, mobile, tablet
	DeviceName     string `json:"device_name,omitempty"`
}

// LocationInfo represents parsed location information
type LocationInfo struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
}

// UserSessionResponse represents session data returned to frontend
// swagger:model UserSessionResponse
type UserSessionResponse struct {
	ID           uint         `json:"id"`
	DeviceInfo   DeviceInfo   `json:"device_info"`
	IPAddress    string       `json:"ip_address"`
	Location     LocationInfo `json:"location"`
	IsCurrent    bool         `json:"is_current"`
	LastActiveAt string       `json:"last_active_at"`
	CreatedAt    string       `json:"created_at"`
}

// ToResponse converts UserSession to UserSessionResponse
func (s *UserSession) ToResponse() UserSessionResponse {
	var deviceInfo DeviceInfo
	var location LocationInfo
	if s.DeviceInfo != nil {
		json.Unmarshal(s.DeviceInfo, &deviceInfo)
	}
	if s.Location != nil {
		json.Unmarshal(s.Location, &location)
	}
	return UserSessionResponse{
		ID:           s.ID,
		DeviceInfo:   deviceInfo,
		IPAddress:    s.IPAddress,
		Location:     location,
		IsCurrent:    s.IsCurrent,
		LastActiveAt: s.LastActiveAt.Format(time.RFC3339),
		CreatedAt:    s.CreatedAt.Format(time.RFC3339),
	}
}

// ============ Two-Factor Auth Models ============

// UserTwoFactor represents 2FA configuration for a user
// swagger:model UserTwoFactor
type UserTwoFactor struct {
	ID                   uint            `json:"id" gorm:"primaryKey"`
	UserID               uint            `json:"user_id" gorm:"uniqueIndex;not null"`
	EncryptedSecret      string          `json:"-" gorm:"column:encrypted_secret;type:text;not null"`
	IsEnabled            bool            `json:"is_enabled" gorm:"default:false"`
	BackupCodesHash      json.RawMessage `json:"-" gorm:"column:backup_codes_hash;type:jsonb"`
	BackupCodesRemaining int             `json:"backup_codes_remaining" gorm:"default:10"`
	VerifiedAt           *string         `json:"verified_at,omitempty"`
	LastUsedAt           *string         `json:"last_used_at,omitempty"`
	FailedAttempts       int             `json:"-" gorm:"default:0"`
	LockedUntil          *string         `json:"-"`
	CreatedAt            string          `json:"created_at"`
	UpdatedAt            string          `json:"updated_at"`
}

// TwoFactorStatusResponse represents 2FA status returned to frontend
// swagger:model TwoFactorStatusResponse
type TwoFactorStatusResponse struct {
	Enabled              bool   `json:"enabled"`
	BackupCodesRemaining int    `json:"backup_codes_remaining"`
	VerifiedAt           string `json:"verified_at,omitempty"`
}

// TwoFactorSetupResponse represents 2FA setup data
// swagger:model TwoFactorSetupResponse
type TwoFactorSetupResponse struct {
	Secret     string `json:"secret"`
	QRCodeURL  string `json:"qr_code_url"`
	OTPAuthURL string `json:"otpauth_url"`
}

// TwoFactorVerifyRequest represents a request to verify 2FA setup
// swagger:model TwoFactorVerifyRequest
type TwoFactorVerifyRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// TwoFactorBackupCodesResponse represents backup codes returned to user
// swagger:model TwoFactorBackupCodesResponse
type TwoFactorBackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
	Message     string   `json:"message"`
}

// ============ IP Blocklist Models ============

// IPBlocklist represents a blocked IP entry
// swagger:model IPBlocklist
type IPBlocklist struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	IPAddress string         `json:"ip_address" gorm:"type:varchar(45);not null;index"`
	IPRange   string         `json:"ip_range,omitempty" gorm:"type:varchar(50)"`
	Reason    string         `json:"reason,omitempty" gorm:"type:varchar(500)"`
	BlockType string         `json:"block_type" gorm:"type:varchar(20);default:'manual'"`
	BlockedBy *uint          `json:"blocked_by,omitempty"`
	HitCount  int            `json:"hit_count" gorm:"default:0"`
	ExpiresAt *string        `json:"expires_at,omitempty"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// IPBlocklistResponse represents blocklist entry returned to frontend
// swagger:model IPBlocklistResponse
type IPBlocklistResponse struct {
	ID        uint   `json:"id"`
	IPAddress string `json:"ip_address"`
	IPRange   string `json:"ip_range,omitempty"`
	Reason    string `json:"reason,omitempty"`
	BlockType string `json:"block_type"`
	HitCount  int    `json:"hit_count"`
	ExpiresAt string `json:"expires_at,omitempty"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// CreateIPBlockRequest represents a request to block an IP
// swagger:model CreateIPBlockRequest
type CreateIPBlockRequest struct {
	IPAddress string `json:"ip_address" binding:"required"`
	IPRange   string `json:"ip_range,omitempty"`
	Reason    string `json:"reason,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"` // ISO 8601 format, null for permanent
}

// ============ Announcement Banner Models ============

// AnnouncementBanner represents a site-wide announcement
// swagger:model AnnouncementBanner
type AnnouncementBanner struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	Title         string          `json:"title" gorm:"type:varchar(255);not null"`
	Message       string          `json:"message" gorm:"type:text;not null"`
	Type          string          `json:"type" gorm:"type:varchar(20);default:'info'"`
	DisplayType   string          `json:"display_type" gorm:"type:varchar(20);default:'banner'"`
	Category      string          `json:"category" gorm:"type:varchar(20);default:'update'"`
	LinkURL       string          `json:"link_url,omitempty" gorm:"type:varchar(500)"`
	LinkText      string          `json:"link_text,omitempty" gorm:"type:varchar(100)"`
	IsActive      bool            `json:"is_active" gorm:"default:true"`
	IsDismissible bool            `json:"is_dismissible" gorm:"default:true"`
	ShowOnPages   json.RawMessage `json:"show_on_pages" gorm:"type:jsonb"`
	TargetRoles   pq.StringArray  `json:"target_roles,omitempty" gorm:"type:text[]"`
	Priority      int             `json:"priority" gorm:"default:0"`
	StartsAt      *string         `json:"starts_at,omitempty"`
	EndsAt        *string         `json:"ends_at,omitempty"`
	PublishedAt   *string         `json:"published_at,omitempty"`
	EmailSent     bool            `json:"email_sent" gorm:"default:false"`
	EmailSentAt   *string         `json:"email_sent_at,omitempty"`
	ViewCount     int             `json:"view_count" gorm:"default:0"`
	DismissCount  int             `json:"dismiss_count" gorm:"default:0"`
	CreatedBy     *uint           `json:"created_by,omitempty"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}

// AnnouncementBannerResponse represents announcement returned to frontend
// swagger:model AnnouncementBannerResponse
type AnnouncementBannerResponse struct {
	ID            uint     `json:"id"`
	Title         string   `json:"title"`
	Message       string   `json:"message"`
	Type          string   `json:"type"`
	DisplayType   string   `json:"display_type"`
	Category      string   `json:"category"`
	LinkURL       string   `json:"link_url,omitempty"`
	LinkText      string   `json:"link_text,omitempty"`
	IsDismissible bool     `json:"is_dismissible"`
	Priority      int      `json:"priority"`
	StartsAt      string   `json:"starts_at,omitempty"`
	EndsAt        string   `json:"ends_at,omitempty"`
	PublishedAt   string   `json:"published_at,omitempty"`
	IsActive      bool     `json:"is_active"`
	TargetRoles   []string `json:"target_roles,omitempty"`
}

// CreateAnnouncementRequest represents a request to create an announcement
// swagger:model CreateAnnouncementRequest
type CreateAnnouncementRequest struct {
	Title         string   `json:"title" binding:"required,max=255"`
	Message       string   `json:"message" binding:"required"`
	Type          string   `json:"type" binding:"omitempty,oneof=info warning error success maintenance"`
	DisplayType   string   `json:"display_type" binding:"omitempty,oneof=banner modal"`
	Category      string   `json:"category" binding:"omitempty,oneof=update feature bugfix"`
	LinkURL       string   `json:"link_url,omitempty"`
	LinkText      string   `json:"link_text,omitempty"`
	IsDismissible bool     `json:"is_dismissible"`
	ShowOnPages   []string `json:"show_on_pages,omitempty"`
	TargetRoles   []string `json:"target_roles,omitempty"`
	Priority      int      `json:"priority"`
	StartsAt      string   `json:"starts_at,omitempty"`
	EndsAt        string   `json:"ends_at,omitempty"`
	SendEmail     bool     `json:"send_email"`
	IsActive      bool     `json:"is_active"`
}

// UpdateAnnouncementRequest represents a request to update an announcement
// swagger:model UpdateAnnouncementRequest
type UpdateAnnouncementRequest struct {
	Title         *string   `json:"title,omitempty"`
	Message       *string   `json:"message,omitempty"`
	Type          *string   `json:"type,omitempty"`
	DisplayType   *string   `json:"display_type,omitempty"`
	Category      *string   `json:"category,omitempty"`
	LinkURL       *string   `json:"link_url,omitempty"`
	LinkText      *string   `json:"link_text,omitempty"`
	IsActive      *bool     `json:"is_active,omitempty"`
	IsDismissible *bool     `json:"is_dismissible,omitempty"`
	ShowOnPages   *[]string `json:"show_on_pages,omitempty"`
	TargetRoles   *[]string `json:"target_roles,omitempty"`
	Priority      *int      `json:"priority,omitempty"`
	StartsAt      *string   `json:"starts_at,omitempty"`
	EndsAt        *string   `json:"ends_at,omitempty"`
}

// ToResponse converts AnnouncementBanner to AnnouncementBannerResponse
func (a *AnnouncementBanner) ToResponse() AnnouncementBannerResponse {
	startsAt := ""
	if a.StartsAt != nil {
		startsAt = *a.StartsAt
	}
	endsAt := ""
	if a.EndsAt != nil {
		endsAt = *a.EndsAt
	}
	publishedAt := ""
	if a.PublishedAt != nil {
		publishedAt = *a.PublishedAt
	}
	return AnnouncementBannerResponse{
		ID:            a.ID,
		Title:         a.Title,
		Message:       a.Message,
		Type:          a.Type,
		DisplayType:   a.DisplayType,
		Category:      a.Category,
		LinkURL:       a.LinkURL,
		LinkText:      a.LinkText,
		IsDismissible: a.IsDismissible,
		Priority:      a.Priority,
		StartsAt:      startsAt,
		EndsAt:        endsAt,
		PublishedAt:   publishedAt,
		IsActive:      a.IsActive,
		TargetRoles:   a.TargetRoles,
	}
}

// UserAnnouncementRead tracks which modal announcements a user has seen
// swagger:model UserAnnouncementRead
type UserAnnouncementRead struct {
	UserID         uint   `gorm:"primaryKey"`
	AnnouncementID uint   `gorm:"primaryKey"`
	ReadAt         string `json:"read_at"`
}

// ChangelogResponse represents paginated changelog data
// swagger:model ChangelogResponse
type ChangelogResponse struct {
	Data []AnnouncementBannerResponse `json:"data"`
	Meta ChangelogMeta                `json:"meta"`
}

// ChangelogMeta represents pagination metadata for changelog
type ChangelogMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// UserDismissedAnnouncement tracks which announcements a user has dismissed
type UserDismissedAnnouncement struct {
	UserID         uint   `gorm:"primaryKey"`
	AnnouncementID uint   `gorm:"primaryKey"`
	DismissedAt    string `json:"dismissed_at"`
}

// ============ Login History Models ============

// LoginHistory represents a login attempt record
// swagger:model LoginHistory
type LoginHistory struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	UserID        uint            `json:"user_id" gorm:"not null;index"`
	Success       bool            `json:"success" gorm:"not null"`
	FailureReason string          `json:"failure_reason,omitempty" gorm:"type:varchar(100)"`
	IPAddress     string          `json:"ip_address" gorm:"type:varchar(45);not null"`
	UserAgent     string          `json:"-" gorm:"type:text"`
	DeviceInfo    json.RawMessage `json:"device_info" gorm:"type:jsonb"`
	Location      json.RawMessage `json:"location" gorm:"type:jsonb"`
	AuthMethod    string          `json:"auth_method" gorm:"type:varchar(20);default:'password'"`
	SessionID     *uint           `json:"session_id,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// TableName specifies the table name for GORM (matches migration)
func (LoginHistory) TableName() string {
	return "login_history"
}

// Login failure reason constants
const (
	LoginFailureInvalidPassword = "invalid_password"
	LoginFailureAccountLocked   = "account_locked"
	LoginFailure2FAFailed       = "2fa_failed"
	LoginFailureAccountInactive = "account_inactive"
	LoginFailureEmailNotFound   = "email_not_found"
)

// Auth method constants
const (
	AuthMethodPassword     = "password"
	AuthMethodOAuthGoogle  = "oauth_google"
	AuthMethodOAuthGitHub  = "oauth_github"
	AuthMethodRefreshToken = "refresh_token"
	AuthMethod2FA          = "2fa"
)

// LoginHistoryResponse represents login history returned to frontend
// swagger:model LoginHistoryResponse
type LoginHistoryResponse struct {
	ID            uint         `json:"id"`
	Success       bool         `json:"success"`
	FailureReason string       `json:"failure_reason,omitempty"`
	IPAddress     string       `json:"ip_address"`
	DeviceInfo    DeviceInfo   `json:"device_info"`
	Location      LocationInfo `json:"location"`
	AuthMethod    string       `json:"auth_method"`
	CreatedAt     string       `json:"created_at"`
}

// ToResponse converts LoginHistory to LoginHistoryResponse
func (l *LoginHistory) ToResponse() LoginHistoryResponse {
	var deviceInfo DeviceInfo
	var location LocationInfo
	if l.DeviceInfo != nil {
		json.Unmarshal(l.DeviceInfo, &deviceInfo)
	}
	if l.Location != nil {
		json.Unmarshal(l.Location, &location)
	}
	return LoginHistoryResponse{
		ID:            l.ID,
		Success:       l.Success,
		FailureReason: l.FailureReason,
		IPAddress:     l.IPAddress,
		DeviceInfo:    deviceInfo,
		Location:      location,
		AuthMethod:    l.AuthMethod,
		CreatedAt:     l.CreatedAt.Format(time.RFC3339),
	}
}

// ============ Email Template Models ============

// EmailTemplate represents a customizable email template
// swagger:model EmailTemplate
type EmailTemplate struct {
	ID                 uint            `json:"id" gorm:"primaryKey"`
	Key                string          `json:"key" gorm:"type:varchar(50);uniqueIndex;not null"`
	Name               string          `json:"name" gorm:"type:varchar(100);not null"`
	Description        string          `json:"description,omitempty" gorm:"type:text"`
	Subject            string          `json:"subject" gorm:"type:varchar(255);not null"`
	BodyHTML           string          `json:"body_html" gorm:"type:text;not null"`
	BodyText           string          `json:"body_text,omitempty" gorm:"type:text"`
	AvailableVariables json.RawMessage `json:"available_variables" gorm:"type:jsonb"`
	IsActive           bool            `json:"is_active" gorm:"default:true"`
	IsSystem           bool            `json:"is_system" gorm:"default:false"`
	LastSentAt         *string         `json:"last_sent_at,omitempty"`
	SendCount          int             `json:"send_count" gorm:"default:0"`
	CreatedBy          *uint           `json:"created_by,omitempty"`
	UpdatedBy          *uint           `json:"updated_by,omitempty"`
	CreatedAt          string          `json:"created_at"`
	UpdatedAt          string          `json:"updated_at"`
}

// TemplateVariable represents a variable available in an email template
type TemplateVariable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// EmailTemplateResponse represents template data returned to frontend
// swagger:model EmailTemplateResponse
type EmailTemplateResponse struct {
	ID                 uint               `json:"id"`
	Key                string             `json:"key"`
	Name               string             `json:"name"`
	Description        string             `json:"description,omitempty"`
	Subject            string             `json:"subject"`
	BodyHTML           string             `json:"body_html"`
	BodyText           string             `json:"body_text,omitempty"`
	AvailableVariables []TemplateVariable `json:"available_variables"`
	IsActive           bool               `json:"is_active"`
	IsSystem           bool               `json:"is_system"`
	SendCount          int                `json:"send_count"`
	UpdatedAt          string             `json:"updated_at"`
}

// ToResponse converts EmailTemplate to EmailTemplateResponse
func (t *EmailTemplate) ToResponse() EmailTemplateResponse {
	var variables []TemplateVariable
	if t.AvailableVariables != nil {
		json.Unmarshal(t.AvailableVariables, &variables)
	}
	return EmailTemplateResponse{
		ID:                 t.ID,
		Key:                t.Key,
		Name:               t.Name,
		Description:        t.Description,
		Subject:            t.Subject,
		BodyHTML:           t.BodyHTML,
		BodyText:           t.BodyText,
		AvailableVariables: variables,
		IsActive:           t.IsActive,
		IsSystem:           t.IsSystem,
		SendCount:          t.SendCount,
		UpdatedAt:          t.UpdatedAt,
	}
}

// UpdateEmailTemplateRequest represents a request to update an email template
// swagger:model UpdateEmailTemplateRequest
type UpdateEmailTemplateRequest struct {
	Subject  *string `json:"subject,omitempty"`
	BodyHTML *string `json:"body_html,omitempty"`
	BodyText *string `json:"body_text,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// PreviewEmailTemplateRequest represents a request to preview an email template
// swagger:model PreviewEmailTemplateRequest
type PreviewEmailTemplateRequest struct {
	Variables map[string]string `json:"variables"`
}

// TestEmailRequest represents a request to send a test email
// swagger:model TestEmailRequest
type TestEmailRequest struct {
	// RecipientEmail is the email address to send the test to (optional, defaults to admin's email)
	RecipientEmail string `json:"recipient_email,omitempty"`
}

// PreviewEmailTemplateResponse represents a rendered email preview
// swagger:model PreviewEmailTemplateResponse
type PreviewEmailTemplateResponse struct {
	Subject  string `json:"subject"`
	BodyHTML string `json:"body_html"`
	BodyText string `json:"body_text,omitempty"`
}

// ============ System Health Models ============

// SystemHealthResponse represents overall system health status
// swagger:model SystemHealthResponse
type SystemHealthResponse struct {
	Status     string               `json:"status"` // healthy, degraded, unhealthy
	Timestamp  string               `json:"timestamp"`
	Components []HealthComponent    `json:"components"`
	Metrics    *SystemHealthMetrics `json:"metrics,omitempty"`
}

// HealthComponent represents the health of a single component
type HealthComponent struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"` // healthy, degraded, unhealthy
	Message   string                 `json:"message,omitempty"`
	Latency   string                 `json:"latency,omitempty"`
	LastCheck string                 `json:"last_check"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// SystemHealthMetrics represents system metrics
type SystemHealthMetrics struct {
	Database *DatabaseMetrics `json:"database,omitempty"`
	Cache    *CacheMetrics    `json:"cache,omitempty"`
	Storage  *StorageMetrics  `json:"storage,omitempty"`
	API      *APIMetrics      `json:"api,omitempty"`
}

// DatabaseMetrics represents database health metrics
type DatabaseMetrics struct {
	Status            string `json:"status"`
	ConnectionsActive int    `json:"connections_active"`
	ConnectionsIdle   int    `json:"connections_idle"`
	ConnectionsMax    int    `json:"connections_max"`
	AvgQueryTime      string `json:"avg_query_time"`
	SlowQueries       int    `json:"slow_queries"`
	Uptime            string `json:"uptime"`
}

// CacheMetrics represents cache health metrics
type CacheMetrics struct {
	Status      string  `json:"status"`
	MemoryUsed  string  `json:"memory_used"`
	MemoryMax   string  `json:"memory_max"`
	HitRate     float64 `json:"hit_rate"`
	Keys        int64   `json:"keys"`
	Connections int     `json:"connections"`
}

// StorageMetrics represents file storage metrics
type StorageMetrics struct {
	Status    string  `json:"status"`
	Used      string  `json:"used"`
	Available string  `json:"available"`
	Total     string  `json:"total"`
	UsedPct   float64 `json:"used_percent"`
	FileCount int64   `json:"file_count"`
}

// APIMetrics represents API performance metrics
type APIMetrics struct {
	RequestsPerMinute int64   `json:"requests_per_minute"`
	AvgResponseTime   string  `json:"avg_response_time"`
	P50ResponseTime   string  `json:"p50_response_time"`
	P95ResponseTime   string  `json:"p95_response_time"`
	P99ResponseTime   string  `json:"p99_response_time"`
	ErrorRate         float64 `json:"error_rate"`
}

// ============ Password Change Models ============

// ChangePasswordRequest represents a request to change password
// swagger:model ChangePasswordRequest
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// ============ Account Deletion Models ============

// RequestAccountDeletionRequest represents a request to delete account
// swagger:model RequestAccountDeletionRequest
type RequestAccountDeletionRequest struct {
	Password string `json:"password" binding:"required"`
	Reason   string `json:"reason,omitempty"`
}

// AccountDeletionStatusResponse represents deletion status
// swagger:model AccountDeletionStatusResponse
type AccountDeletionStatusResponse struct {
	Requested    bool   `json:"requested"`
	RequestedAt  string `json:"requested_at,omitempty"`
	ScheduledFor string `json:"scheduled_for,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

// Data export status constants
const (
	ExportStatusPending    = "pending"
	ExportStatusProcessing = "processing"
	ExportStatusCompleted  = "completed"
	ExportStatusFailed     = "failed"
	ExportStatusExpired    = "expired"
)

// DataExport represents a user data export request
// swagger:model DataExport
type DataExport struct {
	ID           uint    `json:"id" gorm:"primaryKey"`
	UserID       uint    `json:"user_id" gorm:"not null;index"`
	Status       string  `json:"status" gorm:"type:varchar(50);default:'pending'"`
	DownloadURL  *string `json:"download_url,omitempty" gorm:"type:varchar(500)"`
	FilePath     *string `json:"-" gorm:"type:varchar(500)"`
	StorageType  string  `json:"-" gorm:"type:varchar(20);default:'local'"` // "local" or "s3"
	FileSize     int64   `json:"file_size,omitempty"`
	RequestedAt  string  `json:"requested_at" gorm:"not null"`
	CompletedAt  *string `json:"completed_at,omitempty"`
	ExpiresAt    *string `json:"expires_at,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// DataExportResponse represents the response for data export status
// swagger:model DataExportResponse
type DataExportResponse struct {
	ID          uint   `json:"id"`
	Status      string `json:"status"`
	DownloadURL string `json:"download_url,omitempty"`
	RequestedAt string `json:"requested_at"`
	ExpiresAt   string `json:"expires_at,omitempty"`
	FileSize    int64  `json:"file_size,omitempty"`
}

// ToResponse converts a DataExport to DataExportResponse
func (d *DataExport) ToResponse() DataExportResponse {
	resp := DataExportResponse{
		ID:          d.ID,
		Status:      d.Status,
		RequestedAt: d.RequestedAt,
		FileSize:    d.FileSize,
	}
	if d.DownloadURL != nil {
		resp.DownloadURL = *d.DownloadURL
	}
	if d.ExpiresAt != nil {
		resp.ExpiresAt = *d.ExpiresAt
	}
	return resp
}

// ConnectedAccountResponse represents an OAuth connected account
// swagger:model ConnectedAccountResponse
type ConnectedAccountResponse struct {
	Provider       string `json:"provider"`
	ProviderUserID string `json:"provider_user_id,omitempty"`
	Email          string `json:"email,omitempty"`
	ConnectedAt    string `json:"connected_at"`
}

// AvatarUploadResponse represents the response after avatar upload
// swagger:model AvatarUploadResponse
type AvatarUploadResponse struct {
	AvatarURL string `json:"avatar_url"`
}

// ============ User Activity Models ============

// ActivityLogItem represents a single activity entry for the user's activity feed
// swagger:model ActivityLogItem
type ActivityLogItem struct {
	ID         uint                   `json:"id"`
	TargetType string                 `json:"target_type"`
	Action     string                 `json:"action"`
	Changes    map[string]interface{} `json:"changes,omitempty"`
	CreatedAt  string                 `json:"created_at"`
}

// MyActivityResponse represents the response for user's activity feed
// swagger:model MyActivityResponse
type MyActivityResponse struct {
	Activities []ActivityLogItem `json:"activities"`
	Count      int               `json:"count"`
	Total      int               `json:"total"`
}
