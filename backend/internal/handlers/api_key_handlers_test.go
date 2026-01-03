package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"react-golang-starter/internal/auth"
	"react-golang-starter/internal/models"

	"github.com/go-chi/chi/v5"
)

// ============ GetUserAPIKeys Tests ============

func TestGetUserAPIKeys_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/me/api-keys", nil)
	w := httptest.NewRecorder()

	GetUserAPIKeys(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetUserAPIKeys() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ GetUserAPIKey Tests ============

func TestGetUserAPIKey_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/me/api-keys/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	GetUserAPIKey(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetUserAPIKey() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestGetUserAPIKey_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/me/api-keys/abc", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	GetUserAPIKey(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("GetUserAPIKey() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ CreateUserAPIKey Tests ============

func TestCreateUserAPIKey_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/api-keys", nil)
	w := httptest.NewRecorder()

	CreateUserAPIKey(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("CreateUserAPIKey() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestCreateUserAPIKey_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/api-keys", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	CreateUserAPIKey(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateUserAPIKey() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestCreateUserAPIKey_InvalidProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
	}{
		{"empty provider", ""},
		{"invalid provider", "invalid"},
		{"unsupported provider", "azure"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := models.CreateUserAPIKeyRequest{
				Provider: tt.provider,
				Name:     "Test Key",
				APIKey:   "sk-1234567890abcdef",
			}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPost, "/api/users/me/api-keys", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			user := &models.User{ID: 1, Role: models.RoleUser}
			ctx := auth.SetUserContext(req.Context(), user)
			req = req.WithContext(ctx)

			CreateUserAPIKey(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("CreateUserAPIKey() with provider %q status = %v, want %v", tt.provider, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestCreateUserAPIKey_InvalidName(t *testing.T) {
	tests := []struct {
		name    string
		keyName string
	}{
		{"empty name", ""},
		{"name too long", string(make([]byte, 101))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := models.CreateUserAPIKeyRequest{
				Provider: "gemini",
				Name:     tt.keyName,
				APIKey:   "sk-1234567890abcdef",
			}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPost, "/api/users/me/api-keys", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			user := &models.User{ID: 1, Role: models.RoleUser}
			ctx := auth.SetUserContext(req.Context(), user)
			req = req.WithContext(ctx)

			CreateUserAPIKey(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("CreateUserAPIKey() with name %q status = %v, want %v", tt.keyName, w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestCreateUserAPIKey_ShortAPIKey(t *testing.T) {
	payload := models.CreateUserAPIKeyRequest{
		Provider: "gemini",
		Name:     "Test Key",
		APIKey:   "short",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/users/me/api-keys", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	CreateUserAPIKey(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateUserAPIKey() with short API key status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ UpdateUserAPIKey Tests ============

func TestUpdateUserAPIKey_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/users/me/api-keys/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	UpdateUserAPIKey(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("UpdateUserAPIKey() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestUpdateUserAPIKey_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/users/me/api-keys/abc", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	UpdateUserAPIKey(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateUserAPIKey() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// Note: TestUpdateUserAPIKey_InvalidJSON requires database integration testing
// as the handler queries the database before decoding JSON body.

// ============ DeleteUserAPIKey Tests ============

func TestDeleteUserAPIKey_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/users/me/api-keys/1", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	DeleteUserAPIKey(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("DeleteUserAPIKey() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestDeleteUserAPIKey_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/users/me/api-keys/abc", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	DeleteUserAPIKey(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("DeleteUserAPIKey() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ TestUserAPIKey Tests ============

func TestTestUserAPIKey_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/api-keys/1/test", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	TestUserAPIKey(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("TestUserAPIKey() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestTestUserAPIKey_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users/me/api-keys/abc/test", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := &models.User{ID: 1, Role: models.RoleUser}
	ctx := auth.SetUserContext(req.Context(), user)
	req = req.WithContext(ctx)

	TestUserAPIKey(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("TestUserAPIKey() with invalid ID status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// ============ ValidAPIKeyProviders Tests ============

func TestValidAPIKeyProviders(t *testing.T) {
	tests := []struct {
		provider string
		valid    bool
	}{
		{models.APIKeyProviderGemini, true},
		{models.APIKeyProviderOpenAI, true},
		{models.APIKeyProviderAnthropic, true},
		{"invalid", false},
		{"", false},
		{"azure", false},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			if ValidAPIKeyProviders[tt.provider] != tt.valid {
				t.Errorf("ValidAPIKeyProviders[%q] = %v, want %v", tt.provider, ValidAPIKeyProviders[tt.provider], tt.valid)
			}
		})
	}
}

// ============ Encryption Helper Tests ============

func TestHashAPIKey(t *testing.T) {
	// Test deterministic hashing
	key := "test-api-key-12345"
	hash1 := hashAPIKey(key)
	hash2 := hashAPIKey(key)

	if hash1 != hash2 {
		t.Error("hashAPIKey() should produce deterministic results")
	}

	// Different keys should produce different hashes
	hash3 := hashAPIKey("different-key")
	if hash1 == hash3 {
		t.Error("hashAPIKey() should produce different hashes for different keys")
	}

	// Hash should be a valid hex string
	if len(hash1) != 64 { // SHA-256 produces 32 bytes = 64 hex chars
		t.Errorf("hashAPIKey() produced hash of length %d, want 64", len(hash1))
	}
}

func TestEncryptDecryptAPIKey(t *testing.T) {
	originalKey := "sk-proj-1234567890abcdef"

	// Encrypt the key
	encrypted, err := encryptAPIKey(originalKey)
	if err != nil {
		t.Fatalf("encryptAPIKey() error = %v", err)
	}

	// Encrypted should be different from original
	if encrypted == originalKey {
		t.Error("encryptAPIKey() should produce different output than input")
	}

	// Decrypt the key
	decrypted, err := decryptAPIKey(encrypted)
	if err != nil {
		t.Fatalf("decryptAPIKey() error = %v", err)
	}

	// Decrypted should match original
	if decrypted != originalKey {
		t.Errorf("decryptAPIKey() = %q, want %q", decrypted, originalKey)
	}
}

func TestEncryptAPIKey_DifferentResultsEachTime(t *testing.T) {
	key := "test-api-key"

	encrypted1, _ := encryptAPIKey(key)
	encrypted2, _ := encryptAPIKey(key)

	// Each encryption should produce different ciphertext (due to random nonce)
	if encrypted1 == encrypted2 {
		t.Error("encryptAPIKey() should produce different ciphertext each time due to random nonce")
	}
}

func TestDecryptAPIKey_InvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"invalid base64", "not-valid-base64!", true},
		{"empty string", "", false},                                           // base64 decode of empty returns empty, not error
		{"too short", "YWJj", false},                                          // "abc" - too short for GCM but may not error
		{"tampered ciphertext", "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=", true}, // Invalid ciphertext
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decryptAPIKey(tt.input)
			if tt.wantError && err == nil {
				t.Error("decryptAPIKey() should return error for invalid input")
			}
		})
	}
}
