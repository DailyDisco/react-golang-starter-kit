package models

import "gorm.io/gorm"

// User represents a user in the system
// swagger:model User
type User struct {
	// The unique ID of the user
	// example: 1
	ID uint `json:"id" gorm:"primaryKey"`

	// When the user was created
	// example: 2023-08-27T12:00:00Z
	CreatedAt string `json:"created_at"`

	// When the user was last updated
	// example: 2023-08-27T12:00:00Z
	UpdatedAt string `json:"updated_at"`

	// When the user was soft deleted (null if not deleted)
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// The name of the user
	// example: John Doe
	Name string `json:"name" binding:"required"`

	// The email address of the user (must be unique)
	// example: john.doe@example.com
	Email string `json:"email" gorm:"uniqueIndex;not null" binding:"required,email"`

	// Hashed password for authentication
	Password string `json:"-" gorm:"not null" binding:"required"`

	// Whether the user's email has been verified
	EmailVerified bool `json:"email_verified" gorm:"default:false;index"`

	// Email verification token (indexed for fast lookups during verification/password reset)
	VerificationToken string `json:"-" gorm:"uniqueIndex"`

	// Token expiration time
	VerificationExpires string `json:"-"`

	// Refresh token for obtaining new access tokens
	RefreshToken string `json:"-" gorm:"index"`

	// Refresh token expiration time
	RefreshTokenExpires string `json:"-"`

	// Whether the user account is active
	IsActive bool `json:"is_active" gorm:"default:true;index"`

	// The role of the user (e.g., "super_admin", "admin", "premium", "user")
	// example: user
	Role string `json:"role" gorm:"type:varchar(50);default:'user';index"`

	// Stripe customer ID for billing
	StripeCustomerID string `json:"-" gorm:"uniqueIndex"`

	// OAuth provider (if user signed up via OAuth)
	// example: google
	OAuthProvider string `json:"oauth_provider,omitempty" gorm:"type:varchar(50);index"`

	// OAuth provider user ID
	OAuthProviderID string `json:"-" gorm:"type:varchar(255)"`

	// User's avatar URL (from OAuth provider or uploaded)
	AvatarURL string `json:"avatar_url,omitempty" gorm:"type:varchar(500)"`
}

// TokenBlacklist stores revoked JWT tokens to prevent reuse
// swagger:model TokenBlacklist
type TokenBlacklist struct {
	// The unique ID of the blacklist entry
	ID uint `json:"id" gorm:"primaryKey"`

	// SHA-256 hash of the token (for security, we don't store the actual token)
	TokenHash string `json:"-" gorm:"uniqueIndex;not null;size:64"`

	// User ID who owned the token
	UserID uint `json:"user_id" gorm:"not null;index"`

	// When the token expires (for cleanup)
	ExpiresAt string `json:"expires_at" gorm:"not null;index"`

	// When the token was revoked
	RevokedAt string `json:"revoked_at" gorm:"not null"`

	// Reason for revocation (logout, password_change, admin_revoke, etc.)
	Reason string `json:"reason" gorm:"type:varchar(50);default:'logout'"`
}

// UserResponse represents the user data returned to the frontend (without sensitive fields)
// swagger:model UserResponse
type UserResponse struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	IsActive      bool   `json:"is_active"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	Role          string `json:"role"`
	OAuthProvider string `json:"oauth_provider,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
}

// ToUserResponse converts a User to UserResponse (removes sensitive fields)
func (u *User) ToUserResponse() UserResponse {
	return UserResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		IsActive:      u.IsActive,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		Role:          u.Role,
		OAuthProvider: u.OAuthProvider,
		AvatarURL:     u.AvatarURL,
	}
}

// LoginRequest represents the login request payload
// swagger:model LoginRequest
type LoginRequest struct {
	// User's email address (must be valid email format)
	// required: true
	// example: john.doe@example.com
	Email string `json:"email" binding:"required,email" example:"john.doe@example.com"`

	// User's password (minimum 8 characters)
	// required: true
	// example: SecurePass123!
	// minLength: 8
	Password string `json:"password" binding:"required" example:"SecurePass123!"`
}

