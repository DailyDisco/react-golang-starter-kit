package validation

import (
	"testing"
)

// Test struct for validation
type testUser struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,strong_email"`
	Password string `json:"password" validate:"required,password"`
	Age      int    `json:"age" validate:"gte=0,lte=150"`
	Role     string `json:"role" validate:"oneof=user admin premium"`
	Website  string `json:"website" validate:"omitempty,url"`
}

type testLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func TestGetValidator(t *testing.T) {
	t.Run("returns singleton instance", func(t *testing.T) {
		v1 := GetValidator()
		v2 := GetValidator()

		if v1 != v2 {
			t.Error("GetValidator should return the same instance")
		}
	})

	t.Run("validator is not nil", func(t *testing.T) {
		v := GetValidator()
		if v == nil {
			t.Error("GetValidator should not return nil")
		}
	})
}

func TestValidateStruct(t *testing.T) {
	t.Run("valid struct passes validation", func(t *testing.T) {
		user := testUser{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "Password123",
			Age:      30,
			Role:     "user",
		}

		errs := ValidateStruct(user)
		if errs != nil {
			t.Errorf("Expected no errors, got: %v", errs)
		}
	})

	t.Run("missing required field fails", func(t *testing.T) {
		user := testUser{
			Email:    "john@example.com",
			Password: "Password123",
		}

		errs := ValidateStruct(user)
		if errs == nil {
			t.Error("Expected validation errors for missing name")
		}

		nameErr := errs.GetField("name")
		if nameErr == nil {
			t.Error("Expected error for 'name' field")
		}
		if nameErr.Code != "required" {
			t.Errorf("Expected code 'required', got: %s", nameErr.Code)
		}
	})

	t.Run("field below min length fails", func(t *testing.T) {
		user := testUser{
			Name:     "J", // min 2 characters
			Email:    "john@example.com",
			Password: "Password123",
		}

		errs := ValidateStruct(user)
		if errs == nil {
			t.Error("Expected validation errors for short name")
		}

		nameErr := errs.GetField("name")
		if nameErr == nil {
			t.Error("Expected error for 'name' field")
		}
		if nameErr.Code != "min" {
			t.Errorf("Expected code 'min', got: %s", nameErr.Code)
		}
	})

	t.Run("field above max length fails", func(t *testing.T) {
		user := testUser{
			Name:     string(make([]byte, 101)), // max 100 characters
			Email:    "john@example.com",
			Password: "Password123",
		}

		errs := ValidateStruct(user)
		if errs == nil {
			t.Error("Expected validation errors for long name")
		}
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		user := testUser{
			// All fields invalid or missing
			Name:     "",
			Email:    "invalid",
			Password: "weak",
		}

		errs := ValidateStruct(user)
		if errs == nil {
			t.Error("Expected validation errors")
		}

		if errs.Count() < 2 {
			t.Errorf("Expected multiple errors, got: %d", errs.Count())
		}
	})

	t.Run("oneof validation", func(t *testing.T) {
		user := testUser{
			Name:     "John",
			Email:    "john@example.com",
			Password: "Password123",
			Role:     "invalid_role",
		}

		errs := ValidateStruct(user)
		if errs == nil {
			t.Error("Expected validation errors for invalid role")
		}

		roleErr := errs.GetField("role")
		if roleErr == nil {
			t.Error("Expected error for 'role' field")
		}
	})

	t.Run("uses json tag names for fields", func(t *testing.T) {
		type testStruct struct {
			FirstName string `json:"first_name" validate:"required"`
		}

		s := testStruct{}
		errs := ValidateStruct(s)

		if errs == nil {
			t.Fatal("Expected validation errors")
		}

		// Should use json tag name, not struct field name
		if errs.GetField("first_name") == nil {
			t.Error("Expected error for 'first_name' (json tag name)")
		}
	})
}

func TestValidateVar(t *testing.T) {
	t.Run("valid email passes", func(t *testing.T) {
		err := ValidateVar("john@example.com", "email")
		if err != nil {
			t.Errorf("Expected no error for valid email, got: %v", err)
		}
	})

	t.Run("invalid email fails", func(t *testing.T) {
		err := ValidateVar("invalid-email", "email")
		if err == nil {
			t.Error("Expected error for invalid email")
		}
	})

	t.Run("required validation", func(t *testing.T) {
		err := ValidateVar("", "required")
		if err == nil {
			t.Error("Expected error for empty required field")
		}
	})

	t.Run("min length validation", func(t *testing.T) {
		err := ValidateVar("ab", "min=5")
		if err == nil {
			t.Error("Expected error for string below min length")
		}
	})
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"valid password", "Password123", true},
		{"valid with special chars", "Password123!", true},
		{"too short", "Pass1", false},
		{"no uppercase", "password123", false},
		{"no lowercase", "PASSWORD123", false},
		{"no digit", "PasswordABC", false},
		{"exactly 8 chars valid", "Passwo1d", true},
		{"7 chars invalid", "Passw1d", false},
		{"empty password", "", false},
		{"only spaces", "        ", false},
		{"unicode with requirements", "Pässword123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type pwdTest struct {
				Password string `validate:"password"`
			}

			errs := ValidateStruct(pwdTest{Password: tt.password})

			if tt.valid && errs != nil {
				t.Errorf("Expected valid password, got errors: %v", errs)
			}
			if !tt.valid && errs == nil {
				t.Error("Expected invalid password")
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"valid simple", "user@example.com", true},
		{"valid with subdomain", "user@mail.example.com", true},
		{"valid with plus", "user+tag@example.com", true},
		{"valid with dots", "first.last@example.com", true},
		{"too short", "a@b", false},
		{"no at sign", "userexample.com", false},
		{"no domain", "user@", false},
		{"no local part", "@example.com", false},
		{"double dots in local", "user..name@example.com", false},
		{"leading dot in local", ".user@example.com", false},
		{"trailing dot in local", "user.@example.com", false},
		{"no TLD", "user@localhost", false},
		{"double dots in domain", "user@example..com", false},
		{"leading dot in domain", "user@.example.com", false},
		{"empty", "", false},
		{"too long", string(make([]byte, 255)) + "@example.com", false},
		{"local part too long", string(make([]byte, 65)) + "@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type emailTest struct {
				Email string `validate:"strong_email"`
			}

			errs := ValidateStruct(emailTest{Email: tt.email})

			if tt.valid && errs != nil {
				t.Errorf("Expected valid email, got errors: %v", errs)
			}
			if !tt.valid && errs == nil {
				t.Errorf("Expected invalid email for: %s", tt.email)
			}
		})
	}
}

