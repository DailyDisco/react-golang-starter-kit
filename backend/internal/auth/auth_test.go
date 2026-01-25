package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"react-golang-starter/internal/models"
)

// ============ Password Tests ============

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "SecurePass123", false},
		{"empty password", "", false}, // bcrypt accepts empty
		{"long password", string(make([]byte, 72)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	password := "TestPassword123"
	hash, _ := HashPassword(password)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{"correct password", password, hash, true},
		{"wrong password", "WrongPassword", hash, false},
		{"empty password", "", hash, false},
		{"invalid hash", password, "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPassword(tt.password, tt.hash); got != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============ JWT Tests ============

func TestGenerateJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")
	defer os.Unsetenv("JWT_SECRET")

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "user",
	}

	token, err := GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateJWT() returned empty token")
	}
}

func TestGenerateJWT_NoSecret(t *testing.T) {
	// Save and restore JWT_SECRET to avoid affecting other tests
	originalSecret := os.Getenv("JWT_SECRET")
	os.Unsetenv("JWT_SECRET")
	defer func() {
		if originalSecret != "" {
			os.Setenv("JWT_SECRET", originalSecret)
		}
	}()

	user := &models.User{ID: 1, Email: "test@example.com"}
	_, err := GenerateJWT(user)

	if err == nil {
		t.Error("GenerateJWT() should error without JWT_SECRET")
	}
}

func TestGenerateJWT_UsesConfiguredExpiration(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")
	os.Setenv("JWT_EXPIRATION_HOURS", "1")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_EXPIRATION_HOURS")
	}()

	user := &models.User{ID: 1, Email: "test@example.com", Role: "user"}
	token, err := GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}

	// Token should expire in ~1 hour, not 24 hours
	expTime := claims.ExpiresAt.Time
	expected := time.Now().Add(1 * time.Hour)

	// Allow 1 minute tolerance
	diff := expTime.Sub(expected)
	if diff > time.Minute || diff < -time.Minute {
		t.Errorf("Token expires at %v, expected ~%v (diff: %v)", expTime, expected, diff)
	}
}

func TestValidateJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")
	defer os.Unsetenv("JWT_SECRET")

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "admin",
	}

	token, _ := GenerateJWT(user)

	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("claims.UserID = %v, want %v", claims.UserID, user.ID)
	}
	if claims.Email != user.Email {
		t.Errorf("claims.Email = %v, want %v", claims.Email, user.Email)
	}
	if claims.Role != user.Role {
		t.Errorf("claims.Role = %v, want %v", claims.Role, user.Role)
	}
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")
	defer os.Unsetenv("JWT_SECRET")

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"malformed token", "not.a.token"},
		{"random string", "abcdefghijklmnop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateJWT(tt.token)
			if err == nil {
				t.Error("ValidateJWT() should error for invalid token")
			}
		})
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	// Save and restore original JWT_SECRET
	originalSecret := os.Getenv("JWT_SECRET")
	defer func() {
		if originalSecret != "" {
			os.Setenv("JWT_SECRET", originalSecret)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
	}()

	// Generate token with one secret
	os.Setenv("JWT_SECRET", "secret-one")
	user := &models.User{ID: 1, Email: "test@example.com", Role: "user"}
	token, _ := GenerateJWT(user)

	// Try to validate with different secret
	os.Setenv("JWT_SECRET", "secret-two")

	_, err := ValidateJWT(token)
	if err == nil {
		t.Error("ValidateJWT() should error with wrong secret")
	}
}

// ============ Password Validation Tests ============

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "SecurePass123", false},
		{"too short", "Short1", true},
		{"no uppercase", "securepass123", true},
		{"no lowercase", "SECUREPASS123", true},
		{"no digit", "SecurePassword", true},
		{"exactly 8 chars valid", "Secure12", false},
		{"only lowercase and digit", "password1", true},
		{"only uppercase and digit", "PASSWORD1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

// ============ Email Validation Tests ============

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		// Valid emails
		{"valid email", "user@example.com", false},
		{"valid with subdomain", "user@mail.example.com", false},
		{"valid with plus", "user+tag@example.com", false},
		{"valid with dots in local", "user.name@example.com", false},
		{"valid with numbers", "user123@example.com", false},
		{"valid short TLD", "user@example.co", false},

		// Invalid emails - length
		{"too short", "a@b.c", true},
		{"way too short", "a@b", true},

		// Invalid emails - format
		{"no at sign", "userexample.com", true},
		{"no domain", "user@", true},
		{"no local part", "@example.com", true},
		{"no TLD", "user@example", true},
		{"double at", "user@@example.com", true},
		{"spaces", "user @example.com", true},

		// Invalid emails - dots/hyphens at wrong positions
		{"leading dot in local", ".user@example.com", true},
		{"trailing dot in local", "user.@example.com", true},
		{"leading dot in domain", "user@.example.com", true},
		{"trailing dot in domain", "user@example.com.", true},
		{"leading hyphen in domain", "user@-example.com", true},
		{"trailing hyphen in domain", "user@example-.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

// ============ Token Helper Tests ============

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		wantToken  string
		wantErr    bool
	}{
		{"valid bearer", "Bearer abc123", "abc123", false},
		{"valid bearer with long token", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", false},
		{"empty header", "", "", true},
		{"no bearer prefix", "abc123", "", true},
		{"wrong prefix", "Basic abc123", "", true},
		{"bearer only no token", "Bearer", "", true},
		{"lowercase bearer", "bearer abc123", "", true},
		{"extra spaces", "Bearer  abc123", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractTokenFromHeader(tt.authHeader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTokenFromHeader(%q) error = %v, wantErr %v", tt.authHeader, err, tt.wantErr)
			}
			if token != tt.wantToken {
				t.Errorf("ExtractTokenFromHeader(%q) = %v, want %v", tt.authHeader, token, tt.wantToken)
			}
		})
	}
}