// RegisterRequest represents the registration request payload
// swagger:model RegisterRequest
type RegisterRequest struct {
	// User's full name (required, non-empty)
	// required: true
	// example: John Doe
	// minLength: 1
	Name string `json:"name" binding:"required" example:"John Doe"`

	// User's email address (must be valid and unique)
	// required: true
	// example: john.doe@example.com
	Email string `json:"email" binding:"required,email" example:"john.doe@example.com"`

	// User's password (minimum 8 characters, must contain letters and numbers)
	// required: true
	// example: SecurePass123!
	// minLength: 8
	Password string `json:"password" binding:"required,min=8" example:"SecurePass123!"`
}

// AuthResponse represents the authentication response with tokens
// swagger:model AuthResponse
type AuthResponse struct {
	// Authenticated user information
	User UserResponse `json:"user"`

	// JWT access token for subsequent API requests
	// example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// Refresh token for obtaining new access tokens (longer-lived)
	// example: abc123def456...
	RefreshToken string `json:"refresh_token,omitempty" example:"abc123def456..."`

	// Access token expiration time in seconds
	// example: 900
	ExpiresIn int64 `json:"expires_in,omitempty" example:"900"`
}

// RefreshTokenRequest represents a request to refresh the access token
// swagger:model RefreshTokenRequest
type RefreshTokenRequest struct {
	// The refresh token obtained during login
	// required: true
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// PasswordResetRequest represents a password reset request
// swagger:model PasswordResetRequest
type PasswordResetRequest struct {
	// Email address for password reset (must be registered)
	// required: true
	// example: john.doe@example.com
	Email string `json:"email" binding:"required,email" example:"john.doe@example.com"`
}

// PasswordResetConfirm represents password reset confirmation
// swagger:model PasswordResetConfirm
type PasswordResetConfirm struct {
	// Reset token received via email
	// required: true
	// example: abc123def456
	Token string `json:"token" binding:"required" example:"abc123def456"`

	// New password (minimum 8 characters)
	// required: true
	// example: NewSecurePass123!
	// minLength: 8
	Password string `json:"password" binding:"required,min=8" example:"NewSecurePass123!"`
}

// ErrorResponse represents an error response
// swagger:model ErrorResponse
type ErrorResponse struct {
	// Error type/category
	// example: Bad Request
	Error string `json:"error" example:"Bad Request"`

	// Detailed error message
	// example: Invalid email format
	Message string `json:"message,omitempty" example:"Invalid email format"`

	// HTTP status code
	// example: 400
	Code int `json:"code,omitempty" example:"400"`

	// Request ID for tracing and debugging
	// example: 550e8400-e29b-41d4-a716-446655440000
	RequestID string `json:"request_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// SuccessResponse represents a success response
// swagger:model SuccessResponse
type SuccessResponse struct {
	// Operation success status
	// example: true
	Success bool `json:"success" example:"true"`

	// Success message
	// example: User created successfully
	Message string `json:"message" example:"User created successfully"`

	// Response data (varies by endpoint)
	Data interface{} `json:"data,omitempty"`
}

// ComponentStatus represents the status of a single component (e.g., database, redis)
type ComponentStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// RuntimeInfo contains runtime statistics
type RuntimeInfo struct {
	Goroutines  int    `json:"goroutines"`
	MemoryAlloc string `json:"memory_alloc"`
	MemorySys   string `json:"memory_sys"`
	NumGC       uint32 `json:"num_gc"`
	GoVersion   string `json:"go_version"`
	NumCPU      int    `json:"num_cpu"`
	GOOS        string `json:"goos"`
	GOARCH      string `json:"goarch"`
}

// VersionInfo contains application version information
type VersionInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time,omitempty"`
	GitCommit string `json:"git_commit,omitempty"`
}

// HealthStatus represents the overall health of the application
// swagger:model HealthStatus
type HealthStatus struct {
	OverallStatus string            `json:"overall_status"`
	Timestamp     string            `json:"timestamp"`
	Uptime        string            `json:"uptime"`
	Version       VersionInfo       `json:"version"`
	Runtime       *RuntimeInfo      `json:"runtime,omitempty"`
	Components    []ComponentStatus `json:"components"`
}

// UsersResponse represents a paginated list of users response
// swagger:model UsersResponse
type UsersResponse struct {
	// List of users on this page
	Users []UserResponse `json:"users"`

	// Number of users returned in this response
	// example: 10
	Count int `json:"count" example:"10"`

	// Total number of users in the system
	// example: 150
	Total int `json:"total" example:"150"`

	// Current page number
	// example: 1
	Page int `json:"page" example:"1"`

	// Items per page
	// example: 10
	Limit int `json:"limit" example:"10"`

	// Total number of pages available
	// example: 15
	TotalPages int `json:"total_pages" example:"15"`
}

// PaginationQuery represents pagination query parameters
// swagger:model PaginationQuery
type PaginationQuery struct {
	// Page number to retrieve (default: 1, minimum: 1)
	// example: 1
	Page int `json:"page" form:"page" example:"1"`

	// Number of items per page (default: 10, maximum: 100)
	// example: 10
	Limit int `json:"limit" form:"limit" example:"10"`
}

// File represents a file stored in the system
// swagger:model File
type File struct {
	// The unique ID of the file
	// example: 1
	ID uint `json:"id" gorm:"primaryKey"`

	// The ID of the user who owns this file
	// example: 1
	UserID uint `json:"user_id" gorm:"index;not null"`

	// When the file was created
	// example: 2023-08-27T12:00:00Z
	CreatedAt string `json:"created_at"`

	// When the file was last updated
	// example: 2023-08-27T12:00:00Z
	UpdatedAt string `json:"updated_at"`

	// When the file was soft deleted (null if not deleted)
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// The original name of the uploaded file
	// example: my-document.pdf
	FileName string `json:"file_name" gorm:"not null" binding:"required"`

	// MIME type of the file
	// example: application/pdf
	ContentType string `json:"content_type"`

	// Size of the file in bytes
	// example: 1024
	FileSize int64 `json:"file_size"`

	// Storage location - either S3 URL or database identifier
	// example: https://bucket-name.s3.amazonaws.com/uploads/file.pdf
	Location string `json:"location"`

	// Actual file content (only used for database storage)
	Content []byte `json:"-" gorm:"type:bytea"`

	// Storage type (s3 or database)
	// example: s3
	StorageType string `json:"storage_type" gorm:"default:database"`
}

// FileResponse represents the file data returned to the frontend
// swagger:model FileResponse
type FileResponse struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	FileSize    int64  `json:"file_size"`
	Location    string `json:"location"`
	StorageType string `json:"storage_type"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ToFileResponse converts a File to FileResponse
func (f *File) ToFileResponse() FileResponse {
	return FileResponse{
		ID:          f.ID,
		UserID:      f.UserID,
		FileName:    f.FileName,
		ContentType: f.ContentType,
		FileSize:    f.FileSize,
		Location:    f.Location,
		StorageType: f.StorageType,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}

// Role constants
const (
	RoleSuperAdmin = "super_admin" // System administrators (full access)
	RoleAdmin      = "admin"       // Content/service administrators
	RolePremium    = "premium"     // Paid subscribers with extra features
	RoleUser       = "user"        // Regular users (default)
)

// RoleHierarchy defines the permission level for each role (higher number = more permissions)
var RoleHierarchy = map[string]int{
	RoleSuperAdmin: 100,
	RoleAdmin:      50,
	RolePremium:    20,
	RoleUser:       10,
}

// Subscription status constants
const (
	SubscriptionStatusActive   = "active"
	SubscriptionStatusPastDue  = "past_due"
	SubscriptionStatusCanceled = "canceled"
	SubscriptionStatusTrialing = "trialing"
	SubscriptionStatusUnpaid   = "unpaid"
)

// Subscription represents a user's subscription record
// swagger:model Subscription
type Subscription struct {
	// The unique ID of the subscription
	ID uint `json:"id" gorm:"primaryKey"`

	// When the subscription was created
	CreatedAt string `json:"created_at"`

	// When the subscription was last updated
	UpdatedAt string `json:"updated_at"`

	// When the subscription was soft deleted (null if not deleted)
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// User ID (foreign key)
	UserID uint `json:"user_id" gorm:"uniqueIndex;not null"`

	// Stripe subscription ID
	StripeSubscriptionID string `json:"stripe_subscription_id" gorm:"uniqueIndex;not null"`

	// Stripe price ID for the subscribed plan
	StripePriceID string `json:"stripe_price_id" gorm:"not null"`

	// Subscription status (active, past_due, canceled, trialing, unpaid)
	Status string `json:"status" gorm:"type:varchar(50);default:'active';index"`

	// When the current billing period started
	CurrentPeriodStart string `json:"current_period_start"`

	// When the current billing period ends
	CurrentPeriodEnd string `json:"current_period_end"`

	// Whether the subscription will cancel at period end
	CancelAtPeriodEnd bool `json:"cancel_at_period_end" gorm:"default:false"`

	// When the subscription was canceled (if applicable)
	CanceledAt string `json:"canceled_at,omitempty"`
}

// SubscriptionResponse represents subscription data returned to the frontend
// swagger:model SubscriptionResponse
type SubscriptionResponse struct {
	ID                 uint   `json:"id"`
	UserID             uint   `json:"user_id"`
	Status             string `json:"status"`
	StripePriceID      string `json:"stripe_price_id"`
	CurrentPeriodStart string `json:"current_period_start"`
	CurrentPeriodEnd   string `json:"current_period_end"`
	CancelAtPeriodEnd  bool   `json:"cancel_at_period_end"`
	CanceledAt         string `json:"canceled_at,omitempty"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

// ToSubscriptionResponse converts a Subscription to SubscriptionResponse
func (s *Subscription) ToSubscriptionResponse() SubscriptionResponse {
	return SubscriptionResponse{
		ID:                 s.ID,
		UserID:             s.UserID,
		Status:             s.Status,
		StripePriceID:      s.StripePriceID,
		CurrentPeriodStart: s.CurrentPeriodStart,
		CurrentPeriodEnd:   s.CurrentPeriodEnd,
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		CanceledAt:         s.CanceledAt,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
	}
}

// IsActiveSubscription returns true if the subscription is in an active state
func (s *Subscription) IsActiveSubscription() bool {
	return s.Status == SubscriptionStatusActive || s.Status == SubscriptionStatusTrialing
}

// BillingPlan represents an available subscription plan
// swagger:model BillingPlan
type BillingPlan struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PriceID     string   `json:"price_id"`
	Amount      int64    `json:"amount"`   // Price in cents
	Currency    string   `json:"currency"` // e.g., "usd"
	Interval    string   `json:"interval"` // e.g., "month", "year"
	Features    []string `json:"features"`
}

