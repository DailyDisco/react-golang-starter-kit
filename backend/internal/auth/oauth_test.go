package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"react-golang-starter/internal/models"

	"github.com/go-chi/chi/v5"
)

// ============ GetOAuthURL Tests ============

func TestGetOAuthURL_InvalidProvider(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/invalid", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	GetOAuthURL(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("GetOAuthURL() with invalid provider status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestGetOAuthURL_GoogleNotConfigured(t *testing.T) {
	// Save current config
	oldConfig := googleOAuthConfig
	googleOAuthConfig = nil
	defer func() { googleOAuthConfig = oldConfig }()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/google", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", models.OAuthProviderGoogle)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	GetOAuthURL(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("GetOAuthURL() without Google config status = %v, want %v", w.Code, http.StatusNotImplemented)
	}
}

func TestGetOAuthURL_GitHubNotConfigured(t *testing.T) {
	// Save current config
	oldConfig := githubOAuthConfig
	githubOAuthConfig = nil
	defer func() { githubOAuthConfig = oldConfig }()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/github", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", models.OAuthProviderGitHub)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	GetOAuthURL(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("GetOAuthURL() without GitHub config status = %v, want %v", w.Code, http.StatusNotImplemented)
	}
}

// ============ HandleOAuthCallback Tests ============

func TestHandleOAuthCallback_InvalidProvider(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/invalid/callback?code=test&state=test", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	HandleOAuthCallback(w, req)

	// Should redirect with error
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("HandleOAuthCallback() with invalid provider status = %v, want %v", w.Code, http.StatusTemporaryRedirect)
	}
}

func TestHandleOAuthCallback_NoCode(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/google/callback?state=test", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", models.OAuthProviderGoogle)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	HandleOAuthCallback(w, req)

	// Should redirect with error
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("HandleOAuthCallback() without code status = %v, want %v", w.Code, http.StatusTemporaryRedirect)
	}
}

func TestHandleOAuthCallback_OAuthError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/google/callback?error=access_denied&error_description=User%20cancelled", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", models.OAuthProviderGoogle)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	HandleOAuthCallback(w, req)

	// Should redirect with error
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("HandleOAuthCallback() with OAuth error status = %v, want %v", w.Code, http.StatusTemporaryRedirect)
	}
}

func TestHandleOAuthCallback_InvalidState(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/google/callback?code=test&state=invalid", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", models.OAuthProviderGoogle)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	HandleOAuthCallback(w, req)

	// Should redirect with error about invalid state
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("HandleOAuthCallback() with invalid state status = %v, want %v", w.Code, http.StatusTemporaryRedirect)
	}
}

// ============ State Management Tests ============

func TestGenerateState(t *testing.T) {
	state1, err := generateState()
	if err != nil {
		t.Fatalf("generateState() error = %v", err)
	}

	if state1 == "" {
		t.Error("generateState() returned empty string")
	}

	// States should be unique
	state2, err := generateState()
	if err != nil {
		t.Fatalf("generateState() error = %v", err)
	}

	if state1 == state2 {
		t.Error("generateState() should generate unique states")
	}
}

func TestValidateState(t *testing.T) {
	// Generate a state
	state, _ := generateState()

	// Should be valid
	if !validateState(state) {
		t.Error("validateState() should return true for valid state")
	}

	// Should not be valid second time (one-time use)
	if validateState(state) {
		t.Error("validateState() should return false for already used state")
	}
}

func TestValidateState_InvalidState(t *testing.T) {
	if validateState("invalid-state-that-doesnt-exist") {
		t.Error("validateState() should return false for invalid state")
	}
}

func TestValidateState_ExpiredState(t *testing.T) {
	// Manually add an expired state
	oauthStateMutex.Lock()
	expiredState := "expired-test-state"
	oauthStateStore[expiredState] = time.Now().Add(-1 * time.Hour)
	oauthStateMutex.Unlock()

	if validateState(expiredState) {
		t.Error("validateState() should return false for expired state")
	}
}

