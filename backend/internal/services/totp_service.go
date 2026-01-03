package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TOTPService handles two-factor authentication operations
type TOTPService struct {
	issuer        string
	encryptionKey []byte
}

// NewTOTPService creates a new TOTP service instance
func NewTOTPService() *TOTPService {
	issuer := os.Getenv("TOTP_ISSUER")
	if issuer == "" {
		issuer = os.Getenv("SITE_NAME")
		if issuer == "" {
			issuer = "MyApp"
		}
	}

	// Get encryption key from environment
	encKeyStr := os.Getenv("TOTP_ENCRYPTION_KEY")
	if encKeyStr == "" {
		encKeyStr = os.Getenv("JWT_SECRET") // Fallback to JWT secret
	}

	// Create a 32-byte key from the secret using SHA-256
	hash := sha256.Sum256([]byte(encKeyStr))

	return &TOTPService{
		issuer:        issuer,
		encryptionKey: hash[:],
	}
}

// GenerateSecret generates a new TOTP secret for a user
func (s *TOTPService) GenerateSecret(email string) (*models.TwoFactorSetupResponse, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
		Period:      30,
		SecretSize:  32,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	return &models.TwoFactorSetupResponse{
		Secret:     key.Secret(),
		QRCodeURL:  key.URL(),
		OTPAuthURL: key.URL(),
	}, nil
}

// SetupTwoFactor initiates 2FA setup for a user
func (s *TOTPService) SetupTwoFactor(userID uint, email string) (*models.TwoFactorSetupResponse, error) {
	// Generate new secret
	setup, err := s.GenerateSecret(email)
	if err != nil {
		return nil, err
	}

	// Encrypt the secret
	encryptedSecret, err := s.encryptSecret(setup.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// Check if 2FA record exists
	var existing models.UserTwoFactor
	result := database.DB.Where("user_id = ?", userID).First(&existing)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new record
		twoFactor := &models.UserTwoFactor{
			UserID:          userID,
			EncryptedSecret: encryptedSecret,
			IsEnabled:       false,
			CreatedAt:       time.Now().Format(time.RFC3339),
			UpdatedAt:       time.Now().Format(time.RFC3339),
		}
		if err := database.DB.Create(twoFactor).Error; err != nil {
			return nil, fmt.Errorf("failed to create 2FA record: %w", err)
		}
	} else if result.Error != nil {
		return nil, fmt.Errorf("failed to check existing 2FA: %w", result.Error)
	} else {
		// Update existing record with new secret
		existing.EncryptedSecret = encryptedSecret
		existing.IsEnabled = false
		existing.VerifiedAt = nil
		existing.UpdatedAt = time.Now().Format(time.RFC3339)
		if err := database.DB.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("failed to update 2FA record: %w", err)
		}
	}

	return setup, nil
}

// VerifyAndEnable verifies the TOTP code and enables 2FA
func (s *TOTPService) VerifyAndEnable(userID uint, code string) ([]string, error) {
	var twoFactor models.UserTwoFactor
	if err := database.DB.Where("user_id = ?", userID).First(&twoFactor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("2FA not set up for this user")
		}
		return nil, fmt.Errorf("failed to retrieve 2FA record: %w", err)
	}

	// Decrypt the secret
	secret, err := s.decryptSecret(twoFactor.EncryptedSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Validate the code
	valid := totp.Validate(code, secret)
	if !valid {
		// Increment failed attempts
		database.DB.Model(&twoFactor).UpdateColumn("failed_attempts", gorm.Expr("failed_attempts + 1"))
		return nil, fmt.Errorf("invalid verification code")
	}

	// Generate backup codes
	backupCodes, hashedCodes, err := s.generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	hashedCodesJSON, _ := json.Marshal(hashedCodes)

	// Enable 2FA
	verifiedAt := time.Now().Format(time.RFC3339)
	updates := map[string]interface{}{
		"is_enabled":             true,
		"verified_at":            verifiedAt,
		"backup_codes_hash":      hashedCodesJSON,
		"backup_codes_remaining": 10,
		"failed_attempts":        0,
		"updated_at":             time.Now().Format(time.RFC3339),
	}

	if err := database.DB.Model(&twoFactor).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to enable 2FA: %w", err)
	}

	// Update user's 2FA enabled flag
	database.DB.Model(&models.User{}).Where("id = ?", userID).Update("two_factor_enabled", true)

	return backupCodes, nil
}

