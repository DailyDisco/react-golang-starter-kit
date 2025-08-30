package models

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

	// The name of the user
	// example: John Doe
	Name string `json:"name" binding:"required"`

	// The email address of the user (must be unique)
	// example: john.doe@example.com
	Email string `json:"email" gorm:"unique" binding:"required,email"`

	// Hashed password for authentication
	Password string `json:"-" gorm:"not null" binding:"required"`

	// Whether the user's email has been verified
	EmailVerified bool `json:"email_verified" gorm:"default:false"`

	// Email verification token
	VerificationToken string `json:"-" gorm:"unique"`

	// Token expiration time
	VerificationExpires string `json:"-"`

	// Whether the user account is active
	IsActive bool `json:"is_active" gorm:"default:true"`
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

// HealthResponse represents the health check response
// swagger:model HealthResponse
type HealthResponse struct {
	// Health status (ok, error, degraded)
	// example: ok
	Status string `json:"status" example:"ok"`

	// Health status message
	// example: Server is running
	Message string `json:"message" example:"Server is running"`
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
