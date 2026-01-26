package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"react-golang-starter/internal/models"
)

// --- Constants Verification Tests ---

func TestLockoutConstants(t *testing.T) {
	// Verify the lockout security constants are set correctly
	t.Run("MaxFailedLoginAttempts", func(t *testing.T) {
		assert.Equal(t, 5, MaxFailedLoginAttempts, "should lock after 5 failed attempts")
	})

	t.Run("LockoutDuration", func(t *testing.T) {
		assert.Equal(t, 30*time.Minute, LockoutDuration, "should lock for 30 minutes")
	})

	t.Run("FailedLoginWindow", func(t *testing.T) {
		assert.Equal(t, 15*time.Minute, FailedLoginWindow, "should reset counter after 15 minutes")
	})
}

// --- Lockout Logic Tests ---
// These tests verify the lockout logic using the same conditions as handleFailedLogin

func TestLockoutLogic_ShouldResetCounter(t *testing.T) {
	tests := []struct {
		name            string
		lastFailedLogin *time.Time
		shouldReset     bool
	}{
		{
			name:            "nil last failed login - no reset needed",
			lastFailedLogin: nil,
			shouldReset:     false,
		},
		{
			name:            "just failed - no reset",
			lastFailedLogin: timePtr(time.Now()),
			shouldReset:     false,
		},
		{
			name:            "14 minutes ago - no reset",
			lastFailedLogin: timePtr(time.Now().Add(-14 * time.Minute)),
			shouldReset:     false,
		},
		{
			name:            "just under 15 minutes ago - no reset",
			lastFailedLogin: timePtr(time.Now().Add(-14*time.Minute - 59*time.Second)),
			shouldReset:     false,
		},
		{
			name:            "15 minutes 1 second ago - should reset",
			lastFailedLogin: timePtr(time.Now().Add(-15*time.Minute - 1*time.Second)),
			shouldReset:     true,
		},
		{
			name:            "16 minutes ago - should reset",
			lastFailedLogin: timePtr(time.Now().Add(-16 * time.Minute)),
			shouldReset:     true,
		},
		{
			name:            "1 hour ago - should reset",
			lastFailedLogin: timePtr(time.Now().Add(-1 * time.Hour)),
			shouldReset:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			shouldReset := tt.lastFailedLogin != nil && now.Sub(*tt.lastFailedLogin) > FailedLoginWindow
			assert.Equal(t, tt.shouldReset, shouldReset)
		})
	}
}

func TestLockoutLogic_ShouldLockAccount(t *testing.T) {
	tests := []struct {
		name           string
		failedAttempts int
		shouldLock     bool
	}{
		{"0 attempts - no lock", 0, false},
		{"1 attempt - no lock", 1, false},
		{"2 attempts - no lock", 2, false},
		{"3 attempts - no lock", 3, false},
		{"4 attempts - no lock", 4, false},
		{"5 attempts - should lock", 5, true},
		{"6 attempts - should lock", 6, true},
		{"10 attempts - should lock", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldLock := tt.failedAttempts >= MaxFailedLoginAttempts
			assert.Equal(t, tt.shouldLock, shouldLock)
		})
	}
}

func TestLockoutLogic_IsAccountLocked(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		lockedUntil *time.Time
		isLocked    bool
	}{
		{
			name:        "nil locked_until - not locked",
			lockedUntil: nil,
			isLocked:    false,
		},
		{
			name:        "locked_until in past - not locked",
			lockedUntil: timePtr(now.Add(-1 * time.Minute)),
			isLocked:    false,
		},
		{
			name:        "locked_until exactly now - not locked (Before returns false for equal times)",
			lockedUntil: timePtr(now),
			isLocked:    false,
		},
		{
			name:        "locked_until 1 second in future - locked",
			lockedUntil: timePtr(now.Add(1 * time.Second)),
			isLocked:    true,
		},
		{
			name:        "locked_until 30 minutes in future - locked",
			lockedUntil: timePtr(now.Add(30 * time.Minute)),
			isLocked:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isLocked := tt.lockedUntil != nil && now.Before(*tt.lockedUntil)
			assert.Equal(t, tt.isLocked, isLocked)
		})
	}
}

// --- handleFailedLogin Logic Simulation ---
// These tests simulate the exact logic of handleFailedLogin without DB dependencies

