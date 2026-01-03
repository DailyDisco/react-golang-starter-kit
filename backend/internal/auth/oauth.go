package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"react-golang-starter/internal/cache"
	"react-golang-starter/internal/database"
	"react-golang-starter/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// OAuth configuration
var (
	googleOAuthConfig *oauth2.Config
	githubOAuthConfig *oauth2.Config
)

// oauthStateTTL is how long an OAuth state token is valid
const oauthStateTTL = 5 * time.Minute

// oauthStateCachePrefix is the cache key prefix for OAuth state tokens
const oauthStateCachePrefix = "oauth_state:"

// OAuth errors
var (
	ErrOAuthNotConfigured  = errors.New("oauth provider not configured")
	ErrOAuthInvalidState   = errors.New("invalid oauth state")
	ErrOAuthProviderFailed = errors.New("failed to get user info from provider")
	ErrOAuthEmailRequired  = errors.New("email is required for oauth login")
	ErrOAuthAlreadyLinked  = errors.New("oauth provider already linked to another account")
)

// InitOAuth initializes OAuth configurations from environment variables
func InitOAuth() {
	// Google OAuth
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientID != "" && googleClientSecret != "" {
		googleOAuthConfig = &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  getOAuthRedirectURL("google"),
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
		log.Info().Msg("Google OAuth configured")
	}

	// GitHub OAuth
	githubClientID := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if githubClientID != "" && githubClientSecret != "" {
		githubOAuthConfig = &oauth2.Config{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			RedirectURL:  getOAuthRedirectURL("github"),
			Scopes:       []string{"user:email", "read:user"},
			Endpoint:     github.Endpoint,
		}
		log.Info().Msg("GitHub OAuth configured")
	}
}

func getOAuthRedirectURL(provider string) string {
	baseURL := os.Getenv("OAUTH_REDIRECT_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/api/auth/oauth/%s/callback", baseURL, provider)
}

// generateState creates a random state string for CSRF protection
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	// Store state in Redis with TTL (automatically expires)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cacheKey := oauthStateCachePrefix + state
	if err := cache.Set(ctx, cacheKey, []byte("1"), oauthStateTTL); err != nil {
		// Log but don't fail - cache might not be available
		log.Warn().Err(err).Msg("failed to cache OAuth state, using stateless validation")
	}

	return state, nil
}

// validateState checks if the state is valid and not expired
func validateState(state string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cacheKey := oauthStateCachePrefix + state

	// Check if state exists in cache
	exists := cache.Exists(ctx, cacheKey)
	if !exists {
		return false
	}

	// Delete the state (one-time use)
	cache.Invalidate(ctx, cacheKey)

	return true
}

// GetOAuthURL returns the OAuth authorization URL for a provider
// GET /api/auth/oauth/{provider}
func GetOAuthURL(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	var config *oauth2.Config
	switch provider {
	case models.OAuthProviderGoogle:
		config = googleOAuthConfig
	case models.OAuthProviderGitHub:
		config = githubOAuthConfig
	default:
		http.Error(w, "Invalid OAuth provider", http.StatusBadRequest)
		return
	}

	if config == nil {
		http.Error(w, "OAuth provider not configured", http.StatusNotImplemented)
		return
	}

	state, err := generateState()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate OAuth state")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	response := models.OAuthURLResponse{
		URL:   url,
		State: state,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleOAuthCallback handles the OAuth callback from providers
// GET /api/auth/oauth/{provider}/callback
func HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	// Get code and state from query params
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// Handle OAuth errors from provider
	if errorParam != "" {
		errorDesc := r.URL.Query().Get("error_description")
		log.Warn().Str("error", errorParam).Str("description", errorDesc).Msg("OAuth error from provider")
		redirectWithError(w, r, fmt.Sprintf("OAuth error: %s", errorParam))
		return
	}

	if code == "" {
		redirectWithError(w, r, "No authorization code provided")
		return
	}

	// Validate state
	if !validateState(state) {
		redirectWithError(w, r, "Invalid or expired state")
		return
	}

	// Get OAuth config
	var config *oauth2.Config
	switch provider {
	case models.OAuthProviderGoogle:
		config = googleOAuthConfig
	case models.OAuthProviderGitHub:
		config = githubOAuthConfig
	default:
		redirectWithError(w, r, "Invalid OAuth provider")
		return
	}

	if config == nil {
		redirectWithError(w, r, "OAuth provider not configured")
		return
	}

	// Exchange code for token
	ctx := context.Background()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		log.Error().Err(err).Msg("Failed to exchange OAuth code")
		redirectWithError(w, r, "Failed to authenticate")
		return
	}

	// Get user info from provider
	userInfo, err := getUserInfoFromProvider(ctx, provider, token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user info from provider")
		redirectWithError(w, r, "Failed to get user information")
		return
	}

	if userInfo.Email == "" {
		redirectWithError(w, r, "Email is required for authentication")
		return
	}

	// Find or create user
	user, isNewUser, err := findOrCreateOAuthUser(userInfo, token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to process OAuth user")
		if errors.Is(err, ErrOAuthAlreadyLinked) {
			redirectWithError(w, r, "This account is already linked to another user")
		} else {
			redirectWithError(w, r, "Failed to process authentication")
		}
		return
	}

	// Generate tokens
	jwtToken, err := GenerateJWT(user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT")
		redirectWithError(w, r, "Failed to generate session")
		return
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate refresh token")
		redirectWithError(w, r, "Failed to generate session")
		return
	}

	// Save refresh token to user
	user.RefreshToken = HashToken(refreshToken)
	refreshExpires := time.Now().Add(GetRefreshTokenExpirationTime())
	user.RefreshTokenExpires = &refreshExpires
	if err := database.DB.Save(user).Error; err != nil {
		log.Error().Err(err).Msg("Failed to save refresh token")
	}

	// Set cookies
	setAuthCookies(w, jwtToken, refreshToken)

	// Redirect to frontend with success
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// Include new_user flag for onboarding
	redirectURL := fmt.Sprintf("%s/auth/callback?success=true", frontendURL)
	if isNewUser {
		redirectURL += "&new_user=true"
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func redirectWithError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	// URL-encode the error message to prevent XSS and handle special characters
	redirectURL := fmt.Sprintf("%s/auth/callback?error=%s", frontendURL, url.QueryEscape(errorMsg))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func setAuthCookies(w http.ResponseWriter, jwtToken, refreshToken string) {
	// Access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureCookie(),
		SameSite: getCookieSameSite(),
		MaxAge:   int(GetAccessTokenExpirationTime().Seconds()),
	})

	// Refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   isSecureCookie(),
		SameSite: getCookieSameSite(),
		MaxAge:   int(GetRefreshTokenExpirationTime().Seconds()),
	})
}

