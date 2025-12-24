package handlers

import (
	"net/http"

	"react-golang-starter/internal/response"
)

// Re-export error codes from shared response package for backward compatibility
const (
	ErrCodeUnauthorized     = response.ErrCodeUnauthorized
	ErrCodeForbidden        = response.ErrCodeForbidden
	ErrCodeNotFound         = response.ErrCodeNotFound
	ErrCodeBadRequest       = response.ErrCodeBadRequest
	ErrCodeConflict         = response.ErrCodeConflict
	ErrCodeInternalError    = response.ErrCodeInternalError
	ErrCodeValidation       = response.ErrCodeValidation
	ErrCodeRateLimited      = response.ErrCodeRateLimited
	ErrCodeTokenExpired     = response.ErrCodeTokenExpired
	ErrCodeTokenInvalid     = response.ErrCodeTokenInvalid
	ErrCodeEmailNotVerified = response.ErrCodeEmailNotVerified
	ErrCodeAccountInactive  = response.ErrCodeAccountInactive
)

// Public wrappers for backward compatibility
// These delegate to the shared response package

func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	response.JSON(w, statusCode, data)
}

func WriteSuccess(w http.ResponseWriter, message string, data interface{}) {
	response.Success(w, message, data)
}

func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, code string, message string) {
	response.Error(w, r, statusCode, code, message)
}

func WriteBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	response.BadRequest(w, r, message)
}

func WriteUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	response.Unauthorized(w, r, message)
}

func WriteForbidden(w http.ResponseWriter, r *http.Request, message string) {
	response.Forbidden(w, r, message)
}

func WriteNotFound(w http.ResponseWriter, r *http.Request, message string) {
	response.NotFound(w, r, message)
}

func WriteConflict(w http.ResponseWriter, r *http.Request, message string) {
	response.Conflict(w, r, message)
}

func WriteInternalError(w http.ResponseWriter, r *http.Request, message string) {
	response.InternalError(w, r, message)
}

func WriteValidationError(w http.ResponseWriter, r *http.Request, message string) {
	response.ValidationError(w, r, message)
}

func WriteRateLimited(w http.ResponseWriter, r *http.Request) {
	response.RateLimited(w, r)
}
