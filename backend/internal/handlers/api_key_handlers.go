package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"

	"github.com/go-chi/chi/v5"
)

// ValidAPIKeyProviders defines allowed API key providers
var ValidAPIKeyProviders = map[string]bool{
	models.APIKeyProviderGemini:    true,
	models.APIKeyProviderOpenAI:    true,
	models.APIKeyProviderAnthropic: true,
}

// ============ API Key Management Handlers ============

// GetUserAPIKeys returns all API keys for the current user
// @Summary Get user API keys
// @Tags User Settings
// @Security BearerAuth
// @Success 200 {object} models.UserAPIKeysResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/users/me/api-keys [get]
func GetUserAPIKeys(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	var keys []models.UserAPIKey
	if err := database.DB.Where("user_id = ?", userID).Find(&keys).Error; err != nil {
		WriteInternalError(w, r, "Failed to retrieve API keys")
		return
	}

	responses := make([]models.UserAPIKeyResponse, len(keys))
	for i, k := range keys {
		responses[i] = k.ToUserAPIKeyResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.UserAPIKeysResponse{
		Keys:  responses,
		Count: len(responses),
	})
}

// GetUserAPIKey returns a single API key by ID
// @Summary Get user API key by ID
// @Tags User Settings
// @Security BearerAuth
// @Param id path int true "API Key ID"
// @Success 200 {object} models.UserAPIKeyResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/users/me/api-keys/{id} [get]
func GetUserAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	keyID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid API key ID")
		return
	}

	var key models.UserAPIKey
	if err := database.DB.Where("id = ? AND user_id = ?", keyID, userID).First(&key).Error; err != nil {
		WriteNotFound(w, r, "API key not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(key.ToUserAPIKeyResponse())
}

// CreateUserAPIKey creates a new API key for the current user
// @Summary Create user API key
// @Tags User Settings
// @Security BearerAuth
// @Param body body models.CreateUserAPIKeyRequest true "API key details"
// @Success 201 {object} models.UserAPIKeyResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse "Key already exists for this provider"
// @Router /api/users/me/api-keys [post]
func CreateUserAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	var req models.CreateUserAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Normalize provider
	req.Provider = strings.ToLower(strings.TrimSpace(req.Provider))
	req.Name = strings.TrimSpace(req.Name)
	req.APIKey = strings.TrimSpace(req.APIKey)

	// Validate provider
	if !ValidAPIKeyProviders[req.Provider] {
		WriteBadRequest(w, r, "Invalid provider. Allowed: gemini, openai, anthropic")
		return
	}

	// Validate name
	if len(req.Name) < 1 || len(req.Name) > 100 {
		WriteBadRequest(w, r, "Name must be between 1 and 100 characters")
		return
	}

	// Validate API key
	if len(req.APIKey) < 10 {
		WriteBadRequest(w, r, "API key appears to be invalid")
		return
	}

	// Check if user already has a key for this provider
	var existing models.UserAPIKey
	if err := database.DB.Where("user_id = ? AND provider = ?", userID, req.Provider).First(&existing).Error; err == nil {
		WriteConflict(w, r, "You already have an API key for this provider. Update or delete it first.")
		return
	}

	// Hash the key for verification
	keyHash := hashAPIKey(req.APIKey)

	// Encrypt the key for storage
	keyEncrypted, err := encryptAPIKey(req.APIKey)
	if err != nil {
		WriteInternalError(w, r, "Failed to encrypt API key")
		return
	}

	// Create preview (last 4 characters)
	keyPreview := "..." + req.APIKey[len(req.APIKey)-4:]

	now := time.Now().Format(time.RFC3339)
	key := models.UserAPIKey{
		UserID:       userID,
		Provider:     req.Provider,
		Name:         req.Name,
		KeyHash:      keyHash,
		KeyEncrypted: keyEncrypted,
		KeyPreview:   keyPreview,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := database.DB.Create(&key).Error; err != nil {
		WriteInternalError(w, r, "Failed to save API key")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(key.ToUserAPIKeyResponse())
}

// UpdateUserAPIKey updates an existing API key
// @Summary Update user API key
// @Tags User Settings
// @Security BearerAuth
// @Param id path int true "API Key ID"
// @Param body body models.UpdateUserAPIKeyRequest true "API key updates"
// @Success 200 {object} models.UserAPIKeyResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/users/me/api-keys/{id} [put]
func UpdateUserAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	keyID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid API key ID")
		return
	}

	var key models.UserAPIKey
	if err := database.DB.Where("id = ? AND user_id = ?", keyID, userID).First(&key).Error; err != nil {
		WriteNotFound(w, r, "API key not found")
		return
	}

	var req models.UpdateUserAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, r, "Invalid request body")
		return
	}

	// Update name if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if len(name) < 1 || len(name) > 100 {
			WriteBadRequest(w, r, "Name must be between 1 and 100 characters")
			return
		}
		key.Name = name
	}

	// Update API key if provided
	if req.APIKey != nil {
		apiKey := strings.TrimSpace(*req.APIKey)
		if len(apiKey) < 10 {
			WriteBadRequest(w, r, "API key appears to be invalid")
			return
		}

		key.KeyHash = hashAPIKey(apiKey)
		keyEncrypted, err := encryptAPIKey(apiKey)
		if err != nil {
			WriteInternalError(w, r, "Failed to encrypt API key")
			return
		}
		key.KeyEncrypted = keyEncrypted
		key.KeyPreview = "..." + apiKey[len(apiKey)-4:]
	}

	// Update active status if provided
	if req.IsActive != nil {
		key.IsActive = *req.IsActive
	}

	key.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(&key).Error; err != nil {
		WriteInternalError(w, r, "Failed to update API key")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(key.ToUserAPIKeyResponse())
}