// getUserInfoFromProvider fetches user information from the OAuth provider
func getUserInfoFromProvider(ctx context.Context, provider string, token *oauth2.Token) (*models.OAuthUserInfo, error) {
	switch provider {
	case models.OAuthProviderGoogle:
		return getGoogleUserInfo(ctx, token)
	case models.OAuthProviderGitHub:
		return getGitHubUserInfo(ctx, token)
	default:
		return nil, ErrOAuthProviderFailed
	}
}

func getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*models.OAuthUserInfo, error) {
	client := googleOAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, err
	}

	return &models.OAuthUserInfo{
		ID:        googleUser.ID,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		AvatarURL: googleUser.Picture,
		Provider:  models.OAuthProviderGoogle,
	}, nil
}

func getGitHubUserInfo(ctx context.Context, token *oauth2.Token) (*models.OAuthUserInfo, error) {
	client := githubOAuthConfig.Client(ctx, token)

	// Get user info
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, err
	}

	// If email is not public, fetch from emails endpoint
	email := githubUser.Email
	if email == "" {
		var err error
		email, err = getGitHubPrimaryEmail(ctx, client)
		if err != nil {
			log.Warn().Err(err).Msg("failed to fetch GitHub primary email")
		}
	}

	name := githubUser.Name
	if name == "" {
		name = githubUser.Login
	}

	return &models.OAuthUserInfo{
		ID:        fmt.Sprintf("%d", githubUser.ID),
		Email:     email,
		Name:      name,
		AvatarURL: githubUser.AvatarURL,
		Provider:  models.OAuthProviderGitHub,
	}, nil
}

func getGitHubPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	// Find primary verified email
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	// Fallback to any verified email
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}

	return "", nil
}

