// Package security provides comprehensive input validation and sanitization
package security

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Validation errors
var (
	ErrInvalidEmail          = errors.New("invalid email address")
	ErrInvalidUsername       = errors.New("invalid username")
	ErrInvalidPassword       = errors.New("password does not meet requirements")
	ErrInvalidUUID           = errors.New("invalid UUID format")
	ErrInvalidIP             = errors.New("invalid IP address")
	ErrInvalidPort           = errors.New("invalid port number")
	ErrInvalidHostname       = errors.New("invalid hostname")
	ErrInvalidCountryCode    = errors.New("invalid country code")
	ErrInvalidProtocol       = errors.New("invalid protocol")
	ErrInvalidWalletAddress  = errors.New("invalid wallet address")
	ErrInputTooLong          = errors.New("input exceeds maximum length")
	ErrInputTooShort         = errors.New("input below minimum length")
	ErrInputContainsInvalid  = errors.New("input contains invalid characters")
	ErrSQLInjectionDetected  = errors.New("potential SQL injection detected")
	ErrXSSDetected           = errors.New("potential XSS attack detected")
)

// Validation configuration
const (
	MaxEmailLength       = 254
	MaxUsernameLength    = 64
	MinUsernameLength    = 3
	MaxPasswordLength    = 128
	MinPasswordLength    = 6
	MaxHostnameLength    = 253
	MaxCountryCodeLength = 2
	MaxInputLength       = 10000
)

// Allowed values (whitelists)
var (
	// VPN protocols
	AllowedProtocols = map[string]bool{
		"wireguard": true,
		"openvpn":   true,
		"ikev2":     true,
		"ipsec":     true,
	}

	// ISO 3166-1 alpha-2 country codes (subset)
	AllowedCountryCodes = map[string]bool{
		"US": true, "GB": true, "DE": true, "FR": true, "NL": true,
		"CH": true, "SE": true, "NO": true, "FI": true, "DK": true,
		"CA": true, "AU": true, "JP": true, "SG": true, "HK": true,
		"KR": true, "TW": true, "BR": true, "MX": true, "AR": true,
		"IN": true, "ZA": true, "AE": true, "IL": true, "NZ": true,
		"IE": true, "BE": true, "AT": true, "PL": true, "CZ": true,
		"IT": true, "ES": true, "PT": true, "RO": true, "HU": true,
		// Add more as needed
	}

	// Cryptocurrency types
	AllowedCryptoCurrencies = map[string]bool{
		"ethereum": true,
		"bitcoin":  true,
		"litecoin": true,
		"monero":   true,
	}

	// Subscription tiers
	AllowedSubscriptionTiers = map[string]bool{
		"free":    true,
		"basic":   true,
		"premium": true,
	}
)

