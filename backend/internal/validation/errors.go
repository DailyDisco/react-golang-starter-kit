// Package validation provides structured validation error handling for the API.
package validation

import (
	"fmt"
	"strings"
)

// FieldError represents a validation error for a single field.
// swagger:model FieldError
type FieldError struct {
	// The field name that failed validation
	// example: email
	Field string `json:"field"`

	// Human-readable error message
	// example: Invalid email format
	Message string `json:"message"`

	// Error code for programmatic handling (optional)
	// example: email
	Code string `json:"code,omitempty"`

	// The rejected value (optional, excluded for sensitive fields)
	// example: invalid@
	Value any `json:"value,omitempty"`
}

// ValidationErrors holds multiple field errors and implements the error interface.
type ValidationErrors struct {
	Errors []FieldError `json:"errors"`
}

// Error implements the error interface.
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "validation failed"
	}

	var messages []string
	for _, e := range ve.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return strings.Join(messages, "; ")
}

// Add appends a new field error to the validation errors.
func (ve *ValidationErrors) Add(field, message, code string) {
	ve.Errors = append(ve.Errors, FieldError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// AddWithValue appends a new field error with the rejected value.
func (ve *ValidationErrors) AddWithValue(field, message, code string, value any) {
	ve.Errors = append(ve.Errors, FieldError{
		Field:   field,
		Message: message,
		Code:    code,
		Value:   value,
	})
}

// HasErrors returns true if there are any validation errors.
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// Count returns the number of validation errors.
func (ve *ValidationErrors) Count() int {
	return len(ve.Errors)
}

// First returns the first field error, or nil if there are no errors.
func (ve *ValidationErrors) First() *FieldError {
	if len(ve.Errors) == 0 {
		return nil
	}
	return &ve.Errors[0]
}

// GetField returns the first error for a specific field, or nil if not found.
func (ve *ValidationErrors) GetField(field string) *FieldError {
	for i := range ve.Errors {
		if ve.Errors[i].Field == field {
			return &ve.Errors[i]
		}
	}
	return nil
}

// New creates a new empty ValidationErrors.
func New() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]FieldError, 0),
	}
}

// NewWithError creates a new ValidationErrors with a single error.
func NewWithError(field, message, code string) *ValidationErrors {
	ve := New()
	ve.Add(field, message, code)
	return ve
}

// sensitiveFields contains fields that should not have their values included in errors.
var sensitiveFields = map[string]bool{
	"password":         true,
	"password_confirm": true,
	"current_password": true,
	"new_password":     true,
	"token":            true,
	"secret":           true,
	"api_key":          true,
	"credit_card":      true,
	"ssn":              true,
}

// IsSensitiveField returns true if the field value should not be included in error responses.
func IsSensitiveField(field string) bool {
	return sensitiveFields[strings.ToLower(field)]
}
