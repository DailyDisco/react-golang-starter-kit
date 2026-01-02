package validation

import (
	"strings"
	"testing"
)

func TestValidationErrors_Error(t *testing.T) {
	t.Run("empty errors returns generic message", func(t *testing.T) {
		ve := New()
		msg := ve.Error()

		if msg != "validation failed" {
			t.Errorf("Expected 'validation failed', got: %s", msg)
		}
	})

	t.Run("single error formats correctly", func(t *testing.T) {
		ve := New()
		ve.Add("email", "Invalid email format", "email")

		msg := ve.Error()

		if !strings.Contains(msg, "email") {
			t.Error("Expected message to contain 'email'")
		}
		if !strings.Contains(msg, "Invalid email format") {
			t.Error("Expected message to contain 'Invalid email format'")
		}
	})

	t.Run("multiple errors joined with semicolon", func(t *testing.T) {
		ve := New()
		ve.Add("email", "Invalid email", "email")
		ve.Add("password", "Too short", "min")

		msg := ve.Error()

		if !strings.Contains(msg, ";") {
			t.Error("Expected errors to be joined with semicolon")
		}
		if !strings.Contains(msg, "email") {
			t.Error("Expected message to contain 'email'")
		}
		if !strings.Contains(msg, "password") {
			t.Error("Expected message to contain 'password'")
		}
	})
}

func TestValidationErrors_Add(t *testing.T) {
	t.Run("adds error without value", func(t *testing.T) {
		ve := New()
		ve.Add("field", "message", "code")

		if len(ve.Errors) != 1 {
			t.Fatalf("Expected 1 error, got: %d", len(ve.Errors))
		}

		err := ve.Errors[0]
		if err.Field != "field" {
			t.Errorf("Expected field 'field', got: %s", err.Field)
		}
		if err.Message != "message" {
			t.Errorf("Expected message 'message', got: %s", err.Message)
		}
		if err.Code != "code" {
			t.Errorf("Expected code 'code', got: %s", err.Code)
		}
		if err.Value != nil {
			t.Error("Expected Value to be nil")
		}
	})

	t.Run("adds multiple errors", func(t *testing.T) {
		ve := New()
		ve.Add("field1", "message1", "code1")
		ve.Add("field2", "message2", "code2")

		if len(ve.Errors) != 2 {
			t.Fatalf("Expected 2 errors, got: %d", len(ve.Errors))
		}
	})
}

func TestValidationErrors_AddWithValue(t *testing.T) {
	t.Run("adds error with string value", func(t *testing.T) {
		ve := New()
		ve.AddWithValue("email", "Invalid email", "email", "bad@")

		if len(ve.Errors) != 1 {
			t.Fatalf("Expected 1 error, got: %d", len(ve.Errors))
		}

		err := ve.Errors[0]
		if err.Value != "bad@" {
			t.Errorf("Expected value 'bad@', got: %v", err.Value)
		}
	})

	t.Run("adds error with integer value", func(t *testing.T) {
		ve := New()
		ve.AddWithValue("age", "Must be positive", "gte", -5)

		err := ve.Errors[0]
		if err.Value != -5 {
			t.Errorf("Expected value -5, got: %v", err.Value)
		}
	})

	t.Run("adds error with nil value", func(t *testing.T) {
		ve := New()
		ve.AddWithValue("field", "message", "code", nil)

		err := ve.Errors[0]
		if err.Value != nil {
			t.Error("Expected Value to be nil")
		}
	})
}

func TestValidationErrors_HasErrors(t *testing.T) {
	t.Run("returns false for empty errors", func(t *testing.T) {
		ve := New()

		if ve.HasErrors() {
			t.Error("Expected HasErrors to return false")
		}
	})

	t.Run("returns true when errors exist", func(t *testing.T) {
		ve := New()
		ve.Add("field", "message", "code")

		if !ve.HasErrors() {
			t.Error("Expected HasErrors to return true")
		}
	})
}

func TestValidationErrors_Count(t *testing.T) {
	t.Run("returns 0 for empty errors", func(t *testing.T) {
		ve := New()

		if ve.Count() != 0 {
			t.Errorf("Expected count 0, got: %d", ve.Count())
		}
	})

	t.Run("returns correct count", func(t *testing.T) {
		ve := New()
		ve.Add("field1", "message1", "code1")
		ve.Add("field2", "message2", "code2")
		ve.Add("field3", "message3", "code3")

		if ve.Count() != 3 {
			t.Errorf("Expected count 3, got: %d", ve.Count())
		}
	})
}

