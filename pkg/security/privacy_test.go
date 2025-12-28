package security

import (
	"strings"
	"testing"
	"time"
)

// =============================================================================
// Privacy Filter Tests
// =============================================================================

func TestNewPrivacyFilter(t *testing.T) {
	// With nil config should use defaults
	pf := NewPrivacyFilter(nil)
	if pf == nil {
		t.Fatal("NewPrivacyFilter(nil) returned nil")
	}
}

func TestRedactIP(t *testing.T) {
	pf := NewPrivacyFilter(nil)

	tests := []struct {
		name string
		ip   string
		want string // For hashing, we just check format
	}{
		{"IPv4", "192.168.1.100", "ip:"},
		{"IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "ip:"},
		{"invalid", "not-an-ip", "[invalid-ip]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pf.RedactIP(tt.ip)
			if !strings.HasPrefix(got, tt.want) {
				t.Errorf("RedactIP(%q) = %q, want prefix %q", tt.ip, got, tt.want)
			}
		})
	}
}

func TestRedactIPNoHash(t *testing.T) {
	config := DefaultPrivacyConfig()
	config.HashSensitiveData = false
	pf := NewPrivacyFilter(config)

	// IPv4 should show first two octets
	got := pf.RedactIP("192.168.1.100")
	if got != "192.168.x.x" {
		t.Errorf("RedactIP() = %q, want 192.168.x.x", got)
	}
}

func TestRedactIPDisabled(t *testing.T) {
	config := DefaultPrivacyConfig()
	config.RedactIPAddresses = false
	pf := NewPrivacyFilter(config)

	ip := "192.168.1.100"
	got := pf.RedactIP(ip)
	if got != ip {
		t.Errorf("RedactIP() with redaction disabled = %q, want %q", got, ip)
	}
}

func TestRedactEmail(t *testing.T) {
	pf := NewPrivacyFilter(nil)

	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"valid email", "test@example.com", "email:"},
		{"invalid email", "not-an-email", "email:"}, // Still hashes even if invalid format
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pf.RedactEmail(tt.email)
			if !strings.HasPrefix(got, tt.want) {
				t.Errorf("RedactEmail(%q) = %q, want prefix %q", tt.email, got, tt.want)
			}
		})
	}
}

func TestRedactEmailNoHash(t *testing.T) {
	config := DefaultPrivacyConfig()
	config.HashSensitiveData = false
	pf := NewPrivacyFilter(config)

	got := pf.RedactEmail("testuser@example.com")
	if got != "t***@example.com" {
		t.Errorf("RedactEmail() = %q, want t***@example.com", got)
	}
}

func TestRedactUserID(t *testing.T) {
	config := DefaultPrivacyConfig()
	config.RedactUserIDs = true
	config.HashSensitiveData = true
	pf := NewPrivacyFilter(config)

	got := pf.RedactUserID("550e8400-e29b-41d4-a716-446655440000")
	if !strings.HasPrefix(got, "uid:") {
		t.Errorf("RedactUserID() = %q, want prefix uid:", got)
	}
}

func TestRedactUserIDDisabled(t *testing.T) {
	config := DefaultPrivacyConfig()
	config.RedactUserIDs = false
	pf := NewPrivacyFilter(config)

	userID := "550e8400-e29b-41d4-a716-446655440000"
	got := pf.RedactUserID(userID)
	if got != userID {
		t.Errorf("RedactUserID() with redaction disabled = %q, want %q", got, userID)
	}
}

// =============================================================================
// Anonymization Tests
// =============================================================================

func TestAnonymizeIP(t *testing.T) {
	pf := NewPrivacyFilter(nil)

	ip1 := pf.AnonymizeIP("192.168.1.1")
	ip2 := pf.AnonymizeIP("192.168.1.1")
	ip3 := pf.AnonymizeIP("192.168.1.2")

	// Same IP should produce same hash
	if ip1 != ip2 {
		t.Error("AnonymizeIP() should produce consistent hashes")
	}

	// Different IPs should produce different hashes
	if ip1 == ip3 {
		t.Error("AnonymizeIP() should produce different hashes for different IPs")
	}

	// Should have expected prefix
	if !strings.HasPrefix(ip1, HashPrefixIP) {
		t.Errorf("AnonymizeIP() = %q, want prefix %q", ip1, HashPrefixIP)
	}
}

func TestAnonymizeSession(t *testing.T) {
	pf := NewPrivacyFilter(nil)

	session := pf.AnonymizeSession("session-123")

	if !strings.HasPrefix(session, HashPrefixSession) {
		t.Errorf("AnonymizeSession() = %q, want prefix %q", session, HashPrefixSession)
	}
}

func TestAnonymizeDevice(t *testing.T) {
	pf := NewPrivacyFilter(nil)

	device := pf.AnonymizeDevice("device-fingerprint")

	if !strings.HasPrefix(device, HashPrefixDevice) {
		t.Errorf("AnonymizeDevice() = %q, want prefix %q", device, HashPrefixDevice)
	}
}

// =============================================================================
// Log Sanitization Tests
// =============================================================================