// ValidateCode validates a TOTP code or backup code
func (s *TOTPService) ValidateCode(userID uint, code string) (bool, error) {
	var twoFactor models.UserTwoFactor
	if err := database.DB.Where("user_id = ? AND is_enabled = ?", userID, true).First(&twoFactor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("2FA not enabled for this user")
		}
		return false, fmt.Errorf("failed to retrieve 2FA record: %w", err)
	}

	// Check if account is locked
	if twoFactor.LockedUntil != nil {
		lockedUntil, _ := time.Parse(time.RFC3339, *twoFactor.LockedUntil)
		if time.Now().Before(lockedUntil) {
			return false, fmt.Errorf("account temporarily locked due to too many failed attempts")
		}
		// Unlock the account
		database.DB.Model(&twoFactor).Updates(map[string]interface{}{
			"locked_until":    nil,
			"failed_attempts": 0,
		})
	}

	// Decrypt the secret
	secret, err := s.decryptSecret(twoFactor.EncryptedSecret)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt secret: %w", err)
	}

	// Try TOTP validation first
	if totp.Validate(code, secret) {
		// Update last used
		database.DB.Model(&twoFactor).Updates(map[string]interface{}{
			"last_used_at":    time.Now().Format(time.RFC3339),
			"failed_attempts": 0,
		})
		return true, nil
	}

	// Try backup code validation
	if len(code) == 8 {
		valid, remainingCodes := s.validateBackupCode(twoFactor.BackupCodesHash, code)
		if valid {
			// Update backup codes
			remainingCodesJSON, _ := json.Marshal(remainingCodes)
			database.DB.Model(&twoFactor).Updates(map[string]interface{}{
				"backup_codes_hash":      remainingCodesJSON,
				"backup_codes_remaining": len(remainingCodes),
				"last_used_at":           time.Now().Format(time.RFC3339),
				"failed_attempts":        0,
			})
			return true, nil
		}
	}

	// Invalid code - increment failed attempts
	twoFactor.FailedAttempts++
	updates := map[string]interface{}{
		"failed_attempts": twoFactor.FailedAttempts,
	}

	// Lock after 5 failed attempts
	if twoFactor.FailedAttempts >= 5 {
		lockUntil := time.Now().Add(15 * time.Minute).Format(time.RFC3339)
		updates["locked_until"] = lockUntil
	}

	database.DB.Model(&twoFactor).Updates(updates)
	return false, nil
}

// DisableTwoFactor disables 2FA for a user
func (s *TOTPService) DisableTwoFactor(userID uint, code string) error {
	// Validate the code first
	valid, err := s.ValidateCode(userID, code)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("invalid verification code")
	}

	// Delete 2FA record
	if err := database.DB.Where("user_id = ?", userID).Delete(&models.UserTwoFactor{}).Error; err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	// Update user's 2FA enabled flag
	database.DB.Model(&models.User{}).Where("id = ?", userID).Update("two_factor_enabled", false)

	return nil
}

// GetTwoFactorStatus returns 2FA status for a user
func (s *TOTPService) GetTwoFactorStatus(userID uint) (*models.TwoFactorStatusResponse, error) {
	var twoFactor models.UserTwoFactor
	if err := database.DB.Where("user_id = ?", userID).First(&twoFactor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &models.TwoFactorStatusResponse{
				Enabled:              false,
				BackupCodesRemaining: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to retrieve 2FA status: %w", err)
	}

	verifiedAt := ""
	if twoFactor.VerifiedAt != nil {
		verifiedAt = *twoFactor.VerifiedAt
	}

	return &models.TwoFactorStatusResponse{
		Enabled:              twoFactor.IsEnabled,
		BackupCodesRemaining: twoFactor.BackupCodesRemaining,
		VerifiedAt:           verifiedAt,
	}, nil
}

// RegenerateBackupCodes generates new backup codes
func (s *TOTPService) RegenerateBackupCodes(userID uint, code string) ([]string, error) {
	// Validate the code first
	valid, err := s.ValidateCode(userID, code)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("invalid verification code")
	}

	// Generate new backup codes
	backupCodes, hashedCodes, err := s.generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	hashedCodesJSON, _ := json.Marshal(hashedCodes)

	// Update backup codes
	if err := database.DB.Model(&models.UserTwoFactor{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"backup_codes_hash":      hashedCodesJSON,
			"backup_codes_remaining": 10,
			"updated_at":             time.Now().Format(time.RFC3339),
		}).Error; err != nil {
		return nil, fmt.Errorf("failed to update backup codes: %w", err)
	}

	return backupCodes, nil
}

// Is2FAEnabled checks if 2FA is enabled for a user
func (s *TOTPService) Is2FAEnabled(userID uint) (bool, error) {
	var count int64
	if err := database.DB.Model(&models.UserTwoFactor{}).
		Where("user_id = ? AND is_enabled = ?", userID, true).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check 2FA status: %w", err)
	}
	return count > 0, nil
}

// Helper functions

func (s *TOTPService) encryptSecret(secret string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *TOTPService) decryptSecret(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(data) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (s *TOTPService) generateBackupCodes(count int) ([]string, []string, error) {
	codes := make([]string, count)
	hashedCodes := make([]string, count)

	for i := 0; i < count; i++ {
		// Generate 8 random alphanumeric characters
		code := generateRandomCode(8)
		codes[i] = code

		// Hash the code with bcrypt
		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			return nil, nil, err
		}
		hashedCodes[i] = string(hash)
	}

	return codes, hashedCodes, nil
}

func (s *TOTPService) validateBackupCode(hashesJSON json.RawMessage, code string) (bool, []string) {
	var hashes []string
	if err := json.Unmarshal(hashesJSON, &hashes); err != nil {
		return false, nil
	}

	for i, hash := range hashes {
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)); err == nil {
			// Remove the used code
			remaining := append(hashes[:i], hashes[i+1:]...)
			return true, remaining
		}
	}

	return false, hashes
}

func generateRandomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// HashToken creates a SHA-256 hash of a token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// FormatBackupCode formats a backup code for display (adds hyphen in middle)
func FormatBackupCode(code string) string {
	if len(code) != 8 {
		return code
	}
	return code[:4] + "-" + code[4:]
}

// UnformatBackupCode removes formatting from a backup code
func UnformatBackupCode(code string) string {
	return strings.ReplaceAll(code, "-", "")
}