func TestValidationErrors_First(t *testing.T) {
	t.Run("returns nil for empty errors", func(t *testing.T) {
		ve := New()

		first := ve.First()
		if first != nil {
			t.Error("Expected First to return nil")
		}
	})

	t.Run("returns first error", func(t *testing.T) {
		ve := New()
		ve.Add("field1", "message1", "code1")
		ve.Add("field2", "message2", "code2")

		first := ve.First()
		if first == nil {
			t.Fatal("Expected First to return an error")
		}
		if first.Field != "field1" {
			t.Errorf("Expected first field 'field1', got: %s", first.Field)
		}
	})
}

func TestValidationErrors_GetField(t *testing.T) {
	t.Run("returns nil for non-existent field", func(t *testing.T) {
		ve := New()
		ve.Add("email", "Invalid", "email")

		result := ve.GetField("password")
		if result != nil {
			t.Error("Expected GetField to return nil for non-existent field")
		}
	})

	t.Run("returns error for existing field", func(t *testing.T) {
		ve := New()
		ve.Add("email", "Invalid email", "email")
		ve.Add("password", "Too short", "min")

		result := ve.GetField("email")
		if result == nil {
			t.Fatal("Expected GetField to return an error")
		}
		if result.Field != "email" {
			t.Errorf("Expected field 'email', got: %s", result.Field)
		}
		if result.Message != "Invalid email" {
			t.Errorf("Expected message 'Invalid email', got: %s", result.Message)
		}
	})

	t.Run("returns first error when multiple for same field", func(t *testing.T) {
		ve := New()
		ve.Add("email", "First error", "required")
		ve.Add("email", "Second error", "email")

		result := ve.GetField("email")
		if result == nil {
			t.Fatal("Expected GetField to return an error")
		}
		if result.Message != "First error" {
			t.Errorf("Expected 'First error', got: %s", result.Message)
		}
	})
}

func TestNew(t *testing.T) {
	t.Run("creates empty validation errors", func(t *testing.T) {
		ve := New()

		if ve == nil {
			t.Fatal("Expected non-nil ValidationErrors")
		}
		if ve.Errors == nil {
			t.Error("Expected Errors slice to be initialized")
		}
		if len(ve.Errors) != 0 {
			t.Errorf("Expected empty Errors slice, got: %d", len(ve.Errors))
		}
	})
}

func TestNewWithError(t *testing.T) {
	t.Run("creates validation errors with single error", func(t *testing.T) {
		ve := NewWithError("email", "Invalid email format", "email")

		if ve == nil {
			t.Fatal("Expected non-nil ValidationErrors")
		}
		if len(ve.Errors) != 1 {
			t.Fatalf("Expected 1 error, got: %d", len(ve.Errors))
		}

		err := ve.Errors[0]
		if err.Field != "email" {
			t.Errorf("Expected field 'email', got: %s", err.Field)
		}
		if err.Message != "Invalid email format" {
			t.Errorf("Expected message 'Invalid email format', got: %s", err.Message)
		}
		if err.Code != "email" {
			t.Errorf("Expected code 'email', got: %s", err.Code)
		}
	})
}

func TestIsSensitiveField(t *testing.T) {
	tests := []struct {
		field     string
		sensitive bool
	}{
		{"password", true},
		{"PASSWORD", true},
		{"Password", true},
		{"password_confirm", true},
		{"current_password", true},
		{"new_password", true},
		{"token", true},
		{"secret", true},
		{"api_key", true},
		{"credit_card", true},
		{"ssn", true},
		{"email", false},
		{"name", false},
		{"username", false},
		{"role", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := IsSensitiveField(tt.field)
			if result != tt.sensitive {
				t.Errorf("IsSensitiveField(%s) = %v, want %v", tt.field, result, tt.sensitive)
			}
		})
	}
}

func TestValidationErrors_Implements_Error_Interface(t *testing.T) {
	var _ error = (*ValidationErrors)(nil)

	// Also test that it works as an error
	ve := NewWithError("field", "message", "code")
	var err error = ve

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func BenchmarkValidationErrors_Add(b *testing.B) {
	ve := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ve.Add("field", "message", "code")
	}
}

func BenchmarkValidationErrors_GetField(b *testing.B) {
	ve := New()
	for i := 0; i < 10; i++ {
		ve.Add("field"+string(rune('a'+i)), "message", "code")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ve.GetField("fielde")
	}
}

func BenchmarkIsSensitiveField(b *testing.B) {
	fields := []string{"password", "email", "name", "api_key", "username"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsSensitiveField(fields[i%len(fields)])
	}
}
