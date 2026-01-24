package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-golang-starter/internal/models"
)

// ============ ErrorKind Constants Tests ============

func TestErrorKindConstants(t *testing.T) {
	tests := []struct {
		name string
		kind ErrorKind
		want int
	}{
		{"unknown", KindUnknown, 0},
		{"not found", KindNotFound, 1},
		{"conflict", KindConflict, 2},
		{"forbidden", KindForbidden, 3},
		{"unauthorized", KindUnauthorized, 4},
		{"validation", KindValidation, 5},
		{"bad request", KindBadRequest, 6},
		{"rate limited", KindRateLimited, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.kind) != tt.want {
				t.Errorf("ErrorKind = %d, want %d", int(tt.kind), tt.want)
			}
		})
	}
}

// ============ DomainError Tests ============

func TestDomainError_Error_WithMessage(t *testing.T) {
	err := &DomainError{
		Kind:    KindNotFound,
		Message: "User not found",
	}

	if err.Error() != "User not found" {
		t.Errorf("Error() = %q, want %q", err.Error(), "User not found")
	}
}

func TestDomainError_Error_WithWrappedError(t *testing.T) {
	wrappedErr := errors.New("database connection failed")
	err := &DomainError{
		Kind: KindUnknown,
		Err:  wrappedErr,
	}

	if err.Error() != "database connection failed" {
		t.Errorf("Error() = %q, want %q", err.Error(), "database connection failed")
	}
}

func TestDomainError_Error_EmptyMessageAndErr(t *testing.T) {
	err := &DomainError{
		Kind: KindUnknown,
	}

	if err.Error() != "unknown error" {
		t.Errorf("Error() = %q, want %q", err.Error(), "unknown error")
	}
}

func TestDomainError_Error_MessageTakesPrecedence(t *testing.T) {
	err := &DomainError{
		Kind:    KindNotFound,
		Message: "Custom message",
		Err:     errors.New("wrapped error"),
	}

	// Message should take precedence over wrapped error
	if err.Error() != "Custom message" {
		t.Errorf("Error() = %q, want %q", err.Error(), "Custom message")
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("original error")
	err := &DomainError{
		Kind: KindUnknown,
		Err:  wrappedErr,
	}

	if err.Unwrap() != wrappedErr {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), wrappedErr)
	}
}

func TestDomainError_Unwrap_Nil(t *testing.T) {
	err := &DomainError{
		Kind:    KindNotFound,
		Message: "Not found",
	}

	if err.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", err.Unwrap())
	}
}

// ============ NewDomainError Tests ============

func TestNewDomainError(t *testing.T) {
	wrappedErr := errors.New("wrapped")
	err := NewDomainError(KindConflict, "conflict message", wrappedErr)

	if err.Kind != KindConflict {
		t.Errorf("Kind = %v, want %v", err.Kind, KindConflict)
	}

	if err.Message != "conflict message" {
		t.Errorf("Message = %q, want %q", err.Message, "conflict message")
	}

	if err.Err != wrappedErr {
		t.Errorf("Err = %v, want %v", err.Err, wrappedErr)
	}
}

// ============ Domain Error Constructor Tests ============

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("resource not found")

	if err.Kind != KindNotFound {
		t.Errorf("Kind = %v, want %v", err.Kind, KindNotFound)
	}

	if err.Message != "resource not found" {
		t.Errorf("Message = %q, want %q", err.Message, "resource not found")
	}

	if err.Err != nil {
		t.Errorf("Err = %v, want nil", err.Err)
	}
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("already exists")

	if err.Kind != KindConflict {
		t.Errorf("Kind = %v, want %v", err.Kind, KindConflict)
	}

	if err.Message != "already exists" {
		t.Errorf("Message = %q, want %q", err.Message, "already exists")
	}
}