func TestGetTokenExpirationTime(t *testing.T) {
	// GetTokenExpirationTime is now a wrapper for GetAccessTokenExpirationTime
	// and defaults to 15 minutes for short-lived access tokens

	// Clear all related env vars
	os.Unsetenv("ACCESS_TOKEN_EXPIRATION_MINUTES")
	os.Unsetenv("JWT_EXPIRATION_HOURS")

	// Test default (15 minutes for short-lived access tokens)
	if got := GetTokenExpirationTime(); got != 15*time.Minute {
		t.Errorf("GetTokenExpirationTime() default = %v, want 15m", got)
	}

	// Test custom value via ACCESS_TOKEN_EXPIRATION_MINUTES
	os.Setenv("ACCESS_TOKEN_EXPIRATION_MINUTES", "30")
	if got := GetTokenExpirationTime(); got != 30*time.Minute {
		t.Errorf("GetTokenExpirationTime() custom = %v, want 30m", got)
	}

	// Test invalid value (should fallback to default 15m)
	os.Setenv("ACCESS_TOKEN_EXPIRATION_MINUTES", "invalid")
	if got := GetTokenExpirationTime(); got != 15*time.Minute {
		t.Errorf("GetTokenExpirationTime() invalid = %v, want 15m fallback", got)
	}

	// Test negative value (should fallback to default 15m)
	os.Setenv("ACCESS_TOKEN_EXPIRATION_MINUTES", "-5")
	if got := GetTokenExpirationTime(); got != 15*time.Minute {
		t.Errorf("GetTokenExpirationTime() negative = %v, want 15m fallback", got)
	}

	// Test zero (should fallback to default 15m)
	os.Setenv("ACCESS_TOKEN_EXPIRATION_MINUTES", "0")
	if got := GetTokenExpirationTime(); got != 15*time.Minute {
		t.Errorf("GetTokenExpirationTime() zero = %v, want 15m fallback", got)
	}

	os.Unsetenv("ACCESS_TOKEN_EXPIRATION_MINUTES")
	os.Unsetenv("JWT_EXPIRATION_HOURS")
}

// ============ Verification Token Tests ============

func TestGenerateVerificationToken(t *testing.T) {
	token1, err := GenerateVerificationToken()
	if err != nil {
		t.Fatalf("GenerateVerificationToken() error = %v", err)
	}

	if len(token1) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("GenerateVerificationToken() length = %d, want 64", len(token1))
	}

	// Verify tokens are unique
	token2, _ := GenerateVerificationToken()
	if token1 == token2 {
		t.Error("GenerateVerificationToken() should generate unique tokens")
	}
}

// ============ Cookie Configuration Tests ============

func TestIsSecureCookie(t *testing.T) {
	tests := []struct {
		name  string
		goEnv string
		want  bool
	}{
		{"production should be secure", "production", true},
		{"prod should be secure", "prod", true},
		{"development should not be secure", "development", false},
		{"empty should not be secure", "", false},
		{"test should not be secure", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.goEnv != "" {
				t.Setenv("GO_ENV", tt.goEnv)
			} else {
				os.Unsetenv("GO_ENV")
			}

			got := isSecureCookie()
			if got != tt.want {
				t.Errorf("isSecureCookie() with GO_ENV=%q = %v, want %v", tt.goEnv, got, tt.want)
			}
		})
	}
}

