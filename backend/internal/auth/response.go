package auth

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

// Package-private wrappers for backward compatibility
// These delegate to the shared response package

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	response.JSON(w, statusCode, data)
}

func writeSuccess(w http.ResponseWriter, message string, data interface{}) {
	response.Success(w, message, data)
}

func writeError(w http.ResponseWriter, r *http.Request, statusCode int, code string, message string) {
	response.Error(w, r, statusCode, code, message)
}

func writeBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	response.BadRequest(w, r, message)
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	response.Unauthorized(w, r, message)
}

func writeForbidden(w http.ResponseWriter, r *http.Request, message string) {
	response.Forbidden(w, r, message)
}

func writeNotFound(w http.ResponseWriter, r *http.Request, message string) {
	response.NotFound(w, r, message)
}

func writeConflict(w http.ResponseWriter, r *http.Request, message string) {
	response.Conflict(w, r, message)
}

func writeInternalError(w http.ResponseWriter, r *http.Request, message string) {
	response.InternalError(w, r, message)
}

func writeValidationError(w http.ResponseWriter, r *http.Request, message string) {
	response.ValidationError(w, r, message)
}

func writeTokenExpired(w http.ResponseWriter, r *http.Request, message string) {
	response.TokenExpired(w, r, message)
}

func writeTokenInvalid(w http.ResponseWriter, r *http.Request, message string) {
	response.TokenInvalid(w, r, message)
}

func writeEmailNotVerified(w http.ResponseWriter, r *http.Request, message string) {
	response.EmailNotVerified(w, r, message)
}

func writeAccountInactive(w http.ResponseWriter, r *http.Request, message string) {
	response.AccountInactive(w, r, message)
}