func TestSanitizeLogMessage(t *testing.T) {
	pf := NewPrivacyFilter(nil)

	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			"IPv4 address",
			"Connection from 192.168.1.100 failed",
			"Connection from [ip-redacted] failed",
		},
		{
			"email address",
			"User user@example.com logged in",
			"User [email-redacted] logged in",
		},
		{
			"JWT token",
			"Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			"Token: [token-redacted]",
		},
		{
			"Ethereum address",
			"Sent to 0x742d35Cc6634C0532925a3b844Bc9e7595f5c123",
			"Sent to [eth-address-redacted]",
		},
		{
			"password in URL",
			"Connecting to postgres://user:password=secret@localhost/db",
			"Connecting to postgres://user:[password-redacted]",
		},
		{
			"no sensitive data",
			"Server started successfully",
			"Server started successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pf.SanitizeLogMessage(tt.message)
			if got != tt.want {
				t.Errorf("SanitizeLogMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// Privacy Logger Tests
// =============================================================================

func TestPrivacyLogger(t *testing.T) {
	pf := NewPrivacyFilter(nil)
	logger := NewPrivacyLogger(pf, "test-component")

	entry := logger.CreateLogEntry(
		"INFO",
		"User logged in from 192.168.1.1",
		"user-123",
		"session-456",
		"req-789",
		"login",
		true,
		"",
		100*time.Millisecond,
	)

	if entry.Level != "INFO" {
		t.Errorf("LogEntry.Level = %q, want INFO", entry.Level)
	}

	if entry.Component != "test-component" {
		t.Errorf("LogEntry.Component = %q, want test-component", entry.Component)
	}

	// IP should be redacted in message
	if strings.Contains(entry.Message, "192.168.1.1") {
		t.Error("LogEntry.Message should not contain raw IP")
	}

	// UserHash should be set
	if entry.UserHash == "" {
		t.Error("LogEntry.UserHash should be set")
	}

	// SessionHash should be set
	if entry.SessionHash == "" {
		t.Error("LogEntry.SessionHash should be set")
	}
}

// =============================================================================
// Minimal Session Data Tests
// =============================================================================

func TestCreateMinimalSessionData(t *testing.T) {
	pf := NewPrivacyFilter(nil)

	connectedAt := time.Now().Add(-1 * time.Hour)
	disconnectedAt := time.Now()

	data := pf.CreateMinimalSessionData(
		"session-123",
		"node-456",
		"wireguard",
		connectedAt,
		&disconnectedAt,
		1024*1024, // 1 MB in
		512*1024,  // 512 KB out
	)

	// All IDs should be hashed
	if !strings.HasPrefix(data.SessionHash, HashPrefixSession) {
		t.Errorf("SessionHash = %q, want prefix %q", data.SessionHash, HashPrefixSession)
	}

	// Bytes should be aggregated
	expectedBytes := int64(1024*1024 + 512*1024)
	if data.BytesTransferred != expectedBytes {
		t.Errorf("BytesTransferred = %d, want %d", data.BytesTransferred, expectedBytes)
	}

	// Time should be truncated to minute
	if data.ConnectedAt.Second() != 0 || data.ConnectedAt.Nanosecond() != 0 {
		t.Error("ConnectedAt should be truncated to minute precision")
	}
}

// =============================================================================
// Data Retention Policy Tests
// =============================================================================

func TestDefaultDataRetentionPolicy(t *testing.T) {
	policy := DefaultDataRetentionPolicy()

	if policy.AccessLogRetention != 7*24*time.Hour {
		t.Errorf("AccessLogRetention = %v, want 7 days", policy.AccessLogRetention)
	}

	if policy.SecurityLogRetention != 90*24*time.Hour {
		t.Errorf("SecurityLogRetention = %v, want 90 days", policy.SecurityLogRetention)
	}

	if policy.PaymentRecordRetention != 7*365*24*time.Hour {
		t.Errorf("PaymentRecordRetention = %v, want 7 years", policy.PaymentRecordRetention)
	}
}

// =============================================================================
// Warrant Canary Tests
// =============================================================================

func TestCreateWarrantCanary(t *testing.T) {
	privateKey, _ := SecureBytes(32)

	canary, err := CreateWarrantCanary(privateKey)
	if err != nil {
		t.Fatalf("CreateWarrantCanary() error = %v", err)
	}

	if canary.Statement == "" {
		t.Error("WarrantCanary.Statement is empty")
	}

	if !canary.NoGagOrders {
		t.Error("WarrantCanary.NoGagOrders should be true")
	}

	if !canary.NoNSLs {
		t.Error("WarrantCanary.NoNSLs should be true")
	}

	if canary.Signature == "" {
		t.Error("WarrantCanary.Signature is empty")
	}

	if canary.LastUpdated.IsZero() {
		t.Error("WarrantCanary.LastUpdated is zero")
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkRedactIP(b *testing.B) {
	pf := NewPrivacyFilter(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.RedactIP("192.168.1.100")
	}
}

func BenchmarkSanitizeLogMessage(b *testing.B) {
	pf := NewPrivacyFilter(nil)
	msg := "User user@example.com connected from 192.168.1.1 with token eyJhbGci..."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.SanitizeLogMessage(msg)
	}
}

func BenchmarkAnonymizeIP(b *testing.B) {
	pf := NewPrivacyFilter(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.AnonymizeIP("192.168.1.100")
	}
}
