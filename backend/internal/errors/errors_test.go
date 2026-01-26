package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Sentinel Errors Tests ---

func TestSentinelErrors_AreDistinct(t *testing.T) {
	// All sentinel errors should be unique
	errs := []error{
		ErrUserNotFound,
		ErrDuplicateEmail,
		ErrInvalidCredentials,
		ErrAccountInactive,
		ErrEmailNotVerified,
		ErrTokenExpired,
		ErrTokenInvalid,
		ErrTokenRevoked,
		ErrTokenMalformed,
		ErrUnauthorized,
		ErrForbidden,
		ErrNotFound,
		ErrConflict,
		ErrRateLimited,
		ErrBadRequest,
		ErrInternal,
		ErrValidation,
		ErrDatabase,
		ErrNoRows,
		ErrDuplicateKey,
	}

	// Check that all errors are unique
	seen := make(map[string]bool)
	for _, err := range errs {
		msg := err.Error()
		if seen[msg] {
			t.Errorf("Duplicate error message: %s", msg)
		}
		seen[msg] = true
	}
}

func TestSentinelErrors_ErrorMethod(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrUserNotFound", ErrUserNotFound, "user not found"},
		{"ErrDuplicateEmail", ErrDuplicateEmail, "email already exists"},
		{"ErrInvalidCredentials", ErrInvalidCredentials, "invalid credentials"},
		{"ErrAccountInactive", ErrAccountInactive, "account inactive"},
		{"ErrEmailNotVerified", ErrEmailNotVerified, "email not verified"},
		{"ErrTokenExpired", ErrTokenExpired, "token expired"},
		{"ErrTokenInvalid", ErrTokenInvalid, "token invalid"},
		{"ErrTokenRevoked", ErrTokenRevoked, "token revoked"},
		{"ErrTokenMalformed", ErrTokenMalformed, "token malformed"},
		{"ErrUnauthorized", ErrUnauthorized, "unauthorized"},
		{"ErrForbidden", ErrForbidden, "forbidden"},
		{"ErrNotFound", ErrNotFound, "resource not found"},
		{"ErrConflict", ErrConflict, "resource conflict"},
		{"ErrRateLimited", ErrRateLimited, "rate limited"},
		{"ErrBadRequest", ErrBadRequest, "bad request"},
		{"ErrInternal", ErrInternal, "internal error"},
		{"ErrValidation", ErrValidation, "validation failed"},
		{"ErrDatabase", ErrDatabase, "database error"},
		{"ErrNoRows", ErrNoRows, "no rows found"},
		{"ErrDuplicateKey", ErrDuplicateKey, "duplicate key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

// --- DomainError Tests ---

func TestDomainError_Error_WithMessage(t *testing.T) {
	underlyingErr := errors.New("connection refused")
	de := &DomainError{
		Op:      "users.GetByID",
		Kind:    KindDatabase,
		Message: "failed to retrieve user",
		Err:     underlyingErr,
	}

	result := de.Error()

	assert.Equal(t, "users.GetByID: database: failed to retrieve user: connection refused", result)
}

func TestDomainError_Error_WithoutMessage(t *testing.T) {
	underlyingErr := errors.New("not found")
	de := &DomainError{
		Op:   "users.GetByID",
		Kind: KindNotFound,
		Err:  underlyingErr,
	}

	result := de.Error()

	assert.Equal(t, "users.GetByID: not_found: not found", result)
}

func TestDomainError_Unwrap(t *testing.T) {
	underlyingErr := ErrUserNotFound
	de := &DomainError{
		Op:   "users.GetByID",
		Kind: KindNotFound,
		Err:  underlyingErr,
	}

	unwrapped := de.Unwrap()

	assert.Equal(t, underlyingErr, unwrapped)
}

func TestDomainError_ErrorsIs(t *testing.T) {
	de := &DomainError{
		Op:   "auth.Login",
		Kind: KindNotFound,
		Err:  ErrUserNotFound,
	}

	// errors.Is should work through Unwrap
	assert.True(t, errors.Is(de, ErrUserNotFound))
	assert.False(t, errors.Is(de, ErrDuplicateEmail))
}

func TestDomainError_ErrorsAs(t *testing.T) {
	de := &DomainError{
		Op:      "auth.Login",
		Kind:    KindAuth,
		Message: "login failed",
		Err:     ErrInvalidCredentials,
	}

	var target *DomainError
	require.True(t, errors.As(de, &target))
	assert.Equal(t, "auth.Login", target.Op)
	assert.Equal(t, KindAuth, target.Kind)
	assert.Equal(t, "login failed", target.Message)
}

func TestDomainError_NestedUnwrap(t *testing.T) {
	// Create a nested error chain
	inner := &DomainError{
		Op:   "repo.Save",
		Kind: KindDatabase,
		Err:  ErrDatabase,
	}
	outer := &DomainError{
		Op:   "service.Create",
		Kind: KindInternal,
		Err:  inner,
	}

	// Should be able to find ErrDatabase through the chain
	assert.True(t, errors.Is(outer, ErrDatabase))

	// Should be able to extract the inner DomainError
	var target *DomainError
	require.True(t, errors.As(outer, &target))
	// errors.As returns the first matching type
	assert.Equal(t, "service.Create", target.Op)
}

// --- Wrap Function Tests ---

func TestWrap_CreatesCorrectDomainError(t *testing.T) {
	underlyingErr := errors.New("connection failed")

	de := Wrap("db.Connect", KindDatabase, underlyingErr, "failed to connect to database")

	require.NotNil(t, de)
	assert.Equal(t, "db.Connect", de.Op)
	assert.Equal(t, KindDatabase, de.Kind)
	assert.Equal(t, "failed to connect to database", de.Message)
	assert.Equal(t, underlyingErr, de.Err)
}

func TestWrap_WithEmptyMessage(t *testing.T) {
	underlyingErr := errors.New("some error")

	de := Wrap("op.Name", KindInternal, underlyingErr, "")

	assert.Empty(t, de.Message)
	assert.Equal(t, "op.Name: internal: some error", de.Error())
}

// --- WrapWithOp Tests ---

func TestWrapWithOp_CreatesCorrectDomainError(t *testing.T) {
	de := WrapWithOp("users.Create", ErrDuplicateEmail)

	require.NotNil(t, de)
	assert.Equal(t, "users.Create", de.Op)
	assert.Equal(t, KindConflict, de.Kind)
	assert.Equal(t, ErrDuplicateEmail, de.Err)
}

// --- inferKind Tests ---

func TestInferKind(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Kind
	}{
		// NotFound errors
		{"ErrUserNotFound", ErrUserNotFound, KindNotFound},
		{"ErrNotFound", ErrNotFound, KindNotFound},
		{"ErrNoRows", ErrNoRows, KindNotFound},

		// Conflict errors
		{"ErrDuplicateEmail", ErrDuplicateEmail, KindConflict},
		{"ErrConflict", ErrConflict, KindConflict},
		{"ErrDuplicateKey", ErrDuplicateKey, KindConflict},

		// Auth errors (token-related)
		{"ErrInvalidCredentials", ErrInvalidCredentials, KindAuth},
		{"ErrTokenExpired", ErrTokenExpired, KindAuth},
		{"ErrTokenInvalid", ErrTokenInvalid, KindAuth},
		{"ErrTokenRevoked", ErrTokenRevoked, KindAuth},

		// Unauthorized errors
		{"ErrUnauthorized", ErrUnauthorized, KindUnauthorized},
		{"ErrAccountInactive", ErrAccountInactive, KindUnauthorized},
		{"ErrEmailNotVerified", ErrEmailNotVerified, KindUnauthorized},

		// Forbidden errors
		{"ErrForbidden", ErrForbidden, KindForbidden},

		// Validation errors
		{"ErrValidation", ErrValidation, KindValidation},

		// Rate limit errors
		{"ErrRateLimited", ErrRateLimited, KindRateLimit},

		// Bad request errors
		{"ErrBadRequest", ErrBadRequest, KindBadRequest},

		// Database errors
		{"ErrDatabase", ErrDatabase, KindDatabase},

		// Unknown errors default to Internal
		{"unknown error", errors.New("random error"), KindInternal},
		{"wrapped unknown", fmt.Errorf("wrapped: %w", errors.New("unknown")), KindInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferKind(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInferKind_WrappedErrors(t *testing.T) {
	// errors.Is should work with wrapped errors
	wrappedNotFound := fmt.Errorf("user lookup failed: %w", ErrUserNotFound)
	wrappedDuplicate := fmt.Errorf("create failed: %w", ErrDuplicateEmail)

	assert.Equal(t, KindNotFound, inferKind(wrappedNotFound))
	assert.Equal(t, KindConflict, inferKind(wrappedDuplicate))
}

// --- Factory Function Tests ---

func TestNew_CreatesDomainError(t *testing.T) {
	de := New("op.Test", KindValidation, "test message")

	require.NotNil(t, de)
	assert.Equal(t, "op.Test", de.Op)
	assert.Equal(t, KindValidation, de.Kind)
	assert.Equal(t, "test message", de.Message)
	assert.Nil(t, de.Err)
}

func TestNewValidation_SetsCorrectKindAndError(t *testing.T) {
	de := NewValidation("form.Validate", "email is required")

	require.NotNil(t, de)
	assert.Equal(t, "form.Validate", de.Op)
	assert.Equal(t, KindValidation, de.Kind)
	assert.Equal(t, "email is required", de.Message)
	assert.True(t, errors.Is(de, ErrValidation))
}

func TestNewNotFound_SetsCorrectKindAndError(t *testing.T) {
	de := NewNotFound("users.GetByID", "user 123 not found")

	require.NotNil(t, de)
	assert.Equal(t, "users.GetByID", de.Op)
	assert.Equal(t, KindNotFound, de.Kind)
	assert.Equal(t, "user 123 not found", de.Message)
	assert.True(t, errors.Is(de, ErrNotFound))
}

func TestNewUnauthorized_SetsCorrectKindAndError(t *testing.T) {
	de := NewUnauthorized("auth.Verify", "invalid session")

	require.NotNil(t, de)
	assert.Equal(t, "auth.Verify", de.Op)
	assert.Equal(t, KindUnauthorized, de.Kind)
	assert.Equal(t, "invalid session", de.Message)
	assert.True(t, errors.Is(de, ErrUnauthorized))
}

func TestNewForbidden_SetsCorrectKindAndError(t *testing.T) {
	de := NewForbidden("admin.Delete", "insufficient permissions")

	require.NotNil(t, de)
	assert.Equal(t, "admin.Delete", de.Op)
	assert.Equal(t, KindForbidden, de.Kind)
	assert.Equal(t, "insufficient permissions", de.Message)
	assert.True(t, errors.Is(de, ErrForbidden))
}

func TestNewInternal_PreservesUnderlyingError(t *testing.T) {
	underlyingErr := errors.New("database connection lost")

	de := NewInternal("db.Query", underlyingErr, "query failed")

	require.NotNil(t, de)
	assert.Equal(t, "db.Query", de.Op)
	assert.Equal(t, KindInternal, de.Kind)
	assert.Equal(t, "query failed", de.Message)
	assert.Equal(t, underlyingErr, de.Err)
	assert.True(t, errors.Is(de, underlyingErr))
}

// --- Getter Function Tests ---

func TestGetKind_FromDomainError(t *testing.T) {
	de := &DomainError{
		Op:   "test.Op",
		Kind: KindConflict,
		Err:  ErrDuplicateEmail,
	}

	kind := GetKind(de)

	assert.Equal(t, KindConflict, kind)
}

func TestGetKind_FromRegularError(t *testing.T) {
	regularErr := errors.New("some error")

	kind := GetKind(regularErr)

	assert.Equal(t, KindInternal, kind)
}

func TestGetKind_FromWrappedDomainError(t *testing.T) {
	de := &DomainError{
		Op:   "test.Op",
		Kind: KindNotFound,
		Err:  ErrUserNotFound,
	}
	wrapped := fmt.Errorf("wrapped: %w", de)

	kind := GetKind(wrapped)

	assert.Equal(t, KindNotFound, kind)
}

func TestGetOp_FromDomainError(t *testing.T) {
	de := &DomainError{
		Op:   "users.Create",
		Kind: KindValidation,
		Err:  ErrValidation,
	}

	op := GetOp(de)

	assert.Equal(t, "users.Create", op)
}

func TestGetOp_FromRegularError(t *testing.T) {
	regularErr := errors.New("some error")

	op := GetOp(regularErr)

	assert.Empty(t, op)
}

func TestGetOp_FromWrappedDomainError(t *testing.T) {
	de := &DomainError{
		Op:   "auth.Login",
		Kind: KindAuth,
		Err:  ErrInvalidCredentials,
	}
	wrapped := fmt.Errorf("wrapped: %w", de)

	op := GetOp(wrapped)

	assert.Equal(t, "auth.Login", op)
}

func TestGetMessage_FromDomainErrorWithMessage(t *testing.T) {
	de := &DomainError{
		Op:      "test.Op",
		Kind:    KindValidation,
		Message: "custom message",
		Err:     ErrValidation,
	}

	msg := GetMessage(de)

	assert.Equal(t, "custom message", msg)
}

func TestGetMessage_FromDomainErrorWithoutMessage(t *testing.T) {
	de := &DomainError{
		Op:   "test.Op",
		Kind: KindNotFound,
		Err:  ErrUserNotFound,
	}

	msg := GetMessage(de)

	// Should return the full error string
	assert.Equal(t, de.Error(), msg)
}

func TestGetMessage_FromRegularError(t *testing.T) {
	regularErr := errors.New("regular error message")

	msg := GetMessage(regularErr)

	assert.Equal(t, "regular error message", msg)
}

// --- Is Function Tests ---

func TestIs_MatchesKind(t *testing.T) {
	de := &DomainError{
		Op:   "test.Op",
		Kind: KindValidation,
		Err:  ErrValidation,
	}

	assert.True(t, Is(de, KindValidation))
}

func TestIs_NoMatch(t *testing.T) {
	de := &DomainError{
		Op:   "test.Op",
		Kind: KindValidation,
		Err:  ErrValidation,
	}

	assert.False(t, Is(de, KindNotFound))
	assert.False(t, Is(de, KindAuth))
	assert.False(t, Is(de, KindInternal))
}

func TestIs_WithRegularError(t *testing.T) {
	regularErr := errors.New("some error")

	// Regular errors default to KindInternal
	assert.True(t, Is(regularErr, KindInternal))
	assert.False(t, Is(regularErr, KindValidation))
}

func TestIs_WithWrappedDomainError(t *testing.T) {
	de := &DomainError{
		Op:   "test.Op",
		Kind: KindConflict,
		Err:  ErrDuplicateEmail,
	}
	wrapped := fmt.Errorf("wrapped: %w", de)

	assert.True(t, Is(wrapped, KindConflict))
	assert.False(t, Is(wrapped, KindNotFound))
}

// --- Kind Constants Tests ---

func TestKindConstants(t *testing.T) {
	// Ensure all kind constants are distinct strings
	kinds := []Kind{
		KindValidation,
		KindDatabase,
		KindAuth,
		KindNotFound,
		KindConflict,
		KindInternal,
		KindRateLimit,
		KindBadRequest,
		KindUnauthorized,
		KindForbidden,
	}

	seen := make(map[Kind]bool)
	for _, k := range kinds {
		if seen[k] {
			t.Errorf("Duplicate kind constant: %s", k)
		}
		seen[k] = true
		// Ensure it's not empty
		assert.NotEmpty(t, string(k))
	}
}

// --- Edge Cases ---

func TestDomainError_NilErr(t *testing.T) {
	de := &DomainError{
		Op:      "test.Op",
		Kind:    KindValidation,
		Message: "validation failed",
		Err:     nil,
	}

	// Should not panic
	result := de.Error()
	assert.Contains(t, result, "test.Op")
	assert.Contains(t, result, "validation")

	// Unwrap should return nil
	assert.Nil(t, de.Unwrap())
}

func TestDomainError_EmptyOp(t *testing.T) {
	de := &DomainError{
		Op:   "",
		Kind: KindInternal,
		Err:  errors.New("test"),
	}

	result := de.Error()
	assert.Contains(t, result, "internal")
}

func TestWrapWithOp_NilError(t *testing.T) {
	// Wrapping nil should still create a DomainError
	de := WrapWithOp("test.Op", nil)

	require.NotNil(t, de)
	assert.Equal(t, "test.Op", de.Op)
	assert.Equal(t, KindInternal, de.Kind)
	assert.Nil(t, de.Err)
}
