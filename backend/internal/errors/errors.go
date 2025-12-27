// Package errors provides domain-specific error types with context for debugging.
package errors

import (
	"errors"
	"fmt"
)

// Standard sentinel errors for domain-specific error handling.
// Use errors.Is() to check for these errors.
var (
	// User-related errors
	ErrUserNotFound       = errors.New("user not found")
	ErrDuplicateEmail     = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountInactive    = errors.New("account inactive")
	ErrEmailNotVerified   = errors.New("email not verified")

	// Token-related errors
	ErrTokenExpired   = errors.New("token expired")
	ErrTokenInvalid   = errors.New("token invalid")
	ErrTokenRevoked   = errors.New("token revoked")
	ErrTokenMalformed = errors.New("token malformed")

	// Authentication/Authorization errors
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")

	// Resource errors
	ErrNotFound    = errors.New("resource not found")
	ErrConflict    = errors.New("resource conflict")
	ErrRateLimited = errors.New("rate limited")
	ErrBadRequest  = errors.New("bad request")
	ErrInternal    = errors.New("internal error")

	// Validation errors
	ErrValidation = errors.New("validation failed")

	// Database errors
	ErrDatabase     = errors.New("database error")
	ErrNoRows       = errors.New("no rows found")
	ErrDuplicateKey = errors.New("duplicate key")
)

// Kind represents the category of an error.
type Kind string

// Error kinds for categorization.
const (
	KindValidation   Kind = "validation"
	KindDatabase     Kind = "database"
	KindAuth         Kind = "auth"
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindInternal     Kind = "internal"
	KindRateLimit    Kind = "rate_limit"
	KindBadRequest   Kind = "bad_request"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
)

// DomainError wraps errors with additional context for debugging.
// It implements the error interface and supports error unwrapping.
type DomainError struct {
	// Op is the operation that failed (e.g., "auth.RegisterUser", "users.GetByID")
	Op string

	// Kind is the category of the error
	Kind Kind

	// Err is the underlying error
	Err error

	// Message is a user-friendly message (optional)
	Message string
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s: %s: %v", e.Op, e.Kind, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s: %v", e.Op, e.Kind, e.Err)
}

// Unwrap implements the errors.Unwrap interface for error chain support.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// Wrap creates a new DomainError wrapping an existing error.
func Wrap(op string, kind Kind, err error, message string) *DomainError {
	return &DomainError{
		Op:      op,
		Kind:    kind,
		Err:     err,
		Message: message,
	}
}

// WrapWithOp wraps an error with just the operation name, inferring kind from the error.
func WrapWithOp(op string, err error) *DomainError {
	return &DomainError{
		Op:   op,
		Kind: inferKind(err),
		Err:  err,
	}
}

// inferKind attempts to determine the error kind from the underlying error.
func inferKind(err error) Kind {
	switch {
	case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrNotFound), errors.Is(err, ErrNoRows):
		return KindNotFound
	case errors.Is(err, ErrDuplicateEmail), errors.Is(err, ErrConflict), errors.Is(err, ErrDuplicateKey):
		return KindConflict
	case errors.Is(err, ErrInvalidCredentials), errors.Is(err, ErrTokenExpired),
		errors.Is(err, ErrTokenInvalid), errors.Is(err, ErrTokenRevoked):
		return KindAuth
	case errors.Is(err, ErrUnauthorized), errors.Is(err, ErrAccountInactive), errors.Is(err, ErrEmailNotVerified):
		return KindUnauthorized
	case errors.Is(err, ErrForbidden):
		return KindForbidden
	case errors.Is(err, ErrValidation):
		return KindValidation
	case errors.Is(err, ErrRateLimited):
		return KindRateLimit
	case errors.Is(err, ErrBadRequest):
		return KindBadRequest
	case errors.Is(err, ErrDatabase):
		return KindDatabase
	default:
		return KindInternal
	}
}

// New creates a new DomainError with the given parameters.
func New(op string, kind Kind, message string) *DomainError {
	return &DomainError{
		Op:      op,
		Kind:    kind,
		Message: message,
	}
}

// NewValidation creates a new validation error.
func NewValidation(op string, message string) *DomainError {
	return &DomainError{
		Op:      op,
		Kind:    KindValidation,
		Err:     ErrValidation,
		Message: message,
	}
}

// NewNotFound creates a new not found error.
func NewNotFound(op string, message string) *DomainError {
	return &DomainError{
		Op:      op,
		Kind:    KindNotFound,
		Err:     ErrNotFound,
		Message: message,
	}
}

// NewUnauthorized creates a new unauthorized error.
func NewUnauthorized(op string, message string) *DomainError {
	return &DomainError{
		Op:      op,
		Kind:    KindUnauthorized,
		Err:     ErrUnauthorized,
		Message: message,
	}
}

// NewForbidden creates a new forbidden error.
func NewForbidden(op string, message string) *DomainError {
	return &DomainError{
		Op:      op,
		Kind:    KindForbidden,
		Err:     ErrForbidden,
		Message: message,
	}
}

// NewInternal creates a new internal error.
func NewInternal(op string, err error, message string) *DomainError {
	return &DomainError{
		Op:      op,
		Kind:    KindInternal,
		Err:     err,
		Message: message,
	}
}

// GetKind returns the kind of the error if it's a DomainError, otherwise returns KindInternal.
func GetKind(err error) Kind {
	var de *DomainError
	if errors.As(err, &de) {
		return de.Kind
	}
	return KindInternal
}

// GetOp returns the operation of the error if it's a DomainError, otherwise returns empty string.
func GetOp(err error) string {
	var de *DomainError
	if errors.As(err, &de) {
		return de.Op
	}
	return ""
}

// GetMessage returns the user-friendly message if it's a DomainError, otherwise returns the error string.
func GetMessage(err error) string {
	var de *DomainError
	if errors.As(err, &de) && de.Message != "" {
		return de.Message
	}
	return err.Error()
}

// Is checks if the error is of a specific kind.
func Is(err error, kind Kind) bool {
	return GetKind(err) == kind
}