// ============ GetLinkedProviders Tests ============

func TestGetLinkedProviders_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/providers", nil)
	w := httptest.NewRecorder()

	GetLinkedProviders(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetLinkedProviders() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ UnlinkProvider Tests ============

func TestUnlinkProvider_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/auth/oauth/google", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", "google")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	UnlinkProvider(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("UnlinkProvider() without auth status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

// ============ Configuration Tests ============

func TestIsOAuthConfigured(t *testing.T) {
	// Save current configs
	oldGoogle := googleOAuthConfig
	oldGitHub := githubOAuthConfig
	defer func() {
		googleOAuthConfig = oldGoogle
		githubOAuthConfig = oldGitHub
	}()

	// No config
	googleOAuthConfig = nil
	githubOAuthConfig = nil
	if IsOAuthConfigured() {
		t.Error("IsOAuthConfigured() should return false when no providers configured")
	}
}

func TestIsGoogleOAuthConfigured(t *testing.T) {
	// Save current config
	oldConfig := googleOAuthConfig
	defer func() { googleOAuthConfig = oldConfig }()

	googleOAuthConfig = nil
	if IsGoogleOAuthConfigured() {
		t.Error("IsGoogleOAuthConfigured() should return false when not configured")
	}
}

func TestIsGitHubOAuthConfigured(t *testing.T) {
	// Save current config
	oldConfig := githubOAuthConfig
	defer func() { githubOAuthConfig = oldConfig }()

	githubOAuthConfig = nil
	if IsGitHubOAuthConfigured() {
		t.Error("IsGitHubOAuthConfigured() should return false when not configured")
	}
}

// ============ Error Variable Tests ============

func TestOAuthErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrOAuthNotConfigured", ErrOAuthNotConfigured},
		{"ErrOAuthInvalidState", ErrOAuthInvalidState},
		{"ErrOAuthProviderFailed", ErrOAuthProviderFailed},
		{"ErrOAuthEmailRequired", ErrOAuthEmailRequired},
		{"ErrOAuthAlreadyLinked", ErrOAuthAlreadyLinked},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s should have an error message", tt.name)
			}
		})
	}
}

// ============ Helper Function Tests ============

func TestGetOAuthRedirectURL(t *testing.T) {
	tests := []struct {
		provider string
		wantEnd  string
	}{
		{"google", "/api/auth/oauth/google/callback"},
		{"github", "/api/auth/oauth/github/callback"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			url := getOAuthRedirectURL(tt.provider)
			if url == "" {
				t.Error("getOAuthRedirectURL() returned empty string")
			}
			// URL should end with the expected callback path
			if len(url) < len(tt.wantEnd) || url[len(url)-len(tt.wantEnd):] != tt.wantEnd {
				t.Errorf("getOAuthRedirectURL() = %v, want to end with %v", url, tt.wantEnd)
			}
		})
	}
}

func TestCleanupExpiredStatesLocked(t *testing.T) {
	// Add some expired states
	oauthStateMutex.Lock()
	oauthStateStore["expired1"] = time.Now().Add(-1 * time.Hour)
	oauthStateStore["expired2"] = time.Now().Add(-2 * time.Hour)
	oauthStateStore["valid"] = time.Now().Add(5 * time.Minute)

	cleanupExpiredStatesLocked()

	// Expired states should be removed
	if _, exists := oauthStateStore["expired1"]; exists {
		t.Error("expired state 'expired1' should have been cleaned up")
	}
	if _, exists := oauthStateStore["expired2"]; exists {
		t.Error("expired state 'expired2' should have been cleaned up")
	}
	// Valid state should remain
	if _, exists := oauthStateStore["valid"]; !exists {
		t.Error("valid state should not have been cleaned up")
	}

	// Clean up test data
	delete(oauthStateStore, "valid")
	oauthStateMutex.Unlock()
}