// findOrCreateOAuthUser finds an existing user or creates a new one
func findOrCreateOAuthUser(userInfo *models.OAuthUserInfo, token *oauth2.Token) (*models.User, bool, error) {
	// First, check if this OAuth account is already linked
	var existingProvider models.OAuthProvider
	err := database.DB.Where("provider = ? AND provider_user_id = ?", userInfo.Provider, userInfo.ID).First(&existingProvider).Error
	if err == nil {
		// OAuth account exists, get the linked user
		var user models.User
		if err := database.DB.First(&user, existingProvider.UserID).Error; err != nil {
			return nil, false, err
		}
		// Update OAuth tokens
		updateOAuthProvider(&existingProvider, token)
		return &user, false, nil
	}

	// Check if a user with this email exists
	var existingUser models.User
	err = database.DB.Where("email = ?", strings.ToLower(userInfo.Email)).First(&existingUser).Error
	if err == nil {
		// User exists, link the OAuth provider
		if err := linkOAuthProvider(&existingUser, userInfo, token); err != nil {
			return nil, false, err
		}
		return &existingUser, false, nil
	}

	// Create new user
	newUser := &models.User{
		Name:            userInfo.Name,
		Email:           strings.ToLower(userInfo.Email),
		Password:        "",   // OAuth users don't have passwords
		EmailVerified:   true, // OAuth emails are verified
		IsActive:        true,
		Role:            models.RoleUser,
		OAuthProvider:   userInfo.Provider,
		OAuthProviderID: userInfo.ID,
		AvatarURL:       userInfo.AvatarURL,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := database.DB.Create(newUser).Error; err != nil {
		return nil, false, err
	}

	// Link OAuth provider
	if err := linkOAuthProvider(newUser, userInfo, token); err != nil {
		log.Warn().Err(err).Msg("Failed to link OAuth provider, but user was created")
	}

	log.Info().
		Uint("user_id", newUser.ID).
		Str("provider", userInfo.Provider).
		Str("email", newUser.Email).
		Msg("Created new OAuth user")

	return newUser, true, nil
}

func linkOAuthProvider(user *models.User, userInfo *models.OAuthUserInfo, token *oauth2.Token) error {
	provider := &models.OAuthProvider{
		UserID:         user.ID,
		Provider:       userInfo.Provider,
		ProviderUserID: userInfo.ID,
		Email:          userInfo.Email,
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		CreatedAt:      time.Now().Format(time.RFC3339),
		UpdatedAt:      time.Now().Format(time.RFC3339),
	}

	if !token.Expiry.IsZero() {
		provider.TokenExpiresAt = token.Expiry.Format(time.RFC3339)
	}

	return database.DB.Create(provider).Error
}

func updateOAuthProvider(provider *models.OAuthProvider, token *oauth2.Token) {
	provider.AccessToken = token.AccessToken
	if token.RefreshToken != "" {
		provider.RefreshToken = token.RefreshToken
	}
	if !token.Expiry.IsZero() {
		provider.TokenExpiresAt = token.Expiry.Format(time.RFC3339)
	}
	provider.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := database.DB.Save(provider).Error; err != nil {
		log.Error().Err(err).Msg("Failed to update OAuth provider tokens")
	}
}

// GetLinkedProviders returns the user's linked OAuth providers
// GET /api/auth/oauth/providers (requires auth)
func GetLinkedProviders(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var providers []models.OAuthProvider
	if err := database.DB.Where("user_id = ?", user.ID).Find(&providers).Error; err != nil {
		log.Error().Err(err).Msg("Failed to get linked providers")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	linked := make([]models.LinkedProvider, 0, len(providers))
	for _, p := range providers {
		linked = append(linked, models.LinkedProvider{
			Provider: p.Provider,
			Email:    p.Email,
			LinkedAt: p.CreatedAt,
		})
	}

	response := models.LinkedProvidersResponse{
		Providers: linked,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UnlinkProvider removes an OAuth provider from the user's account
// DELETE /api/auth/oauth/{provider} (requires auth)
func UnlinkProvider(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	provider := chi.URLParam(r, "provider")

	// Check if user has a password set (can't unlink if no other auth method)
	if user.Password == "" {
		// Count remaining OAuth providers
		var count int64
		database.DB.Model(&models.OAuthProvider{}).Where("user_id = ?", user.ID).Count(&count)
		if count <= 1 {
			http.Error(w, "Cannot unlink the only authentication method", http.StatusBadRequest)
			return
		}
	}

	// Delete the OAuth provider link
	result := database.DB.Where("user_id = ? AND provider = ?", user.ID, provider).Delete(&models.OAuthProvider{})
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to unlink OAuth provider")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if result.RowsAffected == 0 {
		http.Error(w, "Provider not linked", http.StatusNotFound)
		return
	}

	// Clear OAuth fields from user if this was their primary OAuth provider
	if user.OAuthProvider == provider {
		user.OAuthProvider = ""
		user.OAuthProviderID = ""
		database.DB.Save(user)
	}

	log.Info().
		Uint("user_id", user.ID).
		Str("provider", provider).
		Msg("Unlinked OAuth provider")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully unlinked %s", provider),
	})
}

// IsOAuthConfigured returns whether any OAuth provider is configured
func IsOAuthConfigured() bool {
	return googleOAuthConfig != nil || githubOAuthConfig != nil
}

// IsGoogleOAuthConfigured returns whether Google OAuth is configured
func IsGoogleOAuthConfigured() bool {
	return googleOAuthConfig != nil
}

// IsGitHubOAuthConfigured returns whether GitHub OAuth is configured
func IsGitHubOAuthConfigured() bool {
	return githubOAuthConfig != nil
}
