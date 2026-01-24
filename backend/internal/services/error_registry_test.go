package services

import (
	"testing"
)

// ============ Error Registry Tests ============

// TestErrorRegistration_Init verifies that the init() function runs without panicking
// The init() function in error_registry.go registers sentinel errors with the response package
func TestErrorRegistration_Init(t *testing.T) {
	// This test verifies that the init() function in error_registry.go
	// successfully registers all sentinel errors. If it panics, this test fails.
	// The fact that we can run tests in this package means init() succeeded.

	// Verify all org errors exist and are not nil
	orgErrors := []error{
		ErrOrgNotFound,
		ErrOrgSlugTaken,
		ErrInvalidSlug,
		ErrNotMember,
		ErrInsufficientRole,
		ErrCannotRemoveOwner,
		ErrInvitationNotFound,
		ErrInvitationExpired,
		ErrInvitationAccepted,
		ErrAlreadyMember,
		ErrCannotChangeOwnRole,
		ErrMustHaveOwner,
		ErrInvitationEmailTaken,
		ErrSeatLimitExceeded,
	}

	for _, err := range orgErrors {
		if err == nil {
			t.Error("Found nil error in org errors list")
		}
	}

	// Verify file errors exist
	if ErrAccessDenied == nil {
		t.Error("ErrAccessDenied is nil")
	}
}

// ============ Error Messages Tests ============

func TestErrorMessages_Org(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantMessage string
	}{
		{"ErrOrgNotFound", ErrOrgNotFound, "organization not found"},
		{"ErrOrgSlugTaken", ErrOrgSlugTaken, "organization slug is already taken"},
		{"ErrInvalidSlug", ErrInvalidSlug, "invalid slug format"},
		{"ErrNotMember", ErrNotMember, "user is not a member of this organization"},
		{"ErrInsufficientRole", ErrInsufficientRole, "insufficient role permissions"},
		{"ErrCannotRemoveOwner", ErrCannotRemoveOwner, "cannot remove the organization owner"},
		{"ErrInvitationNotFound", ErrInvitationNotFound, "invitation not found"},
		{"ErrInvitationExpired", ErrInvitationExpired, "invitation has expired"},
		{"ErrInvitationAccepted", ErrInvitationAccepted, "invitation has already been accepted"},
		{"ErrAlreadyMember", ErrAlreadyMember, "user is already a member"},
		{"ErrCannotChangeOwnRole", ErrCannotChangeOwnRole, "cannot change your own role"},
		{"ErrMustHaveOwner", ErrMustHaveOwner, "organization must have at least one owner"},
		{"ErrInvitationEmailTaken", ErrInvitationEmailTaken, "an invitation for this email already exists"},
		{"ErrSeatLimitExceeded", ErrSeatLimitExceeded, "organization has reached its seat limit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.wantMessage {
				t.Errorf("%s message = %q, want %q", tt.name, tt.err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestErrorMessages_File(t *testing.T) {
	if ErrAccessDenied.Error() != "access denied" {
		t.Errorf("ErrAccessDenied.Error() = %q, want %q", ErrAccessDenied.Error(), "access denied")
	}
}

// ============ Sentinel Error Identity Tests ============

func TestSentinelErrors_AreUnique(t *testing.T) {
	errors := []error{
		ErrOrgNotFound,
		ErrOrgSlugTaken,
		ErrInvalidSlug,
		ErrNotMember,
		ErrInsufficientRole,
		ErrCannotRemoveOwner,
		ErrInvitationNotFound,
		ErrInvitationExpired,
		ErrInvitationAccepted,
		ErrAlreadyMember,
		ErrCannotChangeOwnRole,
		ErrMustHaveOwner,
		ErrInvitationEmailTaken,
		ErrSeatLimitExceeded,
		ErrAccessDenied,
	}

	// Verify all errors are unique by message
	seen := make(map[string]bool)
	for _, err := range errors {
		msg := err.Error()
		if seen[msg] {
			t.Errorf("Duplicate error message found: %q", msg)
		}
		seen[msg] = true
	}
}

func TestSentinelErrors_AreNotNil(t *testing.T) {
	errors := []struct {
		name string
		err  error
	}{
		{"ErrOrgNotFound", ErrOrgNotFound},
		{"ErrOrgSlugTaken", ErrOrgSlugTaken},
		{"ErrInvalidSlug", ErrInvalidSlug},
		{"ErrNotMember", ErrNotMember},
		{"ErrInsufficientRole", ErrInsufficientRole},
		{"ErrCannotRemoveOwner", ErrCannotRemoveOwner},
		{"ErrInvitationNotFound", ErrInvitationNotFound},
		{"ErrInvitationExpired", ErrInvitationExpired},
		{"ErrInvitationAccepted", ErrInvitationAccepted},
		{"ErrAlreadyMember", ErrAlreadyMember},
		{"ErrCannotChangeOwnRole", ErrCannotChangeOwnRole},
		{"ErrMustHaveOwner", ErrMustHaveOwner},
		{"ErrInvitationEmailTaken", ErrInvitationEmailTaken},
		{"ErrSeatLimitExceeded", ErrSeatLimitExceeded},
		{"ErrAccessDenied", ErrAccessDenied},
	}

	for _, tt := range errors {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil", tt.name)
			}
		})
	}
}

// ============ Error Pointer Identity Tests ============

func TestSentinelErrors_AreDifferentPointers(t *testing.T) {
	// Verify errors are distinct pointers (not accidentally the same)
	errors := []error{
		ErrOrgNotFound,
		ErrOrgSlugTaken,
		ErrInvalidSlug,
		ErrNotMember,
		ErrInsufficientRole,
		ErrCannotRemoveOwner,
		ErrInvitationNotFound,
		ErrInvitationExpired,
		ErrInvitationAccepted,
		ErrAlreadyMember,
		ErrCannotChangeOwnRole,
		ErrMustHaveOwner,
		ErrInvitationEmailTaken,
		ErrSeatLimitExceeded,
	}

	for i, err1 := range errors {
		for j, err2 := range errors {
			if i != j && err1 == err2 {
				t.Errorf("Errors at index %d and %d are the same pointer", i, j)
			}
		}
	}
}