// CreateCheckoutRequest represents a request to create a checkout session
// swagger:model CreateCheckoutRequest
type CreateCheckoutRequest struct {
	PriceID string `json:"price_id" binding:"required"`
}

// CheckoutSessionResponse represents the response from creating a checkout session
// swagger:model CheckoutSessionResponse
type CheckoutSessionResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

// PortalSessionResponse represents the response from creating a billing portal session
// swagger:model PortalSessionResponse
type PortalSessionResponse struct {
	URL string `json:"url"`
}

// BillingConfigResponse represents the public billing configuration
// swagger:model BillingConfigResponse
type BillingConfigResponse struct {
	PublishableKey string `json:"publishable_key"`
}

// ============ OAuth Models ============

// OAuth provider constants
const (
	OAuthProviderGoogle = "google"
	OAuthProviderGitHub = "github"
)

// OAuthProvider represents a linked OAuth provider for a user
// swagger:model OAuthProvider
type OAuthProvider struct {
	// The unique ID
	ID uint `json:"id" gorm:"primaryKey"`

	// User ID (foreign key)
	UserID uint `json:"user_id" gorm:"not null;index"`

	// OAuth provider name (google, github)
	Provider string `json:"provider" gorm:"type:varchar(50);not null;index"`

	// User ID from the OAuth provider
	ProviderUserID string `json:"provider_user_id" gorm:"type:varchar(255);not null"`

	// Email from OAuth provider (may differ from user's primary email)
	Email string `json:"email" gorm:"type:varchar(255)"`

	// OAuth access token (encrypted in production)
	AccessToken string `json:"-" gorm:"type:text"`

	// OAuth refresh token (encrypted in production)
	RefreshToken string `json:"-" gorm:"type:text"`

	// When the OAuth token expires
	TokenExpiresAt string `json:"token_expires_at,omitempty"`

	// Raw profile data from provider (as JSON)
	RawData string `json:"-" gorm:"type:jsonb"`

	// When the link was created
	CreatedAt string `json:"created_at"`

	// When the link was last updated
	UpdatedAt string `json:"updated_at"`
}

