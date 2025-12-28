package security

import (
	"bufio"
	"context"
	"net"
	"strings"
	"sync"
	"testing"
)

// =============================================================================
// NetShield Tests
// =============================================================================

func TestNewNetShield(t *testing.T) {
	// With nil config (uses defaults)
	ns := NewNetShield(nil)
	if ns == nil {
		t.Fatal("NewNetShield(nil) returned nil")
	}

	if ns.config.Level != NetShieldLevel2 {
		t.Errorf("Default level = %d, want %d", ns.config.Level, NetShieldLevel2)
	}

	// With custom config
	config := MaxSecurityConfig()
	ns2 := NewNetShield(config)
	if ns2.config.Level != NetShieldLevel3 {
		t.Errorf("MaxSecurity level = %d, want %d", ns2.config.Level, NetShieldLevel3)
	}
}

func TestNetShieldCheckMalware(t *testing.T) {
	ns := NewNetShield(nil)

	// Check known malware domain
	result := ns.CheckDomain("malware.com")
	if !result.Blocked {
		t.Error("malware.com should be blocked")
	}
	if result.Category != ThreatMalware {
		t.Errorf("Category = %v, want ThreatMalware", result.Category)
	}

	// Check subdomain
	result = ns.CheckDomain("sub.malware.com")
	if !result.Blocked {
		t.Error("sub.malware.com should be blocked")
	}
}

func TestNetShieldCheckPhishing(t *testing.T) {
	ns := NewNetShield(nil)

	result := ns.CheckDomain("phishing.com")
	if !result.Blocked {
		t.Error("phishing.com should be blocked")
	}
	if result.Category != ThreatPhishing {
		t.Errorf("Category = %v, want ThreatPhishing", result.Category)
	}
}

func TestNetShieldCheckAds(t *testing.T) {
	config := DefaultNetShieldConfig()
	ns := NewNetShield(config)

	// Ad domain
	result := ns.CheckDomain("doubleclick.net")
	if !result.Blocked {
		t.Error("doubleclick.net should be blocked at Level 2")
	}
	if result.Category != ThreatAdware {
		t.Errorf("Category = %v, want ThreatAdware", result.Category)
	}

	// Pattern match
	result = ns.CheckDomain("ads.example.com")
	if !result.Blocked {
		t.Error("ads.example.com should be blocked by pattern")
	}
}

func TestNetShieldCheckTrackers(t *testing.T) {
	ns := NewNetShield(nil)

	result := ns.CheckDomain("google-analytics.com")
	if !result.Blocked {
		t.Error("google-analytics.com should be blocked")
	}
	if result.Category != ThreatTracker {
		t.Errorf("Category = %v, want ThreatTracker", result.Category)
	}

	// Pattern match
	result = ns.CheckDomain("tracking.example.com")
	if !result.Blocked {
		t.Error("tracking.example.com should be blocked by pattern")
	}
}

func TestNetShieldCheckCryptojacking(t *testing.T) {
	ns := NewNetShield(nil)

	result := ns.CheckDomain("coinhive.com")
	if !result.Blocked {
		t.Error("coinhive.com should be blocked")
	}
	if result.Category != ThreatCryptojacking {
		t.Errorf("Category = %v, want ThreatCryptojacking", result.Category)
	}
}

func TestNetShieldLevelFiltering(t *testing.T) {
	// Level 1 should only block malware, not ads
	config := DefaultNetShieldConfig()
	config.Level = NetShieldLevel1
	config.BlockAds = false
	config.BlockTrackers = false
	ns := NewNetShield(config)

	// Malware should still be blocked
	result := ns.CheckDomain("malware.com")
	if !result.Blocked {
		t.Error("Malware should be blocked at Level 1")
	}

	// Ads should not be blocked at Level 1
	result = ns.CheckDomain("doubleclick.net")
	if result.Blocked {
		t.Error("Ads should not be blocked at Level 1")
	}
}