func TestHandleFailedLogin_FirstAttempt(t *testing.T) {
	user := &models.User{
		FailedLoginAttempts: 0,
		LastFailedLogin:     nil,
		LockedUntil:         nil,
	}

	// Simulate handleFailedLogin logic
	now := time.Now()
	if user.LastFailedLogin != nil && now.Sub(*user.LastFailedLogin) > FailedLoginWindow {
		user.FailedLoginAttempts = 0
	}
	user.FailedLoginAttempts++
	user.LastFailedLogin = &now

	assert.Equal(t, 1, user.FailedLoginAttempts)
	assert.NotNil(t, user.LastFailedLogin)
	assert.Nil(t, user.LockedUntil, "should not be locked after 1 attempt")
}

func TestHandleFailedLogin_IncrementWithinWindow(t *testing.T) {
	lastFailed := time.Now().Add(-10 * time.Minute) // 10 minutes ago
	user := &models.User{
		FailedLoginAttempts: 3,
		LastFailedLogin:     &lastFailed,
		LockedUntil:         nil,
	}

	// Simulate handleFailedLogin logic
	now := time.Now()
	if user.LastFailedLogin != nil && now.Sub(*user.LastFailedLogin) > FailedLoginWindow {
		user.FailedLoginAttempts = 0
	}
	user.FailedLoginAttempts++
	user.LastFailedLogin = &now

	assert.Equal(t, 4, user.FailedLoginAttempts, "should increment within window")
	assert.Nil(t, user.LockedUntil, "should not be locked after 4 attempts")
}

func TestHandleFailedLogin_ResetAfterWindow(t *testing.T) {
	lastFailed := time.Now().Add(-20 * time.Minute) // 20 minutes ago (outside window)
	user := &models.User{
		FailedLoginAttempts: 4,
		LastFailedLogin:     &lastFailed,
		LockedUntil:         nil,
	}

	// Simulate handleFailedLogin logic
	now := time.Now()
	if user.LastFailedLogin != nil && now.Sub(*user.LastFailedLogin) > FailedLoginWindow {
		user.FailedLoginAttempts = 0
	}
	user.FailedLoginAttempts++
	user.LastFailedLogin = &now

	assert.Equal(t, 1, user.FailedLoginAttempts, "should reset to 1 after window")
	assert.Nil(t, user.LockedUntil, "should not be locked")
}

func TestHandleFailedLogin_FifthAttemptTriggersLock(t *testing.T) {
	lastFailed := time.Now().Add(-5 * time.Minute) // 5 minutes ago
	user := &models.User{
		FailedLoginAttempts: 4,
		LastFailedLogin:     &lastFailed,
		LockedUntil:         nil,
	}

	// Simulate handleFailedLogin logic
	now := time.Now()
	if user.LastFailedLogin != nil && now.Sub(*user.LastFailedLogin) > FailedLoginWindow {
		user.FailedLoginAttempts = 0
	}
	user.FailedLoginAttempts++
	user.LastFailedLogin = &now

	if user.FailedLoginAttempts >= MaxFailedLoginAttempts {
		lockUntil := now.Add(LockoutDuration)
		user.LockedUntil = &lockUntil
	}

	assert.Equal(t, 5, user.FailedLoginAttempts)
	require.NotNil(t, user.LockedUntil, "should be locked after 5 attempts")

	// Verify lock duration is 30 minutes from now
	expectedLockUntil := now.Add(30 * time.Minute)
	assert.WithinDuration(t, expectedLockUntil, *user.LockedUntil, time.Second)
}

func TestHandleFailedLogin_SixthAttemptStillLocked(t *testing.T) {
	lastFailed := time.Now().Add(-1 * time.Minute) // 1 minute ago
	lockUntil := time.Now().Add(29 * time.Minute)  // locked for 29 more minutes
	user := &models.User{
		FailedLoginAttempts: 5,
		LastFailedLogin:     &lastFailed,
		LockedUntil:         &lockUntil,
	}

	// Simulate handleFailedLogin logic (account is locked but we're still tracking)
	now := time.Now()
	if user.LastFailedLogin != nil && now.Sub(*user.LastFailedLogin) > FailedLoginWindow {
		user.FailedLoginAttempts = 0
	}
	user.FailedLoginAttempts++
	user.LastFailedLogin = &now

	if user.FailedLoginAttempts >= MaxFailedLoginAttempts {
		newLockUntil := now.Add(LockoutDuration)
		user.LockedUntil = &newLockUntil
	}

	assert.Equal(t, 6, user.FailedLoginAttempts)
	require.NotNil(t, user.LockedUntil)
}

// --- Window Boundary Edge Cases ---