// Precompiled regex patterns
var (
	uuidPattern     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	usernamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{2,63}$`)
	hostnamePattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

	// Ethereum address (basic check)
	ethAddressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)

	// Bitcoin addresses (Legacy, SegWit, Bech32)
	btcAddressPattern = regexp.MustCompile(`^(1|3)[a-km-zA-HJ-NP-Z1-9]{25,34}$|^bc1[a-zA-HJ-NP-Z0-9]{39,59}$`)

	// Litecoin addresses
	ltcAddressPattern = regexp.MustCompile(`^[LM3][a-km-zA-HJ-NP-Z1-9]{26,33}$|^ltc1[a-zA-HJ-NP-Z0-9]{39,59}$`)

	// SQL injection patterns
	sqlInjectionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\b(select|insert|update|delete|drop|union|alter|create|truncate|exec|execute)\b)`),
		regexp.MustCompile(`(?i)(--|\;|\/\*|\*\/|@@|char\(|nchar\(|varchar\(|nvarchar\()`),
		regexp.MustCompile(`(?i)(xp_|sp_|0x[0-9a-f]+)`),
	}

	// XSS patterns
	xssPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>.*</script>`),
		regexp.MustCompile(`(?i)<[^>]*(on\w+\s*=)`),
		regexp.MustCompile(`(?i)(javascript:|data:|vbscript:)`),
		regexp.MustCompile(`(?i)<\s*(iframe|object|embed|form|input|button|textarea|select)`),
	}
)

// Validator provides secure input validation
type Validator struct {
	strictMode bool
}

// NewValidator creates a new validator
func NewValidator(strictMode bool) *Validator {
	return &Validator{strictMode: strictMode}
}

// ValidateEmail validates an email address
func (v *Validator) ValidateEmail(email string) error {
	if len(email) > MaxEmailLength {
		return ErrInputTooLong
	}

	if len(email) == 0 {
		return ErrInvalidEmail
	}

	// Parse using net/mail
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}

	// Additional checks
	email = addr.Address
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ErrInvalidEmail
	}

	// Validate domain has valid TLD
	domain := parts[1]
	if !strings.Contains(domain, ".") {
		return ErrInvalidEmail
	}

	return nil
}

// ValidateUsername validates a username
func (v *Validator) ValidateUsername(username string) error {
	if len(username) > MaxUsernameLength {
		return ErrInputTooLong
	}

	if len(username) < MinUsernameLength {
		return ErrInputTooShort
	}

	if !usernamePattern.MatchString(username) {
		return ErrInvalidUsername
	}

	return nil
}

// PasswordStrength represents password strength requirements
type PasswordStrength struct {
	MinLength       int
	RequireUpper    bool
	RequireLower    bool
	RequireDigit    bool
	RequireSpecial  bool
	DisallowCommon  bool
	MaxRepeating    int
}

// DefaultPasswordStrength returns default requirements (simplified for easy setup)
func DefaultPasswordStrength() *PasswordStrength {
	return &PasswordStrength{
		MinLength:      6,
		RequireUpper:   false,
		RequireLower:   false,
		RequireDigit:   false,
		RequireSpecial: false,
		DisallowCommon: true,
		MaxRepeating:   5,
	}
}

// DefaultPasswordPolicy is an alias for DefaultPasswordStrength
func DefaultPasswordPolicy() *PasswordStrength {
	return DefaultPasswordStrength()
}

// HighSecurityPasswordPolicy returns stricter password requirements
func HighSecurityPasswordPolicy() *PasswordStrength {
	return &PasswordStrength{
		MinLength:      16,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
		DisallowCommon: true,
		MaxRepeating:   2,
	}
}

// ValidatePassword validates a password against strength requirements
func (v *Validator) ValidatePassword(password string, strength *PasswordStrength) error {
	if strength == nil {
		strength = DefaultPasswordStrength()
	}

	if len(password) > MaxPasswordLength {
		return fmt.Errorf("%w: maximum %d characters", ErrInputTooLong, MaxPasswordLength)
	}

	if len(password) < strength.MinLength {
		return fmt.Errorf("%w: minimum %d characters required", ErrInvalidPassword, strength.MinLength)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	repeatingCount := 1
	var lastRune rune

	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			hasSpecial = true
		}

		// Check repeating characters
		if r == lastRune {
			repeatingCount++
			if repeatingCount > strength.MaxRepeating {
				return fmt.Errorf("%w: too many repeating characters", ErrInvalidPassword)
			}
		} else {
			repeatingCount = 1
		}
		lastRune = r
	}

	if strength.RequireUpper && !hasUpper {
		return fmt.Errorf("%w: must contain uppercase letter", ErrInvalidPassword)
	}
	if strength.RequireLower && !hasLower {
		return fmt.Errorf("%w: must contain lowercase letter", ErrInvalidPassword)
	}
	if strength.RequireDigit && !hasDigit {
		return fmt.Errorf("%w: must contain digit", ErrInvalidPassword)
	}
	if strength.RequireSpecial && !hasSpecial {
		return fmt.Errorf("%w: must contain special character", ErrInvalidPassword)
	}

	if strength.DisallowCommon && v.isCommonPassword(password) {
		return fmt.Errorf("%w: password is too common", ErrInvalidPassword)
	}

	return nil
}

// Common password check (simplified - in production, use a proper database)
func (v *Validator) isCommonPassword(password string) bool {
	common := map[string]bool{
		"password":     true,
		"123456":       true,
		"123456789":    true,
		"qwerty":       true,
		"abc123":       true,
		"password1":    true,
		"password123":  true,
		"admin":        true,
		"letmein":      true,
		"welcome":      true,
		"monkey":       true,
		"dragon":       true,
		"master":       true,
		"login":        true,
		"trustno1":     true,
	}

	lower := strings.ToLower(password)
	return common[lower]
}

// ValidateUUID validates a UUID string
func (v *Validator) ValidateUUID(id string) error {
	if !uuidPattern.MatchString(id) {
		return ErrInvalidUUID
	}
	return nil
}

// ValidateIP validates an IP address
func (v *Validator) ValidateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return ErrInvalidIP
	}
	return nil
}

// ValidatePort validates a port number
func (v *Validator) ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return ErrInvalidPort
	}
	return nil
}

// ValidatePortString validates a port number string
func (v *Validator) ValidatePortString(portStr string) error {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return ErrInvalidPort
	}
	return v.ValidatePort(port)
}

// ValidateHostname validates a hostname
func (v *Validator) ValidateHostname(hostname string) error {
	if len(hostname) > MaxHostnameLength {
		return ErrInputTooLong
	}

	if !hostnamePattern.MatchString(hostname) {
		return ErrInvalidHostname
	}

	return nil
}

// ValidateCountryCode validates an ISO 3166-1 alpha-2 country code
func (v *Validator) ValidateCountryCode(code string) error {
	if len(code) != 2 {
		return ErrInvalidCountryCode
	}

	code = strings.ToUpper(code)
	if !AllowedCountryCodes[code] {
		return ErrInvalidCountryCode
	}

	return nil
}

// ValidateProtocol validates a VPN protocol
func (v *Validator) ValidateProtocol(protocol string) error {
	protocol = strings.ToLower(protocol)
	if !AllowedProtocols[protocol] {
		return ErrInvalidProtocol
	}
	return nil
}

// ValidateCryptoCurrency validates a cryptocurrency type
func (v *Validator) ValidateCryptoCurrency(currency string) error {
	currency = strings.ToLower(currency)
	if !AllowedCryptoCurrencies[currency] {
		return fmt.Errorf("invalid cryptocurrency: %s", currency)
	}
	return nil
}

// ValidateWalletAddress validates a cryptocurrency wallet address
func (v *Validator) ValidateWalletAddress(address, currency string) error {
	currency = strings.ToLower(currency)

	switch currency {
	case "ethereum":
		if !ethAddressPattern.MatchString(address) {
			return ErrInvalidWalletAddress
		}
		// Validate checksum (EIP-55)
		if !v.validateEthChecksum(address) {
			return fmt.Errorf("%w: invalid checksum", ErrInvalidWalletAddress)
		}

	case "bitcoin":
		if !btcAddressPattern.MatchString(address) {
			return ErrInvalidWalletAddress
		}

	case "litecoin":
		if !ltcAddressPattern.MatchString(address) {
			return ErrInvalidWalletAddress
		}

	default:
		// For unknown currencies, just check basic format
		if len(address) < 26 || len(address) > 100 {
			return ErrInvalidWalletAddress
		}
	}

	return nil
}

// validateEthChecksum validates Ethereum EIP-55 checksum
func (v *Validator) validateEthChecksum(address string) bool {
	if len(address) != 42 {
		return false
	}

	// Remove 0x prefix
	addr := strings.ToLower(address[2:])

	// Hash the lowercase address
	hash := sha256.Sum256([]byte(addr))
	hashHex := hex.EncodeToString(hash[:])

	// Compare each character
	for i := 0; i < 40; i++ {
		c := address[i+2]

		// If it's a letter
		if c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' {
			// Get the corresponding hash nibble
			hashNibble := hashHex[i]

			// Should be uppercase if hash nibble >= 8
			shouldBeUpper := hashNibble >= '8'

			if shouldBeUpper && c >= 'a' && c <= 'f' {
				return false
			}
			if !shouldBeUpper && c >= 'A' && c <= 'F' {
				return false
			}
		}
	}

	return true
}

// DetectSQLInjection checks for potential SQL injection patterns
func (v *Validator) DetectSQLInjection(input string) bool {
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// DetectXSS checks for potential XSS attack patterns
func (v *Validator) DetectXSS(input string) bool {
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// SanitizeString removes potentially dangerous characters
func (v *Validator) SanitizeString(input string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = MaxInputLength
	}

	// Limit length
	if len(input) > maxLength {
		input = input[:maxLength]
	}

	// Ensure valid UTF-8
	if !utf8.ValidString(input) {
		input = strings.ToValidUTF8(input, "")
	}

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newline and tab
	result := strings.Builder{}
	for _, r := range input {
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// SanitizeHTML escapes HTML special characters
func (v *Validator) SanitizeHTML(input string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#x27;",
		"/", "&#x2F;",
	)
	return replacer.Replace(input)
}

// ValidateAndSanitize validates and sanitizes input
func (v *Validator) ValidateAndSanitize(input string, maxLength int) (string, error) {
	if len(input) > maxLength {
		return "", ErrInputTooLong
	}

	sanitized := v.SanitizeString(input, maxLength)

	if v.strictMode {
		if v.DetectSQLInjection(sanitized) {
			return "", ErrSQLInjectionDetected
		}
		if v.DetectXSS(sanitized) {
			return "", ErrXSSDetected
		}
	}

	return sanitized, nil
}

// ValidatePagination validates pagination parameters
func (v *Validator) ValidatePagination(limit, offset int, maxLimit int) (int, int, error) {
	if maxLimit <= 0 {
		maxLimit = 100
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	if offset < 0 {
		offset = 0
	}

	return limit, offset, nil
}

// HashForLog creates a truncated hash suitable for logging (no PII)
func HashForLog(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:4]) // Only first 4 bytes
}

// Default validator for convenience functions
var defaultValidator = NewValidator(true)

// Convenience functions using default validator

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	return defaultValidator.ValidateEmail(email)
}

// ValidateUsername validates a username
func ValidateUsername(username string) error {
	return defaultValidator.ValidateUsername(username)
}

// ValidatePassword validates a password with default strength requirements
func ValidatePassword(password string) error {
	return defaultValidator.ValidatePassword(password, nil)
}

// ValidateUUID validates a UUID string
func ValidateUUID(id string) error {
	return defaultValidator.ValidateUUID(id)
}

// ValidateIP validates an IP address
func ValidateIP(ip string) error {
	return defaultValidator.ValidateIP(ip)
}

// ValidatePort validates a port number
func ValidatePort(port int) error {
	return defaultValidator.ValidatePort(port)
}

// ValidateHostname validates a hostname
func ValidateHostname(hostname string) error {
	return defaultValidator.ValidateHostname(hostname)
}

// ValidateCountryCode validates an ISO 3166-1 alpha-2 country code
func ValidateCountryCode(code string) error {
	return defaultValidator.ValidateCountryCode(code)
}

// ValidateProtocol validates a VPN protocol
func ValidateProtocol(protocol string) error {
	return defaultValidator.ValidateProtocol(protocol)
}

// ValidateWalletAddress validates a cryptocurrency wallet address
func ValidateWalletAddress(address, currency string) error {
	return defaultValidator.ValidateWalletAddress(address, currency)
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(input string, maxLength int) string {
	return defaultValidator.SanitizeString(input, maxLength)
}

// SanitizeHTML escapes HTML special characters
func SanitizeHTML(input string) string {
	return defaultValidator.SanitizeHTML(input)
}

// DetectSQLInjection checks for potential SQL injection patterns
func DetectSQLInjection(input string) bool {
	return defaultValidator.DetectSQLInjection(input)
}

// DetectXSS checks for potential XSS attack patterns
func DetectXSS(input string) bool {
	return defaultValidator.DetectXSS(input)
}