// OAuthUserInfo represents user info from an OAuth provider
type OAuthUserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Provider  string `json:"provider"`
}

// OAuthCallbackRequest represents the OAuth callback data
// swagger:model OAuthCallbackRequest
type OAuthCallbackRequest struct {
	// Authorization code from OAuth provider
	Code string `json:"code" binding:"required"`

	// State parameter for CSRF protection
	State string `json:"state"`

	// Optional: Link to existing account (requires auth)
	LinkToAccount bool `json:"link_to_account,omitempty"`
}

// OAuthURLResponse represents the OAuth authorization URL response
// swagger:model OAuthURLResponse
type OAuthURLResponse struct {
	// URL to redirect user to for OAuth authorization
	URL string `json:"url"`

	// State parameter for CSRF verification
	State string `json:"state"`
}

// LinkedProvidersResponse represents the user's linked OAuth providers
// swagger:model LinkedProvidersResponse
type LinkedProvidersResponse struct {
	// List of linked providers
	Providers []LinkedProvider `json:"providers"`
}

// LinkedProvider represents a single linked OAuth provider
type LinkedProvider struct {
	Provider string `json:"provider"`
	Email    string `json:"email"`
	LinkedAt string `json:"linked_at"`
}

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// ============ Audit Log Models ============

// AuditAction constants
const (
	AuditActionCreate          = "create"
	AuditActionUpdate          = "update"
	AuditActionDelete          = "delete"
	AuditActionLogin           = "login"
	AuditActionLogout          = "logout"
	AuditActionImpersonate     = "impersonate"
	AuditActionStopImpersonate = "stop_impersonate"
	AuditActionPasswordReset   = "password_reset"
	AuditActionRoleChange      = "role_change"
)

