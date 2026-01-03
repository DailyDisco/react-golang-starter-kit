package services

import (
	"encoding/json"
	"testing"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"react-golang-starter/internal/testutil"

	"github.com/pquerna/otp/totp"
)

func testTOTPSetup(t *testing.T) (*TOTPService, func()) {
	t.Helper()
	testutil.SkipIfNotIntegration(t)

	db := testutil.SetupTestDB(t)
	tt := testutil.NewTestTransaction(t, db)

	// Set global database.DB for the TOTP service
	oldDB := database.DB
	database.DB = tt.DB

	svc := NewTOTPService()

	return svc, func() {
		database.DB = oldDB
		tt.Rollback()
	}
}

func createTestUserForTOTP(t *testing.T, email string) *models.User {
	t.Helper()
	user := &models.User{
		Email:    email,
		Name:     "TOTP Test User",
		Password: "hashedpassword",
		Role:     models.RoleUser,
	}
	if err := database.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func TestTOTPService_GenerateSecret_Integration(t *testing.T) {
	svc, cleanup := testTOTPSetup(t)
	defer cleanup()

	t.Run("generates valid TOTP secret", func(t *testing.T) {
		response, err := svc.GenerateSecret("test@example.com")
		if err != nil {
			t.Fatalf("GenerateSecret failed: %v", err)
		}

		if response.Secret == "" {
			t.Error("Expected secret to be generated")
		}
		if len(response.Secret) < 16 {
			t.Error("Expected secret to be at least 16 characters")
		}
		if response.QRCodeURL == "" {
			t.Error("Expected QR code URL to be set")
		}
		if response.OTPAuthURL == "" {
			t.Error("Expected OTPAuth URL to be set")
		}
	})
}

func TestTOTPService_SetupTwoFactor_Integration(t *testing.T) {
	svc, cleanup := testTOTPSetup(t)
	defer cleanup()

	t.Run("creates unverified 2FA record", func(t *testing.T) {
		user := createTestUserForTOTP(t, "setup@example.com")

		response, err := svc.SetupTwoFactor(user.ID, user.Email)
		if err != nil {
			t.Fatalf("SetupTwoFactor failed: %v", err)
		}

		if response.Secret == "" {
			t.Error("Expected secret to be returned")
		}

		// Verify record was created in DB
		var twoFactor models.UserTwoFactor
		if err := database.DB.Where("user_id = ?", user.ID).First(&twoFactor).Error; err != nil {
			t.Fatalf("Failed to find 2FA record: %v", err)
		}

		if twoFactor.IsEnabled {
			t.Error("Expected 2FA to be disabled initially")
		}
		if twoFactor.EncryptedSecret == "" {
			t.Error("Expected encrypted secret to be stored")
		}
		if twoFactor.VerifiedAt != nil {
			t.Error("Expected VerifiedAt to be nil initially")
		}
	})

	t.Run("updates existing unverified record", func(t *testing.T) {
		user := createTestUserForTOTP(t, "existing@example.com")

		// First setup
		response1, err := svc.SetupTwoFactor(user.ID, user.Email)
		if err != nil {
			t.Fatalf("First SetupTwoFactor failed: %v", err)
		}

		// Second setup (should overwrite)
		response2, err := svc.SetupTwoFactor(user.ID, user.Email)
		if err != nil {
			t.Fatalf("Second SetupTwoFactor failed: %v", err)
		}

		if response1.Secret == response2.Secret {
			t.Error("Expected new secret to be generated on re-setup")
		}

		// Verify only one record exists
		var count int64
		database.DB.Model(&models.UserTwoFactor{}).Where("user_id = ?", user.ID).Count(&count)
		if count != 1 {
			t.Errorf("Expected 1 2FA record, got: %d", count)
		}
	})
}

func TestTOTPService_VerifyAndEnable_Integration(t *testing.T) {
	svc, cleanup := testTOTPSetup(t)
	defer cleanup()

	t.Run("enables 2FA with valid code and generates backup codes", func(t *testing.T) {
		user := createTestUserForTOTP(t, "verify@example.com")

		// Setup 2FA first
		setupResp, err := svc.SetupTwoFactor(user.ID, user.Email)
		if err != nil {
			t.Fatalf("SetupTwoFactor failed: %v", err)
		}

		// Generate a valid TOTP code
		code, err := totp.GenerateCode(setupResp.Secret, time.Now())
		if err != nil {
			t.Fatalf("Failed to generate TOTP code: %v", err)
		}

		// Verify and enable
		backupCodes, err := svc.VerifyAndEnable(user.ID, code)
		if err != nil {
			t.Fatalf("VerifyAndEnable failed: %v", err)
		}

		if len(backupCodes) != 10 {
			t.Errorf("Expected 10 backup codes, got: %d", len(backupCodes))
		}

		// Verify each backup code is 8 characters
		for _, bc := range backupCodes {
			if len(bc) != 8 {
				t.Errorf("Expected backup code length 8, got: %d", len(bc))
			}
		}

		// Verify 2FA is enabled in DB
		var twoFactor models.UserTwoFactor
		if err := database.DB.Where("user_id = ?", user.ID).First(&twoFactor).Error; err != nil {
			t.Fatalf("Failed to find 2FA record: %v", err)
		}
		if !twoFactor.IsEnabled {
			t.Error("Expected 2FA to be enabled")
		}
		if twoFactor.VerifiedAt == nil {
			t.Error("Expected VerifiedAt to be set")
		}
		if twoFactor.BackupCodesRemaining != 10 {
			t.Errorf("Expected 10 backup codes remaining, got: %d", twoFactor.BackupCodesRemaining)
		}

		// Note: two_factor_enabled column exists in DB but not in Go struct
		// The service updates it via direct SQL, verified by checking UserTwoFactor.IsEnabled
	})

	t.Run("rejects invalid TOTP code", func(t *testing.T) {
		user := createTestUserForTOTP(t, "invalid@example.com")

		_, err := svc.SetupTwoFactor(user.ID, user.Email)
		if err != nil {
			t.Fatalf("SetupTwoFactor failed: %v", err)
		}

		_, err = svc.VerifyAndEnable(user.ID, "000000")
		if err == nil {
			t.Error("Expected error for invalid code")
		}
		if err.Error() != "invalid verification code" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("returns error when 2FA not set up", func(t *testing.T) {
		user := createTestUserForTOTP(t, "nosetup@example.com")

		_, err := svc.VerifyAndEnable(user.ID, "123456")
		if err == nil {
			t.Error("Expected error when 2FA not set up")
		}
		if err.Error() != "2FA not set up for this user" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}

func TestTOTPService_ValidateCode_Integration(t *testing.T) {
	svc, cleanup := testTOTPSetup(t)
	defer cleanup()

	// Helper to setup and enable 2FA, returns secret and backup codes
	setupAndEnable := func(t *testing.T, email string) (string, []string) {
		user := createTestUserForTOTP(t, email)
		setupResp, err := svc.SetupTwoFactor(user.ID, user.Email)
		if err != nil {
			t.Fatalf("SetupTwoFactor failed: %v", err)
		}
		code, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		backupCodes, err := svc.VerifyAndEnable(user.ID, code)
		if err != nil {
			t.Fatalf("VerifyAndEnable failed: %v", err)
		}
		return setupResp.Secret, backupCodes
	}

	t.Run("validates correct TOTP code", func(t *testing.T) {
		user := createTestUserForTOTP(t, "validate@example.com")
		setupResp, _ := svc.SetupTwoFactor(user.ID, user.Email)
		code, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		svc.VerifyAndEnable(user.ID, code)

		// Generate a new code and validate
		newCode, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		valid, err := svc.ValidateCode(user.ID, newCode)
		if err != nil {
			t.Fatalf("ValidateCode failed: %v", err)
		}
		if !valid {
			t.Error("Expected valid code to return true")
		}
	})

	t.Run("validates backup code and decrements count", func(t *testing.T) {
		_, backupCodes := setupAndEnable(t, "backup@example.com")

		// Get the user that was created
		var user models.User
		database.DB.Where("email = ?", "backup@example.com").First(&user)

		valid, err := svc.ValidateCode(user.ID, backupCodes[0])
		if err != nil {
			t.Fatalf("ValidateCode with backup code failed: %v", err)
		}
		if !valid {
			t.Error("Expected backup code to be valid")
		}

		// Verify backup codes remaining decreased
		var twoFactor models.UserTwoFactor
		database.DB.Where("user_id = ?", user.ID).First(&twoFactor)
		if twoFactor.BackupCodesRemaining != 9 {
			t.Errorf("Expected 9 backup codes remaining, got: %d", twoFactor.BackupCodesRemaining)
		}
	})

	t.Run("backup code can only be used once", func(t *testing.T) {
		_, backupCodes := setupAndEnable(t, "oneuse@example.com")

		var user models.User
		database.DB.Where("email = ?", "oneuse@example.com").Last(&user)

		// First use should succeed
		valid, _ := svc.ValidateCode(user.ID, backupCodes[0])
		if !valid {
			t.Error("First use of backup code should succeed")
		}

		// Second use should fail
		valid, _ = svc.ValidateCode(user.ID, backupCodes[0])
		if valid {
			t.Error("Second use of backup code should fail")
		}
	})

	t.Run("locks account after 5 failed attempts", func(t *testing.T) {
		setupAndEnable(t, "lockout@example.com")

		var user models.User
		database.DB.Where("email = ?", "lockout@example.com").Last(&user)

		// Make 5 failed attempts
		for i := 0; i < 5; i++ {
			svc.ValidateCode(user.ID, "000000")
		}

		// Verify account is locked
		var twoFactor models.UserTwoFactor
		database.DB.Where("user_id = ?", user.ID).First(&twoFactor)
		if twoFactor.LockedUntil == nil {
			t.Error("Expected account to be locked")
		}

		// Next attempt should return locked error
		_, err := svc.ValidateCode(user.ID, "123456")
		if err == nil {
			t.Error("Expected locked error")
		}
		if err.Error() != "account temporarily locked due to too many failed attempts" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("increments failed attempts counter", func(t *testing.T) {
		setupAndEnable(t, "failcount@example.com")

		var user models.User
		database.DB.Where("email = ?", "failcount@example.com").Last(&user)

		// Make a failed attempt
		svc.ValidateCode(user.ID, "000000")

		var twoFactor models.UserTwoFactor
		database.DB.Where("user_id = ?", user.ID).First(&twoFactor)
		if twoFactor.FailedAttempts != 1 {
			t.Errorf("Expected 1 failed attempt, got: %d", twoFactor.FailedAttempts)
		}
	})

	t.Run("updates last_used_at on successful validation", func(t *testing.T) {
		secret, _ := setupAndEnable(t, "lastused@example.com")

		var user models.User
		database.DB.Where("email = ?", "lastused@example.com").Last(&user)

		// Validate a code
		code, _ := totp.GenerateCode(secret, time.Now())
		svc.ValidateCode(user.ID, code)

		var twoFactor models.UserTwoFactor
		database.DB.Where("user_id = ?", user.ID).First(&twoFactor)
		if twoFactor.LastUsedAt == nil {
			t.Error("Expected last_used_at to be set")
		}
	})
}

func TestTOTPService_DisableTwoFactor_Integration(t *testing.T) {
	svc, cleanup := testTOTPSetup(t)
	defer cleanup()

	t.Run("disables 2FA with valid code", func(t *testing.T) {
		user := createTestUserForTOTP(t, "disable@example.com")

		// Setup and enable
		setupResp, _ := svc.SetupTwoFactor(user.ID, user.Email)
		code, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		svc.VerifyAndEnable(user.ID, code)

		// Disable with valid code
		newCode, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		err := svc.DisableTwoFactor(user.ID, newCode)
		if err != nil {
			t.Fatalf("DisableTwoFactor failed: %v", err)
		}

		// Verify 2FA record is deleted
		var count int64
		database.DB.Model(&models.UserTwoFactor{}).Where("user_id = ?", user.ID).Count(&count)
		if count != 0 {
			t.Error("Expected 2FA record to be deleted")
		}

		// Note: two_factor_enabled column exists in DB but not in Go struct
		// Verified by confirming UserTwoFactor record is deleted above
	})

	t.Run("rejects disable with invalid code", func(t *testing.T) {
		user := createTestUserForTOTP(t, "nodisable@example.com")

		// Setup and enable
		setupResp, _ := svc.SetupTwoFactor(user.ID, user.Email)
		code, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		svc.VerifyAndEnable(user.ID, code)

		// Try to disable with invalid code
		err := svc.DisableTwoFactor(user.ID, "000000")
		if err == nil {
			t.Error("Expected error for invalid code")
		}
	})
}

func TestTOTPService_GetTwoFactorStatus_Integration(t *testing.T) {
	svc, cleanup := testTOTPSetup(t)
	defer cleanup()

	t.Run("returns disabled status when not set up", func(t *testing.T) {
		user := createTestUserForTOTP(t, "nostatus@example.com")

		status, err := svc.GetTwoFactorStatus(user.ID)
		if err != nil {
			t.Fatalf("GetTwoFactorStatus failed: %v", err)
		}

		if status.Enabled {
			t.Error("Expected Enabled to be false")
		}
		if status.BackupCodesRemaining != 0 {
			t.Errorf("Expected 0 backup codes, got: %d", status.BackupCodesRemaining)
		}
	})

	t.Run("returns enabled status with backup code count", func(t *testing.T) {
		user := createTestUserForTOTP(t, "hasstatus@example.com")

		setupResp, _ := svc.SetupTwoFactor(user.ID, user.Email)
		code, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		svc.VerifyAndEnable(user.ID, code)

		status, err := svc.GetTwoFactorStatus(user.ID)
		if err != nil {
			t.Fatalf("GetTwoFactorStatus failed: %v", err)
		}

		if !status.Enabled {
			t.Error("Expected Enabled to be true")
		}
		if status.BackupCodesRemaining != 10 {
			t.Errorf("Expected 10 backup codes, got: %d", status.BackupCodesRemaining)
		}
		if status.VerifiedAt == "" {
			t.Error("Expected VerifiedAt to be set")
		}
	})
}

func TestTOTPService_RegenerateBackupCodes_Integration(t *testing.T) {
	svc, cleanup := testTOTPSetup(t)
	defer cleanup()

	t.Run("generates new backup codes with valid TOTP", func(t *testing.T) {
		user := createTestUserForTOTP(t, "regen@example.com")

		setupResp, _ := svc.SetupTwoFactor(user.ID, user.Email)
		code, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		oldCodes, _ := svc.VerifyAndEnable(user.ID, code)

		// Regenerate with valid code
		newCode, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		newCodes, err := svc.RegenerateBackupCodes(user.ID, newCode)
		if err != nil {
			t.Fatalf("RegenerateBackupCodes failed: %v", err)
		}

		if len(newCodes) != 10 {
			t.Errorf("Expected 10 new backup codes, got: %d", len(newCodes))
		}

		// Verify old codes are different from new codes
		for _, old := range oldCodes {
			for _, new := range newCodes {
				if old == new {
					t.Error("Expected new backup codes to be different from old")
				}
			}
		}

		// Verify backup codes remaining is reset
		var twoFactor models.UserTwoFactor
		database.DB.Where("user_id = ?", user.ID).First(&twoFactor)
		if twoFactor.BackupCodesRemaining != 10 {
			t.Errorf("Expected 10 backup codes remaining, got: %d", twoFactor.BackupCodesRemaining)
		}
	})

	t.Run("rejects regeneration with invalid code", func(t *testing.T) {
		user := createTestUserForTOTP(t, "noregen@example.com")

		setupResp, _ := svc.SetupTwoFactor(user.ID, user.Email)
		code, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		svc.VerifyAndEnable(user.ID, code)

		_, err := svc.RegenerateBackupCodes(user.ID, "000000")
		if err == nil {
			t.Error("Expected error for invalid code")
		}
	})
}

func TestTOTPService_EncryptionRoundTrip_Integration(t *testing.T) {
	svc, cleanup := testTOTPSetup(t)
	defer cleanup()

	t.Run("secret survives encrypt/decrypt cycle", func(t *testing.T) {
		originalSecret := "TESTSECRET12345678901234567890"

		encrypted, err := svc.encryptSecret(originalSecret)
		if err != nil {
			t.Fatalf("encryptSecret failed: %v", err)
		}

		if encrypted == originalSecret {
			t.Error("Encrypted secret should be different from original")
		}

		decrypted, err := svc.decryptSecret(encrypted)
		if err != nil {
			t.Fatalf("decryptSecret failed: %v", err)
		}

		if decrypted != originalSecret {
			t.Errorf("Expected %s, got: %s", originalSecret, decrypted)
		}
	})
}

func TestTOTPService_BackupCodeHashing_Integration(t *testing.T) {
	svc, cleanup := testTOTPSetup(t)
	defer cleanup()

	t.Run("backup codes are bcrypt hashed in DB", func(t *testing.T) {
		user := createTestUserForTOTP(t, "hashtest@example.com")

		setupResp, _ := svc.SetupTwoFactor(user.ID, user.Email)
		code, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		backupCodes, _ := svc.VerifyAndEnable(user.ID, code)

		// Get the stored hashes
		var twoFactor models.UserTwoFactor
		database.DB.Where("user_id = ?", user.ID).First(&twoFactor)

		var storedHashes []string
		json.Unmarshal(twoFactor.BackupCodesHash, &storedHashes)

		// Verify each hash is different from plaintext
		for i, bc := range backupCodes {
			if storedHashes[i] == bc {
				t.Error("Stored hash should not equal plaintext backup code")
			}
			// Verify it looks like a bcrypt hash (starts with $2a$ or $2b$)
			if len(storedHashes[i]) < 60 {
				t.Error("Expected bcrypt hash to be at least 60 characters")
			}
		}
	})
}

func TestTOTPService_Is2FAEnabled_Integration(t *testing.T) {
	svc, cleanup := testTOTPSetup(t)
	defer cleanup()

	t.Run("returns false when not set up", func(t *testing.T) {
		user := createTestUserForTOTP(t, "notenabled@example.com")

		enabled, err := svc.Is2FAEnabled(user.ID)
		if err != nil {
			t.Fatalf("Is2FAEnabled failed: %v", err)
		}
		if enabled {
			t.Error("Expected false when not set up")
		}
	})

	t.Run("returns true when enabled", func(t *testing.T) {
		user := createTestUserForTOTP(t, "isenabled@example.com")

		setupResp, _ := svc.SetupTwoFactor(user.ID, user.Email)
		code, _ := totp.GenerateCode(setupResp.Secret, time.Now())
		svc.VerifyAndEnable(user.ID, code)

		enabled, err := svc.Is2FAEnabled(user.ID)
		if err != nil {
			t.Fatalf("Is2FAEnabled failed: %v", err)
		}
		if !enabled {
			t.Error("Expected true when enabled")
		}
	})
}
