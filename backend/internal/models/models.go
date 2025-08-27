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
}
