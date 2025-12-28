package security

import (
	"strings"
	"testing"
)

// =============================================================================
// Email Validation Tests
// =============================================================================

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"valid with subdomain", "user@mail.example.com", false},
		{"valid with plus", "user+tag@example.com", false},
		{"valid with dots", "user.name@example.com", false},
		{"empty email", "", true},
		{"missing @", "userexample.com", true},
		{"missing domain", "user@", true},
		{"missing local part", "@example.com", true},
		{"spaces", "user @example.com", true},
		{"multiple @", "user@@example.com", true},
		{"too long", strings.Repeat("a", 250) + "@example.com", true},
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

// =============================================================================
// Username Validation Tests
// =============================================================================

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid lowercase", "testuser", false},
		{"valid with numbers", "user123", false},
		{"valid with underscore", "test_user", false},
		{"valid with hyphen", "test-user", false},
		{"minimum length", "abc", false},
		{"maximum length", strings.Repeat("a", 64), false},
		{"too short", "ab", true},
		{"too long", strings.Repeat("a", 65), true},
		{"empty", "", true},
		{"spaces", "test user", true},
		{"special chars", "test@user", true},
		{"starts with number", "1testuser", true},
		{"uppercase allowed", "TestUser", false}, // Implementation allows uppercase
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername(%q) error = %v, wantErr %v", tt.username, err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Password Validation Tests
// =============================================================================

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid strong password", "SecureP@ssw0rd!", false},
		{"valid with special chars", "Test123!@#$%Ab", false}, // 14 chars min
		{"minimum 12 chars", "Aa1!aabbccdd", false},          // 12 chars
		{"empty password", "", true},
		{"too short 11 chars", "Aa1!aabbccd", true},
		{"no uppercase", "password123!ab", true},
		{"no lowercase", "PASSWORD123!AB", true},
		{"no number", "Password!@#Abc", true},
		{"no special char", "Password123Ab", true},
		{"spaces only", "            ", true},
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

// =============================================================================
// UUID Validation Tests
// =============================================================================

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{"valid uuid v4", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid uuid lowercase", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid uuid uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"empty", "", true},
		{"invalid format", "550e8400-e29b-41d4-a716", true},
		{"no hyphens", "550e8400e29b41d4a716446655440000", true},
		{"invalid chars", "550e8400-e29b-41d4-a716-44665544000g", true},
		{"too short", "550e8400-e29b-41d4", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID(tt.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUUID(%q) error = %v, wantErr %v", tt.uuid, err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Validator Instance Tests
// =============================================================================

func TestValidator(t *testing.T) {
	v := NewValidator(true) // Strict mode

	// Test with validator instance
	if err := v.ValidateEmail("test@example.com"); err != nil {
		t.Errorf("Validator.ValidateEmail() error = %v", err)
	}

	if err := v.ValidateUsername("testuser"); err != nil {
		t.Errorf("Validator.ValidateUsername() error = %v", err)
	}

	if err := v.ValidateUUID("550e8400-e29b-41d4-a716-446655440000"); err != nil {
		t.Errorf("Validator.ValidateUUID() error = %v", err)
	}
}

// =============================================================================
// Password Policy Tests
// =============================================================================

func TestPasswordPolicies(t *testing.T) {
	v := NewValidator(true)

	// Default policy requires 12 chars, no more than 3 repeating
	defaultPolicy := DefaultPasswordPolicy()
	if err := v.ValidatePassword("Abc1!defghij", defaultPolicy); err != nil {
		t.Errorf("ValidatePassword with default policy error = %v", err)
	}

	// High security policy requires 16 chars
	highSecPolicy := HighSecurityPasswordPolicy()
	err := v.ValidatePassword("Aa1!bcde", highSecPolicy)
	if err == nil {
		t.Error("ValidatePassword with high security policy should fail for short password")
	}

	// Valid for high security (16+ chars, no more than 2 repeating)
	if err := v.ValidatePassword("SecureP@ssw0rd12!", highSecPolicy); err != nil {
		t.Errorf("ValidatePassword with high security policy error = %v", err)
	}
}

// =============================================================================
// Input Sanitization Tests
// =============================================================================

func TestSanitizeString(t *testing.T) {
	v := NewValidator(true)

	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"normal string", "hello world", 100, "hello world"},
		{"preserve spaces", "  hello  ", 100, "  hello  "}, // SanitizeString doesn't trim
		{"truncate", "hello world", 5, "hello"},
		{"preserve tags", "<script>alert('xss')</script>", 100, "<script>alert('xss')</script>"}, // SanitizeString doesn't escape HTML
		{"preserve html", "<>&\"'", 100, "<>&\"'"},
		{"null bytes removed", "hello\x00world", 100, "helloworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.SanitizeString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("SanitizeString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	v := NewValidator(true)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"html entities", "<>&\"'", "&lt;&gt;&amp;&quot;&#x27;"},
		{"script tag", "<script>", "&lt;script&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.SanitizeHTML(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeHTML() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// Common Password Detection Tests
// =============================================================================

func TestCommonPasswordDetection(t *testing.T) {
	v := NewValidator(true)

	// These passwords have the base common words in the check list
	commonBases := []string{
		"password",
		"qwerty",
		"admin",
		"letmein",
		"welcome",
	}

	// Test that these common bases are detected by the isCommonPassword check
	for _, base := range commonBases {
		t.Run(base, func(t *testing.T) {
			if !v.isCommonPassword(base) {
				t.Errorf("isCommonPassword(%q) should return true", base)
			}
		})
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkValidateEmail(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateEmail("test@example.com")
	}
}

func BenchmarkValidatePassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidatePassword("SecureP@ssw0rd123!")
	}
}

func BenchmarkValidateUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateUUID("550e8400-e29b-41d4-a716-446655440000")
	}
}
