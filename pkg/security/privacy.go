// Package security provides privacy-preserving utilities for zero-knowledge operations
package security

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Privacy-related constants
const (
	// Hash prefixes for different data types
	HashPrefixIP       = "ip:"
	HashPrefixEmail    = "email:"
	HashPrefixSession  = "sess:"
	HashPrefixDevice   = "dev:"

	// Retention periods
	DefaultLogRetention     = 24 * time.Hour
	MinimalMetadataRetention = 1 * time.Hour
)

// PrivacyConfig holds privacy-related configuration
type PrivacyConfig struct {
	// Logging
	RedactIPAddresses   bool
	RedactEmails        bool
	RedactUserIDs       bool
	HashSensitiveData   bool
	MinimalLogging      bool

	// Data retention
	LogRetentionPeriod  time.Duration
	SessionRetention    time.Duration

	// Anonymization
	AnonymizationSalt   []byte
}

// DefaultPrivacyConfig returns privacy-preserving defaults
func DefaultPrivacyConfig() *PrivacyConfig {
	salt, _ := SecureBytes(32)
	return &PrivacyConfig{
		RedactIPAddresses:   true,
		RedactEmails:        true,
		RedactUserIDs:       false, // UUIDs are generally safe
		HashSensitiveData:   true,
		MinimalLogging:      true,
		LogRetentionPeriod:  24 * time.Hour,
		SessionRetention:    7 * 24 * time.Hour,
		AnonymizationSalt:   salt,
	}
}

// PrivacyFilter provides privacy-preserving data filtering
type PrivacyFilter struct {
	config *PrivacyConfig
	mu     sync.RWMutex
}

// NewPrivacyFilter creates a new privacy filter
func NewPrivacyFilter(config *PrivacyConfig) *PrivacyFilter {
	if config == nil {
		config = DefaultPrivacyConfig()
	}
	return &PrivacyFilter{config: config}
}

// RedactIP redacts an IP address for logging
func (pf *PrivacyFilter) RedactIP(ip string) string {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	if !pf.config.RedactIPAddresses {
		return ip
	}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "[invalid-ip]"
	}

	if pf.config.HashSensitiveData {
		return pf.hashValue(HashPrefixIP, ip)
	}

	// Partial redaction: keep first two octets for IPv4, first 2 blocks for IPv6
	if parsed.To4() != nil {
		parts := strings.Split(ip, ".")
		if len(parts) >= 2 {
			return parts[0] + "." + parts[1] + ".x.x"
		}
	} else {
		parts := strings.Split(ip, ":")
		if len(parts) >= 2 {
			return parts[0] + ":" + parts[1] + ":xxxx:xxxx:xxxx:xxxx:xxxx:xxxx"
		}
	}

	return "[redacted]"
}

// RedactEmail redacts an email address for logging
func (pf *PrivacyFilter) RedactEmail(email string) string {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	if !pf.config.RedactEmails {
		return email
	}

	if pf.config.HashSensitiveData {
		return pf.hashValue(HashPrefixEmail, email)
	}

	// Partial redaction: show first char and domain
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "[redacted]"
	}

	local := parts[0]
	domain := parts[1]

	if len(local) > 1 {
		local = string(local[0]) + "***"
	}

	return local + "@" + domain
}

// RedactUserID redacts a user ID for logging
func (pf *PrivacyFilter) RedactUserID(userID string) string {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	if !pf.config.RedactUserIDs {
		return userID
	}

	if pf.config.HashSensitiveData {
		return pf.hashValue("uid:", userID)
	}

	// Show only first 8 chars of UUID
	if len(userID) > 8 {
		return userID[:8] + "..."
	}

	return "[redacted]"
}

// hashValue creates a privacy-preserving hash
func (pf *PrivacyFilter) hashValue(prefix, value string) string {
	data := append(pf.config.AnonymizationSalt, []byte(value)...)
	hash := sha256.Sum256(data)
	return prefix + hex.EncodeToString(hash[:8])
}