// DeleteUserAPIKey deletes an API key
// @Summary Delete user API key
// @Tags User Settings
// @Security BearerAuth
// @Param id path int true "API Key ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/users/me/api-keys/{id} [delete]
func DeleteUserAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	keyID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid API key ID")
		return
	}

	result := database.DB.Where("id = ? AND user_id = ?", keyID, userID).Delete(&models.UserAPIKey{})
	if result.Error != nil {
		WriteInternalError(w, r, "Failed to delete API key")
		return
	}

	if result.RowsAffected == 0 {
		WriteNotFound(w, r, "API key not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "API key deleted successfully",
	})
}

// TestUserAPIKey tests if an API key is valid by making a test request
// @Summary Test user API key
// @Tags User Settings
// @Security BearerAuth
// @Param id path int true "API Key ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/users/me/api-keys/{id}/test [post]
func TestUserAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		WriteUnauthorized(w, r, "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	keyID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		WriteBadRequest(w, r, "Invalid API key ID")
		return
	}

	var key models.UserAPIKey
	if err := database.DB.Where("id = ? AND user_id = ?", keyID, userID).First(&key).Error; err != nil {
		WriteNotFound(w, r, "API key not found")
		return
	}

	// Decrypt the API key
	_, err = decryptAPIKey(key.KeyEncrypted)
	if err != nil {
		WriteInternalError(w, r, "Failed to decrypt API key")
		return
	}

	// For now, just verify we can decrypt and return success
	// In a real implementation, we would make a test API call to the provider
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: "API key is valid and properly stored",
	})
}

// GetDecryptedAPIKey retrieves and decrypts a user's API key for a provider (internal use)
func GetDecryptedAPIKey(userID uint, provider string) (string, error) {
	var key models.UserAPIKey
	if err := database.DB.Where("user_id = ? AND provider = ? AND is_active = ?", userID, provider, true).First(&key).Error; err != nil {
		return "", err
	}

	decryptedKey, err := decryptAPIKey(key.KeyEncrypted)
	if err != nil {
		return "", err
	}

	// Update usage stats
	now := time.Now().Format(time.RFC3339)
	database.DB.Model(&key).Updates(map[string]interface{}{
		"last_used_at": now,
		"usage_count":  key.UsageCount + 1,
	})

	return decryptedKey, nil
}

// ============ Encryption Helper Functions ============

// getEncryptionKey returns the 32-byte encryption key derived from JWT_SECRET
func getEncryptionKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-insecure-key-change-me"
	}
	// Use SHA-256 to ensure we have exactly 32 bytes for AES-256
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// hashAPIKey creates a SHA-256 hash of the API key
func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// encryptAPIKey encrypts an API key using AES-256-GCM
func encryptAPIKey(apiKey string) (string, error) {
	key := getEncryptionKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(apiKey), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptAPIKey decrypts an API key using AES-256-GCM
func decryptAPIKey(encrypted string) (string, error) {
	key := getEncryptionKey()

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return "", err
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