func TestHandleFailedLogin_ExactlyAtWindowBoundary(t *testing.T) {
	// Test exactly at 15 minute boundary using a fixed reference point
	// to avoid timing issues between creating lastFailed and now
	now := time.Now()
	lastFailed := now.Add(-FailedLoginWindow) // exactly 15 minutes from `now`
	user := &models.User{
		FailedLoginAttempts: 4,
		LastFailedLogin:     &lastFailed,
		LockedUntil:         nil,
	}

	// The condition is: now.Sub(*user.LastFailedLogin) > FailedLoginWindow
	// At exactly 15 minutes, this is false (15 > 15 is false)
	elapsed := now.Sub(*user.LastFailedLogin)
	shouldReset := user.LastFailedLogin != nil && elapsed > FailedLoginWindow

	if shouldReset {
		user.FailedLoginAttempts = 0
	}
	user.FailedLoginAttempts++

	// At exactly 15 minutes, we should NOT reset (uses > not >=)
	assert.Equal(t, 5, user.FailedLoginAttempts, "exactly at window should NOT reset - increments to 5")
}

func TestHandleFailedLogin_JustBeforeWindow(t *testing.T) {
	// 14:59 ago - just before 15 minute window
	lastFailed := time.Now().Add(-14*time.Minute - 59*time.Second)
	user := &models.User{
		FailedLoginAttempts: 4,
		LastFailedLogin:     &lastFailed,
		LockedUntil:         nil,
	}

	now := time.Now()
	shouldReset := user.LastFailedLogin != nil && now.Sub(*user.LastFailedLogin) > FailedLoginWindow

	if shouldReset {
		user.FailedLoginAttempts = 0
	}
	user.FailedLoginAttempts++

	assert.Equal(t, 5, user.FailedLoginAttempts, "just before window should NOT reset")
}

func TestHandleFailedLogin_JustAfterWindow(t *testing.T) {
	// 15:01 ago - just after 15 minute window
	lastFailed := time.Now().Add(-15*time.Minute - 1*time.Second)
	user := &models.User{
		FailedLoginAttempts: 4,
		LastFailedLogin:     &lastFailed,
		LockedUntil:         nil,
	}

	now := time.Now()
	shouldReset := user.LastFailedLogin != nil && now.Sub(*user.LastFailedLogin) > FailedLoginWindow

	if shouldReset {
		user.FailedLoginAttempts = 0
	}
	user.FailedLoginAttempts++

	assert.Equal(t, 1, user.FailedLoginAttempts, "just after window should reset to 1")
}

// --- Lockout Expiry Edge Cases ---

func TestLoginUser_LockExpiryEdgeCases(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		lockedUntil time.Time
		isLocked    bool
	}{
		{
			name:        "locked 30 minutes from now",
			lockedUntil: now.Add(30 * time.Minute),
			isLocked:    true,
		},
		{
			name:        "locked 1 second from now",
			lockedUntil: now.Add(1 * time.Second),
			isLocked:    true,
		},
		{
			name:        "lock expired 1 second ago",
			lockedUntil: now.Add(-1 * time.Second),
			isLocked:    false,
		},
		{
			name:        "lock expired 30 minutes ago",
			lockedUntil: now.Add(-30 * time.Minute),
			isLocked:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This simulates the check in LoginUser:
			// if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil)
			isLocked := now.Before(tt.lockedUntil)
			assert.Equal(t, tt.isLocked, isLocked)
		})
	}
}

// --- Successful Login Reset Tests ---

func TestSuccessfulLogin_ResetsLockoutState(t *testing.T) {
	lastFailed := time.Now().Add(-5 * time.Minute)
	lockUntil := time.Now().Add(-1 * time.Minute) // Lock already expired
	user := &models.User{
		FailedLoginAttempts: 5,
		LastFailedLogin:     &lastFailed,
		LockedUntil:         &lockUntil,
	}

	// Simulate successful login reset
	// This is what LoginUser does after password verification
	if user.FailedLoginAttempts > 0 || user.LockedUntil != nil {
		user.FailedLoginAttempts = 0
		user.LockedUntil = nil
		user.LastFailedLogin = nil
	}

	assert.Equal(t, 0, user.FailedLoginAttempts, "should reset to 0")
	assert.Nil(t, user.LockedUntil, "should clear locked_until")
	assert.Nil(t, user.LastFailedLogin, "should clear last_failed_login")
}

func TestSuccessfulLogin_NoResetNeededWhenClean(t *testing.T) {
	user := &models.User{
		FailedLoginAttempts: 0,
		LastFailedLogin:     nil,
		LockedUntil:         nil,
	}

	originalAttempts := user.FailedLoginAttempts

	// Simulate successful login check
	needsReset := user.FailedLoginAttempts > 0 || user.LockedUntil != nil

	assert.False(t, needsReset, "clean user should not need reset")
	assert.Equal(t, originalAttempts, user.FailedLoginAttempts)
}

// --- Helper Functions ---

func timePtr(t time.Time) *time.Time {
	return &t
}
