package validator

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"

	"github.com/nikola43/aureo-vpn/pkg/errors"
)

// Validator performs input validation
type Validator struct {
	errors []errors.ValidationError
}

// New creates a new validator
func New() *Validator {
	return &Validator{
		errors: []errors.ValidationError{},
	}
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, errors.ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors
func (v *Validator) Errors() []errors.ValidationError {
	return v.errors
}

// Error returns an AppError with all validation errors
func (v *Validator) Error() *errors.AppError {
	if !v.HasErrors() {
		return nil
	}
	return errors.NewValidationError(v.errors)
}

// Required validates that a field is not empty
func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, fmt.Sprintf("%s is required", field))
	}
}

// Email validates an email address
func (v *Validator) Email(field, value string) {
	if value == "" {
		return
	}
	_, err := mail.ParseAddress(value)
	if err != nil {
		v.AddError(field, "invalid email address")
	}
}

// MinLength validates minimum string length
func (v *Validator) MinLength(field, value string, min int) {
	if len(value) < min {
		v.AddError(field, fmt.Sprintf("%s must be at least %d characters", field, min))
	}
}

// MaxLength validates maximum string length
func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.AddError(field, fmt.Sprintf("%s must be at most %d characters", field, max))
	}
}

// Length validates exact string length
func (v *Validator) Length(field, value string, length int) {
	if len(value) != length {
		v.AddError(field, fmt.Sprintf("%s must be exactly %d characters", field, length))
	}
}

// MinValue validates minimum numeric value
func (v *Validator) MinValue(field string, value, min int) {
	if value < min {
		v.AddError(field, fmt.Sprintf("%s must be at least %d", field, min))
	}
}

// MaxValue validates maximum numeric value
func (v *Validator) MaxValue(field string, value, max int) {
	if value > max {
		v.AddError(field, fmt.Sprintf("%s must be at most %d", field, max))
	}
}

// Range validates value is within a range
func (v *Validator) Range(field string, value, min, max int) {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("%s must be between %d and %d", field, min, max))
	}
}

// In validates that value is in a list of allowed values
func (v *Validator) In(field, value string, allowed []string) {
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.AddError(field, fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowed, ", ")))
}

// Match validates that a value matches a regular expression
func (v *Validator) Match(field, value, pattern string) {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		v.AddError(field, fmt.Sprintf("%s has invalid format", field))
	}
}

// Username validates a username
func (v *Validator) Username(field, value string) {
	if value == "" {
		return
	}
	// Username: 3-50 chars, alphanumeric, underscores, hyphens
	pattern := `^[a-zA-Z0-9_-]{3,50}$`
	matched, _ := regexp.MatchString(pattern, value)
	if !matched {
		v.AddError(field, "username must be 3-50 characters and contain only letters, numbers, underscores, and hyphens")
	}
}

// Password validates a password with security requirements
func (v *Validator) Password(field, value string, minLength int) {
	if value == "" {
		return
	}

	if len(value) < minLength {
		v.AddError(field, fmt.Sprintf("password must be at least %d characters", minLength))
		return
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range value {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		v.AddError(field, "password must contain at least one uppercase letter")
	}
	if !hasLower {
		v.AddError(field, "password must contain at least one lowercase letter")
	}
	if !hasNumber {
		v.AddError(field, "password must contain at least one number")
	}
	if !hasSpecial {
		v.AddError(field, "password must contain at least one special character")
	}
}

// UUID validates a UUID
func (v *Validator) UUID(field, value string) {
	if value == "" {
		return
	}
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(pattern, value)
	if !matched {
		v.AddError(field, "invalid UUID format")
	}
}

// IP validates an IP address
func (v *Validator) IP(field, value string) {
	if value == "" {
		return
	}
	// Simple IPv4/IPv6 validation
	pattern := `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$|^(?:[A-F0-9]{1,4}:){7}[A-F0-9]{1,4}$`
	matched, _ := regexp.MatchString(pattern, strings.ToUpper(value))
	if !matched {
		v.AddError(field, "invalid IP address")
	}
}

// URL validates a URL
func (v *Validator) URL(field, value string) {
	if value == "" {
		return
	}
	pattern := `^https?://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=]+$`
	matched, _ := regexp.MatchString(pattern, value)
	if !matched {
		v.AddError(field, "invalid URL")
	}
}

// Hostname validates a hostname
func (v *Validator) Hostname(field, value string) {
	if value == "" {
		return
	}
	pattern := `^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`
	matched, _ := regexp.MatchString(pattern, value)
	if !matched {
		v.AddError(field, "invalid hostname")
	}
}

// CountryCode validates a 2-letter country code
func (v *Validator) CountryCode(field, value string) {
	if value == "" {
		return
	}
	pattern := `^[A-Z]{2}$`
	matched, _ := regexp.MatchString(pattern, strings.ToUpper(value))
	if !matched {
		v.AddError(field, "invalid country code (must be 2 uppercase letters)")
	}
}

// Protocol validates a VPN protocol
func (v *Validator) Protocol(field, value string) {
	allowed := []string{"wireguard", "openvpn", "ipsec", "ikev2"}
	v.In(field, strings.ToLower(value), allowed)
}

// Port validates a port number
func (v *Validator) Port(field string, value int) {
	if value < 1 || value > 65535 {
		v.AddError(field, "port must be between 1 and 65535")
	}
}

// Alphanumeric validates that a string contains only alphanumeric characters
func (v *Validator) Alphanumeric(field, value string) {
	if value == "" {
		return
	}
	pattern := `^[a-zA-Z0-9]+$`
	matched, _ := regexp.MatchString(pattern, value)
	if !matched {
		v.AddError(field, fmt.Sprintf("%s must contain only letters and numbers", field))
	}
}

// Base64 validates base64 encoding
func (v *Validator) Base64(field, value string) {
	if value == "" {
		return
	}
	pattern := `^[A-Za-z0-9+/]*={0,2}$`
	matched, _ := regexp.MatchString(pattern, value)
	if !matched {
		v.AddError(field, "invalid base64 encoding")
	}
}

// Helper functions for common validations

// ValidateRegistration validates user registration request
func ValidateRegistration(email, password, username string) *errors.AppError {
	v := New()
	v.Required("email", email)
	v.Email("email", email)
	v.Required("password", password)
	v.Password("password", password, 8)
	v.Required("username", username)
	v.Username("username", username)
	return v.Error()
}

// ValidateLogin validates login request
func ValidateLogin(email, password string) *errors.AppError {
	v := New()
	v.Required("email", email)
	v.Email("email", email)
	v.Required("password", password)
	return v.Error()
}

// ValidateNodeCreation validates VPN node creation
func ValidateNodeCreation(name, hostname, ip, country, countryCode, city string, wgPort, ovpnPort int) *errors.AppError {
	v := New()
	v.Required("name", name)
	v.MinLength("name", name, 3)
	v.MaxLength("name", name, 100)
	v.Required("hostname", hostname)
	v.Hostname("hostname", hostname)
	v.Required("ip", ip)
	v.IP("ip", ip)
	v.Required("country", country)
	v.Required("country_code", countryCode)
	v.CountryCode("country_code", countryCode)
	v.Required("city", city)
	v.Port("wireguard_port", wgPort)
	v.Port("openvpn_port", ovpnPort)
	return v.Error()
}