// AnonymizeIP creates a consistent anonymous identifier for an IP
// This allows correlation without revealing the actual IP
func (pf *PrivacyFilter) AnonymizeIP(ip string) string {
	return pf.hashValue(HashPrefixIP, ip)
}

// AnonymizeSession creates a consistent anonymous session identifier
func (pf *PrivacyFilter) AnonymizeSession(sessionID string) string {
	return pf.hashValue(HashPrefixSession, sessionID)
}

// AnonymizeDevice creates a consistent anonymous device identifier
func (pf *PrivacyFilter) AnonymizeDevice(fingerprint string) string {
	return pf.hashValue(HashPrefixDevice, fingerprint)
}

// SanitizeLogMessage removes sensitive data from log messages
func (pf *PrivacyFilter) SanitizeLogMessage(message string) string {
	// Patterns to redact
	patterns := []struct {
		regex       *regexp.Regexp
		replacement string
	}{
		// IP addresses (IPv4)
		{
			regex:       regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
			replacement: "[ip-redacted]",
		},
		// IP addresses (IPv6)
		{
			regex:       regexp.MustCompile(`\b([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`),
			replacement: "[ipv6-redacted]",
		},
		// Email addresses
		{
			regex:       regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
			replacement: "[email-redacted]",
		},
		// JWT tokens (long base64 strings with dots)
		{
			regex:       regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`),
			replacement: "[token-redacted]",
		},
		// API keys (common patterns)
		{
			regex:       regexp.MustCompile(`(?i)(api[_-]?key|apikey|access[_-]?token|secret)[=:]["']?[A-Za-z0-9_-]{20,}["']?`),
			replacement: "[credential-redacted]",
		},
		// Private keys
		{
			regex:       regexp.MustCompile(`-----BEGIN [A-Z ]+ PRIVATE KEY-----[\s\S]*?-----END [A-Z ]+ PRIVATE KEY-----`),
			replacement: "[private-key-redacted]",
		},
		// Passwords in URLs
		{
			regex:       regexp.MustCompile(`(?i)(password|passwd|pwd)[=:][^&\s]+`),
			replacement: "[password-redacted]",
		},
		// Credit card numbers
		{
			regex:       regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
			replacement: "[card-redacted]",
		},
		// Ethereum addresses
		{
			regex:       regexp.MustCompile(`0x[a-fA-F0-9]{40}`),
			replacement: "[eth-address-redacted]",
		},
	}

	result := message
	for _, p := range patterns {
		result = p.regex.ReplaceAllString(result, p.replacement)
	}

	return result
}

// LogEntry represents a privacy-safe log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"ts"`
	Level       string                 `json:"level"`
	Message     string                 `json:"msg"`
	Component   string                 `json:"component,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserHash    string                 `json:"user_hash,omitempty"` // Hashed user ID
	SessionHash string                 `json:"session_hash,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Success     bool                   `json:"success,omitempty"`
	ErrorCode   string                 `json:"error_code,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// PrivacyLogger wraps logging with privacy filtering
type PrivacyLogger struct {
	filter    *PrivacyFilter
	component string
}

// NewPrivacyLogger creates a new privacy-aware logger
func NewPrivacyLogger(filter *PrivacyFilter, component string) *PrivacyLogger {
	return &PrivacyLogger{
		filter:    filter,
		component: component,
	}
}

// CreateLogEntry creates a privacy-safe log entry
func (pl *PrivacyLogger) CreateLogEntry(
	level, message string,
	userID, sessionID, requestID string,
	action string,
	success bool,
	errorCode string,
	duration time.Duration,
) *LogEntry {
	entry := &LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Message:   pl.filter.SanitizeLogMessage(message),
		Component: pl.component,
		RequestID: requestID,
		Action:    action,
		Success:   success,
		Duration:  duration,
	}

	if errorCode != "" && !success {
		entry.ErrorCode = errorCode
	}

	if userID != "" {
		entry.UserHash = pl.filter.AnonymizeIP(userID) // Use IP hash for user
	}

	if sessionID != "" {
		entry.SessionHash = pl.filter.AnonymizeSession(sessionID)
	}

	return entry
}

// MinimalSessionData represents minimal session metadata for privacy
type MinimalSessionData struct {
	SessionHash     string    `json:"session_hash"`
	NodeHash        string    `json:"node_hash"`
	ProtocolHash    string    `json:"protocol_hash"`
	ConnectedAt     time.Time `json:"connected_at"`
	DisconnectedAt  *time.Time `json:"disconnected_at,omitempty"`
	BytesTransferred int64    `json:"bytes_transferred"` // Total only, no breakdown
}

// CreateMinimalSessionData creates privacy-preserving session data
func (pf *PrivacyFilter) CreateMinimalSessionData(
	sessionID, nodeID, protocol string,
	connectedAt time.Time,
	disconnectedAt *time.Time,
	bytesIn, bytesOut int64,
) *MinimalSessionData {
	return &MinimalSessionData{
		SessionHash:      pf.AnonymizeSession(sessionID),
		NodeHash:         pf.hashValue("node:", nodeID),
		ProtocolHash:     pf.hashValue("proto:", protocol),
		ConnectedAt:      connectedAt.Truncate(time.Minute), // Reduce time precision
		DisconnectedAt:   disconnectedAt,
		BytesTransferred: bytesIn + bytesOut, // Aggregate only
	}
}

// DataRetentionPolicy defines when data should be purged
type DataRetentionPolicy struct {
	// Session data
	ActiveSessionRetention   time.Duration
	CompletedSessionRetention time.Duration

	// Logs
	SecurityLogRetention     time.Duration
	AccessLogRetention       time.Duration
	ErrorLogRetention        time.Duration

	// User data
	InactiveUserRetention    time.Duration
	DeletedUserRetention     time.Duration

	// Payment data
	PaymentRecordRetention   time.Duration
}

// DefaultDataRetentionPolicy returns GDPR-compliant defaults
func DefaultDataRetentionPolicy() *DataRetentionPolicy {
	return &DataRetentionPolicy{
		ActiveSessionRetention:    0, // Keep while active
		CompletedSessionRetention: 24 * time.Hour,
		SecurityLogRetention:      90 * 24 * time.Hour, // 90 days
		AccessLogRetention:        7 * 24 * time.Hour,  // 7 days
		ErrorLogRetention:         30 * 24 * time.Hour, // 30 days
		InactiveUserRetention:     365 * 24 * time.Hour, // 1 year
		DeletedUserRetention:      30 * 24 * time.Hour,  // 30 days after deletion request
		PaymentRecordRetention:    7 * 365 * 24 * time.Hour, // 7 years (legal requirement)
	}
}

// WarrantCanary represents a warrant canary status
type WarrantCanary struct {
	LastUpdated     time.Time `json:"last_updated"`
	Statement       string    `json:"statement"`
	NoGagOrders     bool      `json:"no_gag_orders"`
	NoNSLs          bool      `json:"no_nsls"`
	NoCourtOrders   bool      `json:"no_court_orders"`
	NoGovernmentReq bool      `json:"no_government_requests"`
	Signature       string    `json:"signature"`
}

// CreateWarrantCanary creates a new warrant canary statement
func CreateWarrantCanary(privateKey []byte) (*WarrantCanary, error) {
	canary := &WarrantCanary{
		LastUpdated:     time.Now().UTC(),
		Statement:       "As of the date above, Aureo VPN has not received any national security letters, FISA orders, or any other classified requests for user information.",
		NoGagOrders:     true,
		NoNSLs:          true,
		NoCourtOrders:   true,
		NoGovernmentReq: true,
	}

	// Sign the canary (simplified - in production, use proper signing)
	data := []byte(canary.Statement + canary.LastUpdated.String())
	hash := sha256.Sum256(append(privateKey, data...))
	canary.Signature = hex.EncodeToString(hash[:])

	return canary, nil
}