func TestNetShieldCustomLists(t *testing.T) {
	ns := NewNetShield(nil)

	// Add custom block
	ns.AddBlockDomain("blocked-by-user.com")

	result := ns.CheckDomain("blocked-by-user.com")
	if !result.Blocked {
		t.Error("Custom blocked domain should be blocked")
	}

	// Add to allowlist
	ns.AddAllowDomain("safe-site.com")

	result = ns.CheckDomain("safe-site.com")
	if result.Blocked {
		t.Error("Allowed domain should not be blocked")
	}

	// Remove from blocklist
	ns.RemoveBlockDomain("blocked-by-user.com")
	result = ns.CheckDomain("blocked-by-user.com")
	if result.Blocked {
		t.Error("Removed domain should no longer be blocked")
	}
}

func TestNetShieldAllowlistPriority(t *testing.T) {
	ns := NewNetShield(nil)

	// Add a malware domain to allowlist
	ns.AddAllowDomain("malware.com")

	result := ns.CheckDomain("malware.com")
	if result.Blocked {
		t.Error("Allowlist should take priority over blocklist")
	}
}

func TestNetShieldNormalization(t *testing.T) {
	ns := NewNetShield(nil)

	testCases := []struct {
		input    string
		expected string
	}{
		{"MALWARE.COM", "malware.com"},
		{"http://malware.com", "malware.com"},
		{"https://malware.com", "malware.com"},
		{"www.malware.com", "malware.com"},
		{"malware.com:443", "malware.com"},
		{"malware.com/path", "malware.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := ns.CheckDomain(tc.input)
			if !result.Blocked {
				t.Errorf("%s should be blocked", tc.input)
			}
		})
	}
}

func TestNetShieldStats(t *testing.T) {
	ns := NewNetShield(nil)

	// Make some requests
	ns.CheckDomain("malware.com")
	ns.CheckDomain("safe-site.com")
	ns.CheckDomain("doubleclick.net")

	stats := ns.Stats()

	if stats.TotalRequests != 3 {
		t.Errorf("TotalRequests = %d, want 3", stats.TotalRequests)
	}

	if stats.BlockedRequests != 2 {
		t.Errorf("BlockedRequests = %d, want 2", stats.BlockedRequests)
	}

	if stats.MalwareBlocked != 1 {
		t.Errorf("MalwareBlocked = %d, want 1", stats.MalwareBlocked)
	}
}

func TestNetShieldSetLevel(t *testing.T) {
	ns := NewNetShield(nil)

	ns.SetLevel(NetShieldLevel1)
	if ns.config.Level != NetShieldLevel1 {
		t.Error("SetLevel didn't update level")
	}

	ns.SetLevel(NetShieldOff)
	if ns.config.BlockMalware {
		t.Error("Level Off should disable all blocking")
	}

	ns.SetLevel(NetShieldLevel3)
	if !ns.config.BlockAdultContent {
		t.Error("Level 3 should enable adult content blocking")
	}
}

func TestNetShieldCheckIP(t *testing.T) {
	ns := NewNetShield(nil)

	// Add a malware IP
	ns.malwareIPs.Add(net.ParseIP("192.168.1.100"))

	result := ns.CheckIP(net.ParseIP("192.168.1.100"))
	if !result.Blocked {
		t.Error("Malware IP should be blocked")
	}

	// Safe IP
	result = ns.CheckIP(net.ParseIP("8.8.8.8"))
	if result.Blocked {
		t.Error("Safe IP should not be blocked")
	}
}

func TestNetShieldCheckURL(t *testing.T) {
	ns := NewNetShield(nil)

	result := ns.CheckURL("https://malware.com/path?query=value")
	if !result.Blocked {
		t.Error("Malware URL should be blocked")
	}

	result = ns.CheckURL("https://safe-site.com/page")
	if result.Blocked {
		t.Error("Safe URL should not be blocked")
	}
}