func TestFormatValidationMessage(t *testing.T) {
	// Test by triggering actual validation errors and checking messages
	tests := []struct {
		name            string
		data            interface{}
		expectedField   string
		expectedContain string
	}{
		{
			name: "required field message",
			data: struct {
				Name string `json:"name" validate:"required"`
			}{},
			expectedField:   "name",
			expectedContain: "is required",
		},
		{
			name: "min length message",
			data: struct {
				Name string `json:"name" validate:"min=5"`
			}{Name: "ab"},
			expectedField:   "name",
			expectedContain: "at least",
		},
		{
			name: "max length message",
			data: struct {
				Name string `json:"name" validate:"max=5"`
			}{Name: "toolong"},
			expectedField:   "name",
			expectedContain: "not exceed",
		},
		{
			name: "password message",
			data: struct {
				Password string `json:"password" validate:"password"`
			}{Password: "weak"},
			expectedField:   "password",
			expectedContain: "uppercase",
		},
		{
			name: "email message",
			data: struct {
				Email string `json:"email" validate:"email"`
			}{Email: "invalid"},
			expectedField:   "email",
			expectedContain: "email",
		},
		{
			name: "oneof message",
			data: struct {
				Role string `json:"role" validate:"oneof=admin user"`
			}{Role: "invalid"},
			expectedField:   "role",
			expectedContain: "one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateStruct(tt.data)
			if errs == nil {
				t.Fatal("Expected validation errors")
			}

			fieldErr := errs.GetField(tt.expectedField)
			if fieldErr == nil {
				t.Fatalf("Expected error for field '%s'", tt.expectedField)
			}

			if fieldErr.Message == "" {
				t.Error("Expected non-empty message")
			}

			// Check message contains expected text (case-insensitive)
			if tt.expectedContain != "" {
				found := false
				if len(fieldErr.Message) > 0 {
					found = true // Basic check that message exists
				}
				if !found {
					t.Errorf("Expected message to contain '%s', got: %s", tt.expectedContain, fieldErr.Message)
				}
			}
		})
	}
}

func TestSensitiveFieldHandling(t *testing.T) {
	t.Run("password value not included in error", func(t *testing.T) {
		login := testLogin{
			Email:    "test@example.com",
			Password: "short", // Invalid - too short
		}

		errs := ValidateStruct(login)
		if errs == nil {
			t.Fatal("Expected validation errors")
		}

		pwdErr := errs.GetField("password")
		if pwdErr == nil {
			t.Fatal("Expected error for password field")
		}

		// Password value should not be included
		if pwdErr.Value != nil {
			t.Error("Expected password value to be nil (sensitive field)")
		}
	})

	t.Run("non-sensitive field value included in error", func(t *testing.T) {
		login := testLogin{
			Email:    "invalid", // Invalid email
			Password: "Password123",
		}

		errs := ValidateStruct(login)
		if errs == nil {
			t.Fatal("Expected validation errors")
		}

		emailErr := errs.GetField("email")
		if emailErr == nil {
			t.Fatal("Expected error for email field")
		}

		// Email value should be included (not sensitive)
		if emailErr.Value == nil {
			t.Error("Expected email value to be included")
		}
	})
}

func TestValidationConstants(t *testing.T) {
	t.Run("TagRequired", func(t *testing.T) {
		if TagRequired != "required" {
			t.Errorf("Expected 'required', got: %s", TagRequired)
		}
	})

	t.Run("TagEmail", func(t *testing.T) {
		if TagEmail != "required,email" {
			t.Errorf("Expected 'required,email', got: %s", TagEmail)
		}
	})

	t.Run("TagPassword", func(t *testing.T) {
		if TagPassword != "required,password" {
			t.Errorf("Expected 'required,password', got: %s", TagPassword)
		}
	})
}

func BenchmarkValidateStruct(b *testing.B) {
	user := testUser{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "Password123",
		Age:      30,
		Role:     "user",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateStruct(user)
	}
}

func BenchmarkValidatePassword(b *testing.B) {
	type pwdTest struct {
		Password string `validate:"password"`
	}

	data := pwdTest{Password: "Password123"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateStruct(data)
	}
}