// AuditTargetType constants
const (
	AuditTargetUser         = "user"
	AuditTargetSubscription = "subscription"
	AuditTargetFile         = "file"
	AuditTargetSettings     = "settings"
	AuditTargetFeatureFlag  = "feature_flag"
)

// AuditLog represents an audit log entry
// swagger:model AuditLog
type AuditLog struct {
	// The unique ID of the audit log entry
	ID uint `json:"id" gorm:"primaryKey"`

	// User ID of the actor (null for system actions)
	UserID *uint `json:"user_id,omitempty" gorm:"index"`

	// User who performed the action (populated via join)
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// Type of resource affected
	TargetType string `json:"target_type" gorm:"type:varchar(50);not null;index"`

	// ID of the affected resource
	TargetID *uint `json:"target_id,omitempty"`

	// Action performed
	Action string `json:"action" gorm:"type:varchar(50);not null;index"`

	// Before/after diff for updates (JSON)
	Changes string `json:"changes,omitempty" gorm:"type:jsonb"`

	// IP address of the request
	IPAddress string `json:"ip_address,omitempty" gorm:"type:varchar(45)"`

	// User agent string
	UserAgent string `json:"user_agent,omitempty" gorm:"type:text"`

	// Additional metadata (JSON)
	Metadata string `json:"metadata,omitempty" gorm:"type:jsonb"`

	// When the action occurred
	CreatedAt string `json:"created_at" gorm:"index"`
}

