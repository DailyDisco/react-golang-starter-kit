package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"react-golang-starter/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Cookie configuration constants
const (
	AuthCookieName    = "auth_token"
	RefreshCookieName = "refresh_token"
)

// emailRegex is a simplified RFC 5322 compliant email validation pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Claims represents the JWT claims structure
type Claims struct {
	UserID         uint   `json:"user_id"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	OriginalUserID uint   `json:"original_user_id,omitempty"` // Set when impersonating
	jwt.RegisteredClaims
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword compares a plain password with a hashed password
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateJWT generates a JWT access token for the given user
// Access tokens are short-lived (default 15 minutes) for security
func GenerateJWT(user *models.User) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET environment variable is not set")
	}

	// Set token expiration time from config (default: 15 minutes for access tokens)
	expirationTime := time.Now().Add(GetAccessTokenExpirationTime())

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateToken generates a regular JWT token (alias for GenerateJWT)
func GenerateToken(user *models.User) (string, error) {
	return GenerateJWT(user)
}

// GenerateImpersonationToken generates a JWT token for impersonation
// The token includes the original admin's user ID for tracking
func GenerateImpersonationToken(targetUser *models.User, originalUserID uint) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET environment variable is not set")
	}

	// Impersonation tokens have shorter expiration (1 hour)
	expirationTime := time.Now().Add(1 * time.Hour)

	claims := &Claims{
		UserID:         targetUser.ID,
		Email:          targetUser.Email,
		Role:           targetUser.Role,
		OriginalUserID: originalUserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*Claims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateVerificationToken generates a random verification token
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRefreshToken generates a cryptographically secure refresh token
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashToken creates a SHA-256 hash of a token for secure storage
// Used for blacklisting tokens without storing the actual token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GetRefreshTokenExpirationTime returns the configured refresh token expiration time
// Default: 7 days
func GetRefreshTokenExpirationTime() time.Duration {
	envDuration := os.Getenv("REFRESH_TOKEN_EXPIRATION_DAYS")
	if envDuration == "" {
		return 7 * 24 * time.Hour // default 7 days
	}

	days, err := strconv.Atoi(envDuration)
	if err != nil || days <= 0 {
		return 7 * 24 * time.Hour // fallback to default
	}

	return time.Duration(days) * 24 * time.Hour
}

// GetAccessTokenExpirationTime returns the configured access token expiration time
// Default: 15 minutes (short-lived for security)
func GetAccessTokenExpirationTime() time.Duration {
	envDuration := os.Getenv("ACCESS_TOKEN_EXPIRATION_MINUTES")
	if envDuration == "" {
		// Check legacy JWT_EXPIRATION_HOURS for backwards compatibility
		legacyDuration := os.Getenv("JWT_EXPIRATION_HOURS")
		if legacyDuration != "" {
			hours, err := strconv.Atoi(legacyDuration)
			if err == nil && hours > 0 {
				return time.Duration(hours) * time.Hour
			}
		}
		return 15 * time.Minute // default 15 minutes
	}

	minutes, err := strconv.Atoi(envDuration)
	if err != nil || minutes <= 0 {
		return 15 * time.Minute // fallback to default
	}

	return time.Duration(minutes) * time.Minute
}

// ExtractTokenFromHeader extracts the JWT token from the Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("authorization header format must be Bearer {token}")
	}

	return parts[1], nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Check for at least one uppercase letter
	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	hasDigit := strings.ContainsAny(password, "0123456789")
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	return nil
}

// ValidateEmail validates email format using RFC 5322 compliant regex
func ValidateEmail(email string) error {
	if len(email) < 5 {
		return errors.New("email must be at least 5 characters")
	}

	if len(email) > 254 {
		return errors.New("email must not exceed 254 characters")
	}

	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	// Additional edge case validation
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.New("invalid email format")
	}

	local, domain := parts[0], parts[1]

	// Local part cannot start or end with a dot
	if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") {
		return errors.New("invalid email format")
	}

	// Domain cannot start or end with a dot or hyphen
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") ||
		strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return errors.New("invalid email format")
	}

	// Check each domain segment for leading/trailing hyphens
	domainParts := strings.Split(domain, ".")
	for _, part := range domainParts {
		if strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return errors.New("invalid email format")
		}
	}

	return nil
}

// GetTokenExpirationTime returns the configured token expiration time
// Deprecated: Use GetAccessTokenExpirationTime instead. This function is kept for backwards compatibility.
func GetTokenExpirationTime() time.Duration {
	return GetAccessTokenExpirationTime()
}

// isSecureCookie returns true if cookies should use the Secure flag
func isSecureCookie() bool {
	env := os.Getenv("GO_ENV")
	return env == "production" || env == "prod"
}

// getCookieSameSite returns the SameSite mode based on configuration
// Defaults to Lax, but can be set to Strict via COOKIE_SAMESITE=strict
func getCookieSameSite() http.SameSite {
	mode := strings.ToLower(os.Getenv("COOKIE_SAMESITE"))
	if mode == "strict" {
		return http.SameSiteStrictMode
	}
	return http.SameSiteLaxMode
}

// SetAuthCookie sets the JWT token as an httpOnly cookie
func SetAuthCookie(w http.ResponseWriter, token string) {
	expiration := GetTokenExpirationTime()
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(expiration.Seconds()),
		HttpOnly: true,
		Secure:   isSecureCookie(),
		SameSite: getCookieSameSite(),
	})
}

// ClearAuthCookie clears the auth cookie (for logout)
func ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureCookie(),
		SameSite: getCookieSameSite(),
	})
}

// SetRefreshCookie sets the refresh token as an httpOnly cookie
func SetRefreshCookie(w http.ResponseWriter, token string) {
	expiration := GetRefreshTokenExpirationTime()
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    token,
		Path:     "/api/v1/auth", // Only sent to auth endpoints (matches /api/v1/auth/refresh)
		MaxAge:   int(expiration.Seconds()),
		HttpOnly: true,
		Secure:   isSecureCookie(),
		SameSite: getCookieSameSite(),
	})
}

// ClearRefreshCookie clears the refresh cookie (for logout)
func ClearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureCookie(),
		SameSite: getCookieSameSite(),
	})
}

// ExtractRefreshTokenFromCookie extracts the refresh token from the cookie
func ExtractRefreshTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil {
		return "", errors.New("refresh cookie not found")
	}
	if cookie.Value == "" {
		return "", errors.New("refresh cookie is empty")
	}
	return cookie.Value, nil
}

// ExtractTokenFromCookie extracts the JWT token from the auth cookie
func ExtractTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(AuthCookieName)
	if err != nil {
		return "", errors.New("auth cookie not found")
	}
	if cookie.Value == "" {
		return "", errors.New("auth cookie is empty")
	}
	return cookie.Value, nil
}