func TestNetShieldLoadBlocklist(t *testing.T) {
	ns := NewNetShield(nil)

	blocklist := `
# Comment line
example-bad-domain1.com
example-bad-domain2.com
0.0.0.0 example-bad-domain3.com
127.0.0.1 example-bad-domain4.com
`

	reader := bufio.NewReader(strings.NewReader(blocklist))
	count, err := ns.LoadBlocklistFromReader(reader, ThreatMalware)

	if err != nil {
		t.Fatalf("LoadBlocklistFromReader error = %v", err)
	}

	if count != 4 {
		t.Errorf("Count = %d, want 4", count)
	}

	// Verify domains are blocked
	for _, domain := range []string{
		"example-bad-domain1.com",
		"example-bad-domain2.com",
		"example-bad-domain3.com",
		"example-bad-domain4.com",
	} {
		result := ns.CheckDomain(domain)
		if !result.Blocked {
			t.Errorf("%s should be blocked", domain)
		}
	}
}

func TestNetShieldConcurrent(t *testing.T) {
	ns := NewNetShield(nil)

	var wg sync.WaitGroup
	numGoroutines := 10
	checksPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < checksPerGoroutine; j++ {
				ns.CheckDomain("malware.com")
				ns.CheckDomain("safe-site.com")
			}
		}()
	}

	wg.Wait()

	stats := ns.Stats()
	expected := uint64(numGoroutines * checksPerGoroutine * 2)
	if stats.TotalRequests != expected {
		t.Errorf("TotalRequests = %d, want %d", stats.TotalRequests, expected)
	}
}

// =============================================================================
// DomainSet Tests
// =============================================================================

func TestDomainSet(t *testing.T) {
	ds := NewDomainSet()

	ds.Add("example.com")
	ds.Add("test.com")

	if !ds.Contains("example.com") {
		t.Error("Should contain example.com")
	}

	if ds.Contains("other.com") {
		t.Error("Should not contain other.com")
	}

	if ds.Size() != 2 {
		t.Errorf("Size = %d, want 2", ds.Size())
	}

	ds.Remove("example.com")
	if ds.Contains("example.com") {
		t.Error("Should not contain removed domain")
	}
}

func TestDomainSetContainsSuffix(t *testing.T) {
	ds := NewDomainSet()
	ds.Add("example.com")

	// Subdomain should match
	if !ds.ContainsSuffix("sub.example.com") {
		t.Error("Should match subdomain")
	}

	if !ds.ContainsSuffix("deep.sub.example.com") {
		t.Error("Should match deep subdomain")
	}

	if ds.ContainsSuffix("other.com") {
		t.Error("Should not match unrelated domain")
	}
}

// =============================================================================
// IPSet Tests
// =============================================================================

func TestIPSet(t *testing.T) {
	is := NewIPSet()

	ip := net.ParseIP("192.168.1.1")
	is.Add(ip)

	if !is.Contains(ip) {
		t.Error("Should contain added IP")
	}

	if is.Contains(net.ParseIP("192.168.1.2")) {
		t.Error("Should not contain non-added IP")
	}
}

func TestIPSetCIDR(t *testing.T) {
	is := NewIPSet()

	err := is.AddCIDR("10.0.0.0/8")
	if err != nil {
		t.Fatalf("AddCIDR error = %v", err)
	}

	if !is.Contains(net.ParseIP("10.1.2.3")) {
		t.Error("Should contain IP in CIDR range")
	}

	if is.Contains(net.ParseIP("192.168.1.1")) {
		t.Error("Should not contain IP outside CIDR range")
	}
}

// =============================================================================
// DNSFilter Tests
// =============================================================================

func TestDNSFilter(t *testing.T) {
	ns := NewNetShield(nil)
	df := NewDNSFilter(ns)

	// Filter blocked domain
	result := df.Filter("malware.com")
	if !result.Blocked {
		t.Error("Malware domain should be blocked")
	}

	// Filter safe domain
	result = df.Filter("safe-site.com")
	if result.Blocked {
		t.Error("Safe domain should not be blocked")
	}
}

func TestDNSFilterCaching(t *testing.T) {
	ns := NewNetShield(nil)
	df := NewDNSFilter(ns)

	// First call - not cached
	result1 := df.Filter("test-cache.com")

	// Second call - should be cached
	result2 := df.Filter("test-cache.com")

	// Results should be the same
	if result1.Blocked != result2.Blocked {
		t.Error("Cached result should match")
	}
}

