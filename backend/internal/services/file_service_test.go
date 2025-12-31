package services

import (
	"testing"
)

// ============ File Service Error Tests ============

func TestErrAccessDenied(t *testing.T) {
	if ErrAccessDenied.Error() != "access denied" {
		t.Errorf("ErrAccessDenied.Error() = %q, want %q", ErrAccessDenied.Error(), "access denied")
	}
}
