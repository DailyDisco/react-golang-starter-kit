package validation

import (
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once

	// Email regex pattern (RFC 5322 compliant)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
)

// GetValidator returns the singleton validator instance.
func GetValidator() *validator.Validate {
	once.Do(func() {
		validate = validator.New()

		// Use JSON tag names for field names in error messages
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return fld.Name
			}
			if name == "" {
				return fld.Name
			}
			return name
		})

		// Register custom validators
		validate.RegisterValidation("password", validatePassword)
		validate.RegisterValidation("strong_email", validateEmail)
	})
	return validate
}

// ValidateStruct validates a struct and returns structured errors.
// Returns nil if validation passes.
func ValidateStruct(s interface{}) *ValidationErrors {
	v := GetValidator()
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		// Not a validation error, return generic error
		ve := New()
		ve.Add("", err.Error(), "unknown")
		return ve
	}

	ve := New()
	for _, e := range validationErrs {
		field := e.Field()
		message := formatValidationMessage(e)
		code := e.Tag()

		// Don't include value for sensitive fields
		if IsSensitiveField(field) {
			ve.Add(field, message, code)
		} else {
			ve.AddWithValue(field, message, code, e.Value())
		}
	}

	return ve
}

// ValidateVar validates a single variable against a tag.
func ValidateVar(field interface{}, tag string) error {
	v := GetValidator()
	return v.Var(field, tag)
}

// formatValidationMessage returns a human-readable message for a validation error.
func formatValidationMessage(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case "required":
		return field + " is required"
	case "email", "strong_email":
		return "Invalid email format"
	case "min":
		if e.Type().Kind() == reflect.String {
			return field + " must be at least " + e.Param() + " characters"
		}
		return field + " must be at least " + e.Param()
	case "max":
		if e.Type().Kind() == reflect.String {
			return field + " must not exceed " + e.Param() + " characters"
		}
		return field + " must not exceed " + e.Param()
	case "len":
		return field + " must be exactly " + e.Param() + " characters"
	case "password":
		return "Password must be at least 8 characters with uppercase, lowercase, and a number"
	case "oneof":
		return field + " must be one of: " + e.Param()
	case "url":
		return "Invalid URL format"
	case "uuid":
		return "Invalid UUID format"
	case "numeric":
		return field + " must be numeric"
	case "alpha":
		return field + " must contain only letters"
	case "alphanum":
		return field + " must contain only letters and numbers"
	case "boolean":
		return field + " must be a boolean"
	case "gt":
		return field + " must be greater than " + e.Param()
	case "gte":
		return field + " must be greater than or equal to " + e.Param()
	case "lt":
		return field + " must be less than " + e.Param()
	case "lte":
		return field + " must be less than or equal to " + e.Param()
	case "eqfield":
		return field + " must match " + e.Param()
	case "nefield":
		return field + " must not match " + e.Param()
	default:
		return field + " failed validation: " + e.Tag()
	}
}

// validatePassword validates password strength.
// Password must be at least 8 characters with uppercase, lowercase, and a number.
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasDigit := strings.ContainsAny(password, "0123456789")

	return hasUpper && hasLower && hasDigit
}

// validateEmail validates email using a comprehensive regex pattern.
func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	if len(email) < 5 || len(email) > 254 {
		return false
	}

	if !emailRegex.MatchString(email) {
		return false
	}

	// Additional edge case validation
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local, domain := parts[0], parts[1]

	// Local part validation
	if len(local) == 0 || len(local) > 64 {
		return false
	}
	if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") {
		return false
	}
	if strings.Contains(local, "..") {
		return false
	}

	// Domain validation
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}
	if strings.Contains(domain, "..") {
		return false
	}
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

// Common validation tag combinations for reuse
const (
	// TagRequired is the tag for required fields
	TagRequired = "required"

	// TagEmail is the tag for email validation
	TagEmail = "required,email"

	// TagStrongEmail is the tag for strong email validation
	TagStrongEmail = "required,strong_email"

	// TagPassword is the tag for password validation
	TagPassword = "required,password"

	// TagName is the tag for name fields
	TagName = "required,min=1,max=100"

	// TagOptionalString is the tag for optional string fields with max length
	TagOptionalString = "omitempty,max=500"
)