// AuditLogResponse represents audit log data returned to the frontend
// swagger:model AuditLogResponse
type AuditLogResponse struct {
	ID         uint        `json:"id"`
	UserID     *uint       `json:"user_id,omitempty"`
	UserName   string      `json:"user_name,omitempty"`
	UserEmail  string      `json:"user_email,omitempty"`
	TargetType string      `json:"target_type"`
	TargetID   *uint       `json:"target_id,omitempty"`
	Action     string      `json:"action"`
	Changes    interface{} `json:"changes,omitempty"`
	IPAddress  string      `json:"ip_address,omitempty"`
	UserAgent  string      `json:"user_agent,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
	CreatedAt  string      `json:"created_at"`
}

// ToAuditLogResponse converts an AuditLog to AuditLogResponse
func (a *AuditLog) ToAuditLogResponse() AuditLogResponse {
	resp := AuditLogResponse{
		ID:         a.ID,
		UserID:     a.UserID,
		TargetType: a.TargetType,
		TargetID:   a.TargetID,
		Action:     a.Action,
		IPAddress:  a.IPAddress,
		UserAgent:  a.UserAgent,
		CreatedAt:  a.CreatedAt,
	}
	if a.User != nil {
		resp.UserName = a.User.Name
		resp.UserEmail = a.User.Email
	}
	return resp
}

// AuditLogsResponse represents a paginated list of audit logs
// swagger:model AuditLogsResponse
type AuditLogsResponse struct {
	Logs       []AuditLogResponse `json:"logs"`
	Count      int                `json:"count"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}