func TestDNSFilterResolveBlocked(t *testing.T) {
	ns := NewNetShield(nil)
	df := NewDNSFilter(ns)

	ctx := context.Background()

	// Try to resolve blocked domain
	_, err := df.Resolve(ctx, "malware.com")
	if err == nil {
		t.Error("Should return error for blocked domain")
	}

	blockedErr, ok := err.(*BlockedDomainError)
	if !ok {
		t.Error("Should return BlockedDomainError")
	}

	if blockedErr.Domain != "malware.com" {
		t.Errorf("Domain = %s, want malware.com", blockedErr.Domain)
	}
}

// =============================================================================
// ThreatCategory Tests
// =============================================================================

func TestThreatCategoryString(t *testing.T) {
	testCases := []struct {
		category ThreatCategory
		expected string
	}{
		{ThreatNone, "none"},
		{ThreatMalware, "malware"},
		{ThreatPhishing, "phishing"},
		{ThreatAdware, "adware"},
		{ThreatTracker, "tracker"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.category.String() != tc.expected {
				t.Errorf("String() = %s, want %s", tc.category.String(), tc.expected)
			}
		})
	}
}

// =============================================================================
// HashDomain Tests
// =============================================================================

func TestHashDomain(t *testing.T) {
	hash1 := HashDomain("example.com")
	hash2 := HashDomain("example.com")
	hash3 := HashDomain("other.com")

	// Same domain should produce same hash
	if hash1 != hash2 {
		t.Error("Same domain should produce same hash")
	}

	// Different domains should produce different hashes
	if hash1 == hash3 {
		t.Error("Different domains should produce different hashes")
	}

	// Hash should be hex-encoded SHA256 (64 characters)
	if len(hash1) != 64 {
		t.Errorf("Hash length = %d, want 64", len(hash1))
	}
}

// =============================================================================
// BlockResult Tests
// =============================================================================

func TestBlockResultFields(t *testing.T) {
	ns := NewNetShield(nil)

	result := ns.CheckDomain("malware.com")

	if result.Domain != "malware.com" {
		t.Errorf("Domain = %s, want malware.com", result.Domain)
	}

	if result.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}

	if result.RuleMatched == "" {
		t.Error("RuleMatched should be set for blocked domains")
	}
}

// =============================================================================
// Config Tests
// =============================================================================

func TestDefaultNetShieldConfig(t *testing.T) {
	config := DefaultNetShieldConfig()

	if config.Level != NetShieldLevel2 {
		t.Errorf("Level = %d, want %d", config.Level, NetShieldLevel2)
	}

	if !config.BlockMalware {
		t.Error("BlockMalware should be true by default")
	}

	if !config.SafeDNS {
		t.Error("SafeDNS should be true by default")
	}
}

func TestMaxSecurityConfig(t *testing.T) {
	config := MaxSecurityConfig()

	if config.Level != NetShieldLevel3 {
		t.Errorf("Level = %d, want %d", config.Level, NetShieldLevel3)
	}

	if !config.BlockAdultContent {
		t.Error("BlockAdultContent should be true for max security")
	}

	if !config.BlockNewDomains {
		t.Error("BlockNewDomains should be true for max security")
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkNetShieldCheckDomain(b *testing.B) {
	ns := NewNetShield(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.CheckDomain("example.com")
	}
}

func BenchmarkNetShieldCheckDomainBlocked(b *testing.B) {
	ns := NewNetShield(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ns.CheckDomain("malware.com")
	}
}

func BenchmarkDNSFilterWithCache(b *testing.B) {
	ns := NewNetShield(nil)
	df := NewDNSFilter(ns)

	// Prime the cache
	df.Filter("test.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		df.Filter("test.com")
	}
}

func BenchmarkDomainSetContains(b *testing.B) {
	ds := NewDomainSet()
	for i := 0; i < 10000; i++ {
		ds.Add("domain" + string(rune(i)) + ".com")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ds.Contains("domain5000.com")
	}
}