func TestGetCookieSameSite(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   http.SameSite
	}{
		{"strict should return Strict", "strict", http.SameSiteStrictMode},
		{"STRICT should return Strict", "STRICT", http.SameSiteStrictMode},
		{"Strict should return Strict", "Strict", http.SameSiteStrictMode},
		{"lax should return Lax", "lax", http.SameSiteLaxMode},
		{"empty should return Lax", "", http.SameSiteLaxMode},
		{"invalid should return Lax", "invalid", http.SameSiteLaxMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				t.Setenv("COOKIE_SAMESITE", tt.envVal)
			} else {
				os.Unsetenv("COOKIE_SAMESITE")
			}

			got := getCookieSameSite()
			if got != tt.want {
				t.Errorf("getCookieSameSite() with COOKIE_SAMESITE=%q = %v, want %v", tt.envVal, got, tt.want)
			}
		})
	}
}

// ============ Cookie Setting Tests ============

func TestSetAuthCookie(t *testing.T) {
	t.Setenv("GO_ENV", "test")
	os.Unsetenv("COOKIE_SAMESITE")

	rr := httptest.NewRecorder()
	SetAuthCookie(rr, "test-token-value")

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != AuthCookieName {
		t.Errorf("cookie name = %q, want %q", cookie.Name, AuthCookieName)
	}
	if cookie.Value != "test-token-value" {
		t.Errorf("cookie value = %q, want 'test-token-value'", cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}
	if cookie.Path != "/" {
		t.Errorf("cookie path = %q, want '/'", cookie.Path)
	}
	if cookie.MaxAge <= 0 {
		t.Error("cookie should have positive MaxAge")
	}
}

func TestClearAuthCookie(t *testing.T) {
	t.Setenv("GO_ENV", "test")
	os.Unsetenv("COOKIE_SAMESITE")

	rr := httptest.NewRecorder()
	ClearAuthCookie(rr)

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != AuthCookieName {
		t.Errorf("cookie name = %q, want %q", cookie.Name, AuthCookieName)
	}
	if cookie.Value != "" {
		t.Errorf("cookie value should be empty, got %q", cookie.Value)
	}
	if cookie.MaxAge != -1 {
		t.Errorf("cookie MaxAge = %d, want -1 (expire immediately)", cookie.MaxAge)
	}
}

func TestSetRefreshCookie(t *testing.T) {
	t.Setenv("GO_ENV", "test")
	os.Unsetenv("COOKIE_SAMESITE")

	rr := httptest.NewRecorder()
	SetRefreshCookie(rr, "refresh-token-value")

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != RefreshCookieName {
		t.Errorf("cookie name = %q, want %q", cookie.Name, RefreshCookieName)
	}
	if cookie.Value != "refresh-token-value" {
		t.Errorf("cookie value = %q, want 'refresh-token-value'", cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}
	if cookie.Path != "/api/v1/auth" {
		t.Errorf("cookie path = %q, want '/api/v1/auth'", cookie.Path)
	}
}

func TestClearRefreshCookie(t *testing.T) {
	t.Setenv("GO_ENV", "test")
	os.Unsetenv("COOKIE_SAMESITE")

	rr := httptest.NewRecorder()
	ClearRefreshCookie(rr)

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != RefreshCookieName {
		t.Errorf("cookie name = %q, want %q", cookie.Name, RefreshCookieName)
	}
	if cookie.Value != "" {
		t.Errorf("cookie value should be empty, got %q", cookie.Value)
	}
	if cookie.MaxAge != -1 {
		t.Errorf("cookie MaxAge = %d, want -1 (expire immediately)", cookie.MaxAge)
	}
}

// ============ Cookie Extraction Tests ============

func TestExtractTokenFromCookie(t *testing.T) {
	tests := []struct {
		name      string
		setupReq  func() *http.Request
		wantToken string
		wantErr   bool
	}{
		{
			"valid cookie",
			func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: "my-token"})
				return req
			},
			"my-token",
			false,
		},
		{
			"no cookie",
			func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/test", nil)
			},
			"",
			true,
		},
		{
			"empty cookie value",
			func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.AddCookie(&http.Cookie{Name: AuthCookieName, Value: ""})
				return req
			},
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			token, err := ExtractTokenFromCookie(req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTokenFromCookie() error = %v, wantErr %v", err, tt.wantErr)
			}
			if token != tt.wantToken {
				t.Errorf("ExtractTokenFromCookie() = %q, want %q", token, tt.wantToken)
			}
		})
	}
}

func TestExtractRefreshTokenFromCookie(t *testing.T) {
	tests := []struct {
		name      string
		setupReq  func() *http.Request
		wantToken string
		wantErr   bool
	}{
		{
			"valid cookie",
			func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: "refresh-token"})
				return req
			},
			"refresh-token",
			false,
		},
		{
			"no cookie",
			func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/test", nil)
			},
			"",
			true,
		},
		{
			"empty cookie value",
			func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: ""})
				return req
			},
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			token, err := ExtractRefreshTokenFromCookie(req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractRefreshTokenFromCookie() error = %v, wantErr %v", err, tt.wantErr)
			}
			if token != tt.wantToken {
				t.Errorf("ExtractRefreshTokenFromCookie() = %q, want %q", token, tt.wantToken)
			}
		})
	}
}