// AuditLogFilter represents filter options for audit log queries
type AuditLogFilter struct {
	UserID     *uint  `form:"user_id"`
	TargetType string `form:"target_type"`
	TargetID   *uint  `form:"target_id"`
	Action     string `form:"action"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
}

// ============ Feature Flag Models ============

// FeatureFlag represents a feature flag
// swagger:model FeatureFlag
type FeatureFlag struct {
	// The unique ID
	ID uint `json:"id" gorm:"primaryKey"`

	// Unique key for the feature flag
	Key string `json:"key" gorm:"type:varchar(100);uniqueIndex;not null"`

	// Human-readable name
	Name string `json:"name" gorm:"type:varchar(255);not null"`

	// Description of the feature
	Description string `json:"description,omitempty" gorm:"type:text"`

	// Whether the flag is globally enabled
	Enabled bool `json:"enabled" gorm:"default:false"`

	// Percentage of users to roll out to (0-100)
	RolloutPercentage int `json:"rollout_percentage" gorm:"default:0"`

	// Roles that have access regardless of rollout
	AllowedRoles string `json:"allowed_roles,omitempty" gorm:"type:text[]"`

	// Additional configuration metadata (JSON)
	Metadata string `json:"metadata,omitempty" gorm:"type:jsonb"`

	// When the flag was created
	CreatedAt string `json:"created_at"`

	// When the flag was last updated
	UpdatedAt string `json:"updated_at"`
}

// FeatureFlagResponse represents feature flag data returned to the frontend
// swagger:model FeatureFlagResponse
type FeatureFlagResponse struct {
	ID                uint     `json:"id"`
	Key               string   `json:"key"`
	Name              string   `json:"name"`
	Description       string   `json:"description,omitempty"`
	Enabled           bool     `json:"enabled"`
	RolloutPercentage int      `json:"rollout_percentage"`
	AllowedRoles      []string `json:"allowed_roles,omitempty"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

// UserFeatureFlag represents a user-specific feature flag override
// swagger:model UserFeatureFlag
type UserFeatureFlag struct {
	// The unique ID
	ID uint `json:"id" gorm:"primaryKey"`

	// User ID
	UserID uint `json:"user_id" gorm:"not null;index"`

	// Feature flag ID
	FeatureFlagID uint `json:"feature_flag_id" gorm:"not null;index"`

	// Override value for this user
	Enabled bool `json:"enabled" gorm:"not null"`

	// When the override was created
	CreatedAt string `json:"created_at"`

	// When the override was last updated
	UpdatedAt string `json:"updated_at"`
}

// CreateFeatureFlagRequest represents a request to create a feature flag
// swagger:model CreateFeatureFlagRequest
type CreateFeatureFlagRequest struct {
	Key               string   `json:"key" binding:"required"`
	Name              string   `json:"name" binding:"required"`
	Description       string   `json:"description,omitempty"`
	Enabled           bool     `json:"enabled"`
	RolloutPercentage int      `json:"rollout_percentage"`
	AllowedRoles      []string `json:"allowed_roles,omitempty"`
}

// UpdateFeatureFlagRequest represents a request to update a feature flag
// swagger:model UpdateFeatureFlagRequest
type UpdateFeatureFlagRequest struct {
	Name              *string   `json:"name,omitempty"`
	Description       *string   `json:"description,omitempty"`
	Enabled           *bool     `json:"enabled,omitempty"`
	RolloutPercentage *int      `json:"rollout_percentage,omitempty"`
	AllowedRoles      *[]string `json:"allowed_roles,omitempty"`
}

// FeatureFlagsResponse represents a list of feature flags
// swagger:model FeatureFlagsResponse
type FeatureFlagsResponse struct {
	Flags []FeatureFlagResponse `json:"flags"`
	Count int                   `json:"count"`
}

// ============ Admin Models ============

// ImpersonateRequest represents a request to impersonate a user
// swagger:model ImpersonateRequest
type ImpersonateRequest struct {
	// User ID to impersonate
	UserID uint `json:"user_id" binding:"required"`
	// Reason for impersonation (for audit logging)
	Reason string `json:"reason,omitempty"`
}

// ImpersonateResponse represents the response from starting impersonation
// swagger:model ImpersonateResponse
type ImpersonateResponse struct {
	// The user being impersonated
	User UserResponse `json:"user"`
	// New JWT token with impersonation claims
	Token string `json:"token"`
	// Original admin user ID (for stopping impersonation)
	OriginalUserID uint `json:"original_user_id"`
}

// AdminStatsResponse represents admin dashboard statistics
// swagger:model AdminStatsResponse
type AdminStatsResponse struct {
	TotalUsers        int64 `json:"total_users"`
	ActiveUsers       int64 `json:"active_users"`
	VerifiedUsers     int64 `json:"verified_users"`
	NewUsersToday     int64 `json:"new_users_today"`
	NewUsersThisWeek  int64 `json:"new_users_this_week"`
	NewUsersThisMonth int64 `json:"new_users_this_month"`

	TotalSubscriptions    int64 `json:"total_subscriptions"`
	ActiveSubscriptions   int64 `json:"active_subscriptions"`
	CanceledSubscriptions int64 `json:"canceled_subscriptions"`

	TotalFiles    int64 `json:"total_files"`
	TotalFileSize int64 `json:"total_file_size"` // in bytes

	UsersByRole map[string]int64 `json:"users_by_role"`
}