func TestNewForbiddenError(t *testing.T) {
	err := NewForbiddenError("access denied")

	if err.Kind != KindForbidden {
		t.Errorf("Kind = %v, want %v", err.Kind, KindForbidden)
	}

	if err.Message != "access denied" {
		t.Errorf("Message = %q, want %q", err.Message, "access denied")
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("not authenticated")

	if err.Kind != KindUnauthorized {
		t.Errorf("Kind = %v, want %v", err.Kind, KindUnauthorized)
	}

	if err.Message != "not authenticated" {
		t.Errorf("Message = %q, want %q", err.Message, "not authenticated")
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("invalid input")

	if err.Kind != KindValidation {
		t.Errorf("Kind = %v, want %v", err.Kind, KindValidation)
	}

	if err.Message != "invalid input" {
		t.Errorf("Message = %q, want %q", err.Message, "invalid input")
	}
}

func TestNewBadRequestError(t *testing.T) {
	err := NewBadRequestError("bad request")

	if err.Kind != KindBadRequest {
		t.Errorf("Kind = %v, want %v", err.Kind, KindBadRequest)
	}

	if err.Message != "bad request" {
		t.Errorf("Message = %q, want %q", err.Message, "bad request")
	}
}

// ============ RegisterSentinelError Tests ============

func TestRegisterSentinelError(t *testing.T) {
	testErr := errors.New("test_sentinel_error")

	RegisterSentinelError(testErr, http.StatusTeapot, "TEST_ERROR")

	// Verify it was registered
	if mapping, ok := sentinelErrorMappings[testErr.Error()]; ok {
		if mapping.statusCode != http.StatusTeapot {
			t.Errorf("statusCode = %d, want %d", mapping.statusCode, http.StatusTeapot)
		}
		if mapping.errorCode != "TEST_ERROR" {
			t.Errorf("errorCode = %q, want %q", mapping.errorCode, "TEST_ERROR")
		}
	} else {
		t.Error("sentinel error was not registered")
	}

	// Clean up
	delete(sentinelErrorMappings, testErr.Error())
}

// ============ HandleError Tests ============

func TestHandleError_NilError(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	HandleError(w, req, nil)

	// Should not write anything
	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d (default)", w.Code, http.StatusOK)
	}

	if w.Body.Len() != 0 {
		t.Errorf("body length = %d, want 0", w.Body.Len())
	}
}

func TestHandleError_DomainError(t *testing.T) {
	tests := []struct {
		name           string
		kind           ErrorKind
		wantStatusCode int
		wantErrorCode  string
	}{
		{"not found", KindNotFound, http.StatusNotFound, ErrCodeNotFound},
		{"conflict", KindConflict, http.StatusConflict, ErrCodeConflict},
		{"forbidden", KindForbidden, http.StatusForbidden, ErrCodeForbidden},
		{"unauthorized", KindUnauthorized, http.StatusUnauthorized, ErrCodeUnauthorized},
		{"validation", KindValidation, http.StatusBadRequest, ErrCodeValidation},
		{"bad request", KindBadRequest, http.StatusBadRequest, ErrCodeBadRequest},
		{"rate limited", KindRateLimited, http.StatusTooManyRequests, ErrCodeRateLimited},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			err := NewDomainError(tt.kind, "test error message", nil)

			HandleError(w, req, err)

			if w.Code != tt.wantStatusCode {
				t.Errorf("status code = %d, want %d", w.Code, tt.wantStatusCode)
			}

			var result models.ErrorResponse
			json.Unmarshal(w.Body.Bytes(), &result)

			if result.Error != tt.wantErrorCode {
				t.Errorf("Error = %q, want %q", result.Error, tt.wantErrorCode)
			}

			if result.Message != "test error message" {
				t.Errorf("Message = %q, want %q", result.Message, "test error message")
			}
		})
	}
}

func TestHandleError_SentinelError(t *testing.T) {
	// Register a test sentinel error
	testErr := errors.New("unique_test_error")
	RegisterSentinelError(testErr, http.StatusBadGateway, "BAD_GATEWAY")
	defer delete(sentinelErrorMappings, testErr.Error())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	HandleError(w, req, testErr)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadGateway)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != "BAD_GATEWAY" {
		t.Errorf("Error = %q, want %q", result.Error, "BAD_GATEWAY")
	}
}

func TestHandleError_UnknownError(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	err := errors.New("some unknown error")

	HandleError(w, req, err)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Error != ErrCodeInternalError {
		t.Errorf("Error = %q, want %q", result.Error, ErrCodeInternalError)
	}

	// Message should be generic for unknown errors
	if result.Message != "An unexpected error occurred" {
		t.Errorf("Message = %q, want generic message", result.Message)
	}
}

func TestHandleError_DomainErrorUnknownKind(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	err := NewDomainError(KindUnknown, "unknown kind error", nil)

	HandleError(w, req, err)

	// Unknown kind should fall through to internal error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// ============ HandleErrorWithMessage Tests ============

func TestHandleErrorWithMessage_NilError(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	HandleErrorWithMessage(w, req, nil, "custom message")

	// Should not write anything
	if w.Body.Len() != 0 {
		t.Errorf("body length = %d, want 0", w.Body.Len())
	}
}

func TestHandleErrorWithMessage_DomainError(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	err := NewNotFoundError("internal error details")

	HandleErrorWithMessage(w, req, err, "User-friendly message")

	if w.Code != http.StatusNotFound {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusNotFound)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	// Should use the user message, not the internal error
	if result.Message != "User-friendly message" {
		t.Errorf("Message = %q, want %q", result.Message, "User-friendly message")
	}
}

func TestHandleErrorWithMessage_SentinelError(t *testing.T) {
	// Register a test sentinel error
	testErr := errors.New("another_sentinel_error")
	RegisterSentinelError(testErr, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE")
	defer delete(sentinelErrorMappings, testErr.Error())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	HandleErrorWithMessage(w, req, testErr, "Custom user message")

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Message != "Custom user message" {
		t.Errorf("Message = %q, want %q", result.Message, "Custom user message")
	}
}

func TestHandleErrorWithMessage_UnknownError(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	err := errors.New("random unknown error")

	HandleErrorWithMessage(w, req, err, "Something went wrong")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var result models.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.Message != "Something went wrong" {
		t.Errorf("Message = %q, want %q", result.Message, "Something went wrong")
	}
}

// ============ Domain Error Mappings Tests ============

func TestDomainErrorMappings_AllKindsHaveMapping(t *testing.T) {
	kindsToTest := []ErrorKind{
		KindNotFound,
		KindConflict,
		KindForbidden,
		KindUnauthorized,
		KindValidation,
		KindBadRequest,
		KindRateLimited,
	}

	for _, kind := range kindsToTest {
		if _, ok := domainErrorMappings[kind]; !ok {
			t.Errorf("ErrorKind %d has no mapping", kind)
		}
	}
}

// ============ Error Wrapping/Unwrapping Tests ============

func TestErrorsAs_WithDomainError(t *testing.T) {
	wrappedErr := NewNotFoundError("resource missing")

	var domainErr *DomainError
	if !errors.As(wrappedErr, &domainErr) {
		t.Error("errors.As should work with DomainError")
	}

	if domainErr.Kind != KindNotFound {
		t.Errorf("Kind = %v, want %v", domainErr.Kind, KindNotFound)
	}
}
