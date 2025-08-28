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
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents the registration request payload
// swagger:model RegisterRequest
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// AuthResponse represents the authentication response with tokens
// swagger:model AuthResponse
type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// PasswordResetRequest represents a password reset request
// swagger:model PasswordResetRequest
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetConfirm represents password reset confirmation
// swagger:model PasswordResetConfirm
type PasswordResetConfirm struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// ErrorResponse represents an error response
// swagger:model ErrorResponse
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse represents a success response
// swagger:model SuccessResponse
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HealthResponse represents the health check response
// swagger:model HealthResponse
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// UsersResponse represents a list of users response
// swagger:model UsersResponse
type UsersResponse struct {
	Users []UserResponse `json:"users"`
	Count int            `json:"count"`
}
