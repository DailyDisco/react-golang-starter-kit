// Package response provides shared HTTP response helpers for consistent API responses.
package response

import (
	"encoding/json"
	"net/http"

	"react-golang-starter/internal/contextkeys"
	"react-golang-starter/internal/models"
)

// getRequestID retrieves the request ID from the request context.
func getRequestID(r *http.Request) string {
	if requestID, ok := r.Context().Value(contextkeys.RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// Common error codes for frontend handling
const (
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeBadRequest       = "BAD_REQUEST"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeInternalError    = "INTERNAL_ERROR"
	ErrCodeValidation       = "VALIDATION_ERROR"
	ErrCodeRateLimited      = "RATE_LIMITED"
	ErrCodeTokenExpired     = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid     = "TOKEN_INVALID"
	ErrCodeEmailNotVerified = "EMAIL_NOT_VERIFIED"
	ErrCodeAccountInactive  = "ACCOUNT_INACTIVE"
)

// JSON writes a JSON response with the given status code
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// Success writes a success response
func Success(w http.ResponseWriter, message string, data interface{}) {
	JSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error writes an error response with consistent format (includes request ID from context)
func Error(w http.ResponseWriter, r *http.Request, statusCode int, code string, message string) {
	requestID := getRequestID(r)
	JSON(w, statusCode, models.ErrorResponse{
		Error:     code,
		Message:   message,
		Code:      statusCode,
		RequestID: requestID,
	})
}

// Common error responses - all include request ID for traceability

// BadRequest writes a 400 Bad Request response
func BadRequest(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusBadRequest, ErrCodeBadRequest, message)
}

// Unauthorized writes a 401 Unauthorized response
func Unauthorized(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

// Forbidden writes a 403 Forbidden response
func Forbidden(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusForbidden, ErrCodeForbidden, message)
}

// NotFound writes a 404 Not Found response
func NotFound(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusNotFound, ErrCodeNotFound, message)
}

// Conflict writes a 409 Conflict response
func Conflict(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusConflict, ErrCodeConflict, message)
}

// InternalError writes a 500 Internal Server Error response
func InternalError(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusInternalServerError, ErrCodeInternalError, message)
}

// ValidationError writes a 400 Bad Request response with validation error code
func ValidationError(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusBadRequest, ErrCodeValidation, message)
}

// ValidationErrorWithDetails writes a 400 Bad Request response with field-level error details.
// This provides structured validation errors that clients can use for field-specific error handling.
func ValidationErrorWithDetails(w http.ResponseWriter, r *http.Request, message string, details []models.FieldError) {
	requestID := getRequestID(r)
	JSON(w, http.StatusBadRequest, models.ErrorResponse{
		Error:     ErrCodeValidation,
		Message:   message,
		Code:      http.StatusBadRequest,
		RequestID: requestID,
		Details:   details,
	})
}

// RateLimited writes a 429 Too Many Requests response
func RateLimited(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusTooManyRequests, ErrCodeRateLimited, "Too many requests. Please try again later.")
}

// TokenExpired writes a 401 Unauthorized response with token expired code
func TokenExpired(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusUnauthorized, ErrCodeTokenExpired, message)
}

// TokenInvalid writes a 401 Unauthorized response with token invalid code
func TokenInvalid(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusUnauthorized, ErrCodeTokenInvalid, message)
}

// EmailNotVerified writes a 400 Bad Request response with email not verified code
func EmailNotVerified(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusBadRequest, ErrCodeEmailNotVerified, message)
}

// AccountInactive writes a 401 Unauthorized response with account inactive code
func AccountInactive(w http.ResponseWriter, r *http.Request, message string) {
	Error(w, r, http.StatusUnauthorized, ErrCodeAccountInactive, message)
}
