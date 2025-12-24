// Package sanitize provides input sanitization utilities to prevent XSS and injection attacks.
package sanitize

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

// Common dangerous patterns to remove
var (
	// scriptTagRegex matches script tags and their content
	scriptTagRegex = regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>`)

	// styleTagRegex matches style tags and their content
	styleTagRegex = regexp.MustCompile(`(?i)<style[^>]*>[\s\S]*?</style>`)

	// eventHandlerRegex matches event handler attributes
	eventHandlerRegex = regexp.MustCompile(`(?i)\s+on\w+\s*=\s*["'][^"']*["']`)

	// htmlTagRegex matches any HTML tag
	htmlTagRegex = regexp.MustCompile(`<[^>]+>`)

	// multipleSpacesRegex matches multiple consecutive spaces
	multipleSpacesRegex = regexp.MustCompile(`\s+`)

	// dangerousProtocolRegex matches dangerous URL protocols
	dangerousProtocolRegex = regexp.MustCompile(`(?i)^\s*(javascript|data|vbscript|file):`)

	// sqlInjectionPatterns contains common SQL injection patterns
	sqlInjectionPatterns = regexp.MustCompile(`(?i)(union\s+select|drop\s+table|insert\s+into|delete\s+from|update\s+.*set|;\s*--)`)

	// emailRegex validates email format
	emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

	// pathTraversalRegex matches path traversal attempts
	pathTraversalRegex = regexp.MustCompile(`\.\.[\\/]`)

	// nullByteRegex matches null bytes
	nullByteRegex = regexp.MustCompile(`\x00`)
)

// Text sanitizes text input by stripping HTML tags and normalizing whitespace.
// Use for usernames, titles, and single-line text inputs.
func Text(input string) string {
	if input == "" {
		return ""
	}

	// Remove script tags and content
	result := scriptTagRegex.ReplaceAllString(input, "")

	// Remove style tags and content
	result = styleTagRegex.ReplaceAllString(result, "")

	// Remove event handlers
	result = eventHandlerRegex.ReplaceAllString(result, "")

	// Remove all HTML tags
	result = htmlTagRegex.ReplaceAllString(result, "")

	// Decode HTML entities
	result = html.UnescapeString(result)

	// Normalize whitespace
	result = multipleSpacesRegex.ReplaceAllString(result, " ")

	// Trim leading/trailing whitespace
	result = strings.TrimSpace(result)

	return result
}

// HTML escapes HTML special characters for safe display.
// Use when you need to display user input in HTML context without allowing formatting.
func HTML(input string) string {
	if input == "" {
		return ""
	}
	return html.EscapeString(input)
}

// URL sanitizes a URL, blocking dangerous protocols.
// Returns empty string if the URL is potentially dangerous.
func URL(input string) string {
	if input == "" {
		return ""
	}

	trimmed := strings.TrimSpace(input)

	// Check for dangerous protocols
	if dangerousProtocolRegex.MatchString(trimmed) {
		return ""
	}

	// Allow only safe protocols or relative URLs
	lowerURL := strings.ToLower(trimmed)
	safeProtocols := []string{"http://", "https://", "mailto:", "tel:"}

	hasProtocol := false
	for _, proto := range safeProtocols {
		if strings.HasPrefix(lowerURL, proto) {
			hasProtocol = true
			break
		}
	}

	// If URL has a protocol but it's not in our safe list, reject it
	if strings.Contains(lowerURL, "://") && !hasProtocol {
		return ""
	}

	return trimmed
}

// Email sanitizes and validates an email address.
// Returns empty string if the email is invalid.
func Email(input string) string {
	if input == "" {
		return ""
	}

	// Strip HTML and normalize
	sanitized := Text(input)
	sanitized = strings.ToLower(sanitized)

	// Validate email format
	if !emailRegex.MatchString(sanitized) {
		return ""
	}

	return sanitized
}

// Filename sanitizes a filename for safe storage.
// Removes path traversal attempts and dangerous characters.
func Filename(input string) string {
	if input == "" {
		return ""
	}

	result := input

	// Remove path traversal attempts
	result = pathTraversalRegex.ReplaceAllString(result, "")
	result = strings.ReplaceAll(result, "/", "")
	result = strings.ReplaceAll(result, "\\", "")

	// Remove null bytes
	result = nullByteRegex.ReplaceAllString(result, "")

	// Remove control characters
	result = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1 // Remove the character
		}
		return r
	}, result)

	// Keep only safe characters: alphanumeric, dots, hyphens, underscores
	var builder strings.Builder
	for _, r := range result {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == '_' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}
	result = builder.String()

	// Prevent empty or dot-only filenames
	if result == "" || result == "." || result == ".." {
		return "unnamed"
	}

	return result
}

// SQLString checks for common SQL injection patterns.
// Returns true if the input appears safe, false if it contains suspicious patterns.
// Note: This is a defense-in-depth measure. Always use parameterized queries!
func SQLString(input string) bool {
	if input == "" {
		return true
	}

	return !sqlInjectionPatterns.MatchString(input)
}

// StripNullBytes removes null bytes from input.
// Null bytes can be used to bypass security filters.
func StripNullBytes(input string) string {
	return nullByteRegex.ReplaceAllString(input, "")
}

// TruncateString truncates a string to the specified length.
// Useful for preventing buffer overflow or storage issues.
func TruncateString(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}
	return input[:maxLength]
}

// NormalizeName normalizes a person's name.
// Capitalizes first letter of each word, removes extra whitespace.
func NormalizeName(input string) string {
	if input == "" {
		return ""
	}

	// First sanitize the input
	sanitized := Text(input)

	// Title case each word
	words := strings.Fields(sanitized)
	for i, word := range words {
		if len(word) > 0 {
			// Capitalize first letter, lowercase rest
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// Password validates password length and content.
// Returns an error message if invalid, empty string if valid.
// Does NOT sanitize - passwords should be hashed, not filtered.
func Password(input string, minLength, maxLength int) string {
	if len(input) < minLength {
		return "password too short"
	}
	if len(input) > maxLength {
		return "password too long"
	}
	// Check for null bytes which could cause issues
	if strings.Contains(input, "\x00") {
		return "password contains invalid characters"
	}
	return ""
}