// ============ Impersonation Token Tests ============

func TestGenerateImpersonationToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-key-for-testing")

	// targetUser is the user being impersonated
	targetUser := &models.User{
		ID:    42,
		Email: "target@example.com",
		Role:  models.RoleUser,
	}
	// originalUserID is the admin doing the impersonation
	originalUserID := uint(1)

	token, err := GenerateImpersonationToken(targetUser, originalUserID)
	if err != nil {
		t.Fatalf("GenerateImpersonationToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateImpersonationToken() returned empty token")
	}

	// Validate the token
	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}

	// UserID should be the target user's ID (the one being impersonated)
	if claims.UserID != targetUser.ID {
		t.Errorf("claims.UserID = %d, want %d", claims.UserID, targetUser.ID)
	}
	// OriginalUserID should be set to the impersonator's ID
	if claims.OriginalUserID != originalUserID {
		t.Errorf("claims.OriginalUserID = %d, want %d", claims.OriginalUserID, originalUserID)
	}
}

func TestGenerateImpersonationToken_NoSecret(t *testing.T) {
	originalSecret := os.Getenv("JWT_SECRET")
	os.Unsetenv("JWT_SECRET")
	defer func() {
		if originalSecret != "" {
			os.Setenv("JWT_SECRET", originalSecret)
		}
	}()

	user := &models.User{ID: 1, Email: "admin@example.com", Role: models.RoleAdmin}
	_, err := GenerateImpersonationToken(user, 42)

	if err == nil {
		t.Error("GenerateImpersonationToken() should error without JWT_SECRET")
	}
}

// ============ Access Token Expiration Tests ============

func TestGetAccessTokenExpirationTime(t *testing.T) {
	// Clear related env vars
	os.Unsetenv("ACCESS_TOKEN_EXPIRATION_MINUTES")
	os.Unsetenv("JWT_EXPIRATION_HOURS")

	// Test default (15 minutes)
	if got := GetAccessTokenExpirationTime(); got != 15*time.Minute {
		t.Errorf("GetAccessTokenExpirationTime() default = %v, want 15m", got)
	}

	// Test custom value
	t.Setenv("ACCESS_TOKEN_EXPIRATION_MINUTES", "30")
	if got := GetAccessTokenExpirationTime(); got != 30*time.Minute {
		t.Errorf("GetAccessTokenExpirationTime() custom = %v, want 30m", got)
	}

	// Test legacy JWT_EXPIRATION_HOURS fallback
	os.Unsetenv("ACCESS_TOKEN_EXPIRATION_MINUTES")
	t.Setenv("JWT_EXPIRATION_HOURS", "2")
	if got := GetAccessTokenExpirationTime(); got != 2*time.Hour {
		t.Errorf("GetAccessTokenExpirationTime() legacy = %v, want 2h", got)
	}

	os.Unsetenv("ACCESS_TOKEN_EXPIRATION_MINUTES")
	os.Unsetenv("JWT_EXPIRATION_HOURS")
}

// ============ Refresh Token Expiration Tests ============

func TestGetRefreshTokenExpirationTime(t *testing.T) {
	// Clear related env vars
	os.Unsetenv("REFRESH_TOKEN_EXPIRATION_DAYS")

	// Test default (7 days)
	if got := GetRefreshTokenExpirationTime(); got != 7*24*time.Hour {
		t.Errorf("GetRefreshTokenExpirationTime() default = %v, want 7 days", got)
	}

	// Test custom value
	t.Setenv("REFRESH_TOKEN_EXPIRATION_DAYS", "30")
	if got := GetRefreshTokenExpirationTime(); got != 30*24*time.Hour {
		t.Errorf("GetRefreshTokenExpirationTime() custom = %v, want 30 days", got)
	}

	// Test invalid value (should fallback to default)
	t.Setenv("REFRESH_TOKEN_EXPIRATION_DAYS", "invalid")
	if got := GetRefreshTokenExpirationTime(); got != 7*24*time.Hour {
		t.Errorf("GetRefreshTokenExpirationTime() invalid = %v, want 7 days fallback", got)
	}

	// Test negative value (should fallback to default)
	t.Setenv("REFRESH_TOKEN_EXPIRATION_DAYS", "-5")
	if got := GetRefreshTokenExpirationTime(); got != 7*24*time.Hour {
		t.Errorf("GetRefreshTokenExpirationTime() negative = %v, want 7 days fallback", got)
	}

	os.Unsetenv("REFRESH_TOKEN_EXPIRATION_DAYS")
}
