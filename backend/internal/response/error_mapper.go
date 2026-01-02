package response

import (
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"
)

// DomainError represents a typed domain error with additional context
type DomainError struct {
	Kind    ErrorKind
	Message string
	Err     error
}

// ErrorKind categorizes domain errors for HTTP mapping
type ErrorKind int

const (
	KindUnknown ErrorKind = iota
	KindNotFound
	KindConflict
	KindForbidden
	KindUnauthorized
	KindValidation
	KindBadRequest
	KindRateLimited
)

func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a new domain error with the given kind
func NewDomainError(kind ErrorKind, message string, err error) *DomainError {
	return &DomainError{
		Kind:    kind,
		Message: message,
		Err:     err,
	}
}

// Common domain error constructors
func NewNotFoundError(message string) *DomainError {
	return NewDomainError(KindNotFound, message, nil)
}

func NewConflictError(message string) *DomainError {
	return NewDomainError(KindConflict, message, nil)
}

func NewForbiddenError(message string) *DomainError {
	return NewDomainError(KindForbidden, message, nil)
}

func NewUnauthorizedError(message string) *DomainError {
	return NewDomainError(KindUnauthorized, message, nil)
}

func NewValidationError(message string) *DomainError {
	return NewDomainError(KindValidation, message, nil)
}

func NewBadRequestError(message string) *DomainError {
	return NewDomainError(KindBadRequest, message, nil)
}

// errorMapping defines how domain errors map to HTTP responses
type errorMapping struct {
	statusCode int
	errorCode  string
}

var domainErrorMappings = map[ErrorKind]errorMapping{
	KindNotFound:     {http.StatusNotFound, ErrCodeNotFound},
	KindConflict:     {http.StatusConflict, ErrCodeConflict},
	KindForbidden:    {http.StatusForbidden, ErrCodeForbidden},
	KindUnauthorized: {http.StatusUnauthorized, ErrCodeUnauthorized},
	KindValidation:   {http.StatusBadRequest, ErrCodeValidation},
	KindBadRequest:   {http.StatusBadRequest, ErrCodeBadRequest},
	KindRateLimited:  {http.StatusTooManyRequests, ErrCodeRateLimited},
}

// sentinelErrorMapping maps specific sentinel errors to HTTP responses
var sentinelErrorMappings = map[string]errorMapping{}

// RegisterSentinelError registers a sentinel error for automatic mapping
// Call this from init() in your services package
func RegisterSentinelError(err error, statusCode int, errorCode string) {
	sentinelErrorMappings[err.Error()] = errorMapping{statusCode, errorCode}
}

// HandleError maps an error to an appropriate HTTP response.
// It handles DomainError, registered sentinel errors, and falls back to 500 for unknown errors.
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// Check for DomainError
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		if mapping, ok := domainErrorMappings[domainErr.Kind]; ok {
			Error(w, r, mapping.statusCode, mapping.errorCode, domainErr.Error())
			return
		}
	}

	// Check for registered sentinel errors
	if mapping, ok := sentinelErrorMappings[err.Error()]; ok {
		Error(w, r, mapping.statusCode, mapping.errorCode, err.Error())
		return
	}

	// Unknown error - log and return 500
	log.Error().
		Err(err).
		Str("request_id", getRequestID(r)).
		Str("path", r.URL.Path).
		Str("method", r.Method).
		Msg("Unhandled error in request")

	InternalError(w, r, "An unexpected error occurred")
}

// HandleErrorWithMessage maps an error to an HTTP response with a custom user-facing message.
// Use this when you want to hide the internal error message from users.
func HandleErrorWithMessage(w http.ResponseWriter, r *http.Request, err error, userMessage string) {
	if err == nil {
		return
	}

	// Log the original error
	log.Error().
		Err(err).
		Str("request_id", getRequestID(r)).
		Str("path", r.URL.Path).
		Str("method", r.Method).
		Str("user_message", userMessage).
		Msg("Error in request")

	// Check for DomainError to get correct status code
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		if mapping, ok := domainErrorMappings[domainErr.Kind]; ok {
			Error(w, r, mapping.statusCode, mapping.errorCode, userMessage)
			return
		}
	}

	// Check for registered sentinel errors
	if mapping, ok := sentinelErrorMappings[err.Error()]; ok {
		Error(w, r, mapping.statusCode, mapping.errorCode, userMessage)
		return
	}

	InternalError(w, r, userMessage)
}
