// Package security provides NetShield - comprehensive threat protection
// This blocks malware, ads, trackers, and malicious domains at the network level
package security

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// NetShieldLevel defines protection intensity
type NetShieldLevel int

const (
	NetShieldOff NetShieldLevel = iota
	NetShieldLevel1 // Block malware only
	NetShieldLevel2 // Block malware + ads + trackers
	NetShieldLevel3 // Block malware + ads + trackers + adult content
)

// ThreatCategory defines types of threats
type ThreatCategory int

const (
	ThreatNone ThreatCategory = iota
	ThreatMalware
	ThreatPhishing
	ThreatAdware
	ThreatTracker
	ThreatSpyware
	ThreatRansomware
	ThreatCryptojacking
	ThreatBotnet
	ThreatAdultContent
	ThreatGambling
	ThreatSocialMedia
)

// String returns threat category name
func (t ThreatCategory) String() string {
	names := map[ThreatCategory]string{
		ThreatNone:          "none",
		ThreatMalware:       "malware",
		ThreatPhishing:      "phishing",
		ThreatAdware:        "adware",
		ThreatTracker:       "tracker",
		ThreatSpyware:       "spyware",
		ThreatRansomware:    "ransomware",
		ThreatCryptojacking: "cryptojacking",
		ThreatBotnet:        "botnet",
		ThreatAdultContent:  "adult",
		ThreatGambling:      "gambling",
		ThreatSocialMedia:   "social",
	}
	return names[t]
}

// NetShieldConfig configures NetShield protection
type NetShieldConfig struct {
	Level NetShieldLevel

	// Enable specific protections
	BlockMalware       bool
	BlockPhishing      bool
	BlockAds           bool
	BlockTrackers      bool
	BlockCryptojacking bool
	BlockAdultContent  bool

	// Custom blocklists
	CustomBlockDomains []string
	CustomAllowDomains []string

	// DNS settings
	SafeDNS           bool
	DNSOverHTTPS      bool
	DNSOverTLS        bool

	// Filtering options
	BlockNewDomains   bool   // Block domains registered < 30 days
	NewDomainAgeDays  int    // Threshold for "new" domains
	BlockParkedDomains bool  // Block parked/unused domains

	// Logging
	LogBlocks         bool
	LogCategories     []ThreatCategory
}

// DefaultNetShieldConfig returns default settings
func DefaultNetShieldConfig() *NetShieldConfig {
	return &NetShieldConfig{
		Level:              NetShieldLevel2,
		BlockMalware:       true,
		BlockPhishing:      true,
		BlockAds:           true,
		BlockTrackers:      true,
		BlockCryptojacking: true,
		BlockAdultContent:  false,
		SafeDNS:            true,
		DNSOverHTTPS:       true,
		NewDomainAgeDays:   30,
		LogBlocks:          true,
	}
}

// MaxSecurityConfig returns maximum protection settings
func MaxSecurityConfig() *NetShieldConfig {
	return &NetShieldConfig{
		Level:              NetShieldLevel3,
		BlockMalware:       true,
		BlockPhishing:      true,
		BlockAds:           true,
		BlockTrackers:      true,
		BlockCryptojacking: true,
		BlockAdultContent:  true,
		BlockNewDomains:    true,
		BlockParkedDomains: true,
		SafeDNS:            true,
		DNSOverHTTPS:       true,
		DNSOverTLS:         true,
		NewDomainAgeDays:   30,
		LogBlocks:          true,
	}
}

// NetShield provides comprehensive threat protection
type NetShield struct {
	config *NetShieldConfig

	// Domain blocklists (by category)
	malwareDomains    *DomainSet
	phishingDomains   *DomainSet
	adDomains         *DomainSet
	trackerDomains    *DomainSet
	cryptoDomains     *DomainSet
	adultDomains      *DomainSet

	// IP blocklists
	malwareIPs        *IPSet
	botnetIPs         *IPSet

	// Custom lists
	customBlock       *DomainSet
	customAllow       *DomainSet

	// URL pattern matching
	adPatterns        []*regexp.Regexp
	trackerPatterns   []*regexp.Regexp

	// Statistics
	stats NetShieldStats

	// Update tracking
	lastUpdate time.Time
	updateMu   sync.RWMutex

	mu sync.RWMutex
}

// NetShieldStats tracks protection statistics
type NetShieldStats struct {
	TotalRequests      atomic.Uint64
	BlockedRequests    atomic.Uint64
	MalwareBlocked     atomic.Uint64
	PhishingBlocked    atomic.Uint64
	AdsBlocked         atomic.Uint64
	TrackersBlocked    atomic.Uint64
	CryptoBlocked      atomic.Uint64
	AdultBlocked       atomic.Uint64
	CustomBlocked      atomic.Uint64
}

// BlockResult represents the result of a threat check
type BlockResult struct {
	Blocked     bool
	Category    ThreatCategory
	Reason      string
	Domain      string
	Timestamp   time.Time
	RuleMatched string
}

// NewNetShield creates a new NetShield instance
func NewNetShield(config *NetShieldConfig) *NetShield {
	if config == nil {
		config = DefaultNetShieldConfig()
	}

	ns := &NetShield{
		config:          config,
		malwareDomains:  NewDomainSet(),
		phishingDomains: NewDomainSet(),
		adDomains:       NewDomainSet(),
		trackerDomains:  NewDomainSet(),
		cryptoDomains:   NewDomainSet(),
		adultDomains:    NewDomainSet(),
		malwareIPs:      NewIPSet(),
		botnetIPs:       NewIPSet(),
		customBlock:     NewDomainSet(),
		customAllow:     NewDomainSet(),
		lastUpdate:      time.Now(),
	}

	// Initialize with built-in patterns
	ns.initBuiltinPatterns()

	// Load custom lists
	ns.loadCustomLists()

	return ns
}

// initBuiltinPatterns initializes built-in blocking patterns
func (ns *NetShield) initBuiltinPatterns() {
	// Common ad patterns
	adPatternStrs := []string{
		`^ads?\.`,
		`\.ads?\.`,
		`^adserver\.`,
		`\.adserver\.`,
		`^adservice\.`,
		`^doubleclick\.`,
		`^googlesyndication\.`,
		`^googleadservices\.`,
		`\.advertising\.`,
		`^banner\.`,
		`^banners\.`,
		`^advert\.`,
		`^tracking\.`,
		`^pixel\.`,
		`^beacon\.`,
		`^analytics\.`,
		`^telemetry\.`,
		`^metrics\.`,
		`^stat\.`,
		`^stats\.`,
	}

	ns.adPatterns = make([]*regexp.Regexp, 0, len(adPatternStrs))
	for _, pattern := range adPatternStrs {
		if re, err := regexp.Compile(pattern); err == nil {
			ns.adPatterns = append(ns.adPatterns, re)
		}
	}

	// Common tracker patterns
	trackerPatternStrs := []string{
		`^track\.`,
		`^tracker\.`,
		`^tracking\.`,
		`^pixel\.`,
		`^collect\.`,
		`^analytics\.`,
		`^telemetry\.`,
		`^fingerprint\.`,
		`^beacon\.`,
		`^log\.`,
		`^logs\.`,
		`\.tracker\.`,
		`\.tracking\.`,
		`\.analytics\.`,
	}

	ns.trackerPatterns = make([]*regexp.Regexp, 0, len(trackerPatternStrs))
	for _, pattern := range trackerPatternStrs {
		if re, err := regexp.Compile(pattern); err == nil {
			ns.trackerPatterns = append(ns.trackerPatterns, re)
		}
	}

	// Load known malware/phishing domains
	ns.loadBuiltinDomains()
}

// loadBuiltinDomains loads common known-bad domains
func (ns *NetShield) loadBuiltinDomains() {
	// Known malware domains (sample)
	malwareDomains := []string{
		"malware.com",
		"virus.com",
		"trojan.net",
		"ransomware.org",
		"exploit.io",
		"badware.xyz",
		"maliciousdomain.net",
	}
	for _, d := range malwareDomains {
		ns.malwareDomains.Add(d)
	}

	// Known phishing domains (sample)
	phishingDomains := []string{
		"phishing.com",
		"fakebanklogin.com",
		"secure-login-verify.com",
		"account-update-required.com",
	}
	for _, d := range phishingDomains {
		ns.phishingDomains.Add(d)
	}

	// Known ad domains (sample)
	adDomains := []string{
		"doubleclick.net",
		"googlesyndication.com",
		"googleadservices.com",
		"googleads.g.doubleclick.net",
		"adservice.google.com",
		"pagead2.googlesyndication.com",
		"adsserver.com",
		"adnxs.com",
		"facebook.com/tr",
		"connect.facebook.net",
	}
	for _, d := range adDomains {
		ns.adDomains.Add(d)
	}

	// Known tracker domains (sample)
	trackerDomains := []string{
		"google-analytics.com",
		"googletagmanager.com",
		"facebook.com/tr",
		"hotjar.com",
		"segment.io",
		"mixpanel.com",
		"amplitude.com",
		"fullstory.com",
		"heap.io",
		"mouseflow.com",
	}
	for _, d := range trackerDomains {
		ns.trackerDomains.Add(d)
	}

	// Known cryptojacking domains
	cryptoDomains := []string{
		"coinhive.com",
		"crypto-loot.org",
		"coin-hive.com",
		"jsecoin.com",
		"cryptonight.wasm",
	}
	for _, d := range cryptoDomains {
		ns.cryptoDomains.Add(d)
	}
}

// loadCustomLists loads user-defined block/allow lists
func (ns *NetShield) loadCustomLists() {
	for _, domain := range ns.config.CustomBlockDomains {
		ns.customBlock.Add(normalizeDomain(domain))
	}

	for _, domain := range ns.config.CustomAllowDomains {
		ns.customAllow.Add(normalizeDomain(domain))
	}
}

// CheckDomain checks if a domain should be blocked
func (ns *NetShield) CheckDomain(domain string) BlockResult {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	ns.stats.TotalRequests.Add(1)

	domain = normalizeDomain(domain)

	result := BlockResult{
		Domain:    domain,
		Timestamp: time.Now(),
	}

	// Check allow list first
	if ns.customAllow.Contains(domain) || ns.customAllow.ContainsSuffix(domain) {
		return result // Not blocked
	}

	// Check custom block list
	if ns.customBlock.Contains(domain) || ns.customBlock.ContainsSuffix(domain) {
		result.Blocked = true
		result.Category = ThreatNone
		result.Reason = "custom blocklist"
		result.RuleMatched = "custom"
		ns.stats.CustomBlocked.Add(1)
		ns.stats.BlockedRequests.Add(1)
		return result
	}

	// Check malware (always if enabled)
	if ns.config.BlockMalware {
		if ns.malwareDomains.Contains(domain) || ns.malwareDomains.ContainsSuffix(domain) {
			result.Blocked = true
			result.Category = ThreatMalware
			result.Reason = "known malware domain"
			result.RuleMatched = "malware-list"
			ns.stats.MalwareBlocked.Add(1)
			ns.stats.BlockedRequests.Add(1)
			return result
		}
	}

	// Check phishing
	if ns.config.BlockPhishing {
		if ns.phishingDomains.Contains(domain) || ns.phishingDomains.ContainsSuffix(domain) {
			result.Blocked = true
			result.Category = ThreatPhishing
			result.Reason = "known phishing domain"
			result.RuleMatched = "phishing-list"
			ns.stats.PhishingBlocked.Add(1)
			ns.stats.BlockedRequests.Add(1)
			return result
		}
	}

	// Check cryptojacking
	if ns.config.BlockCryptojacking {
		if ns.cryptoDomains.Contains(domain) || ns.cryptoDomains.ContainsSuffix(domain) {
			result.Blocked = true
			result.Category = ThreatCryptojacking
			result.Reason = "cryptojacking domain"
			result.RuleMatched = "crypto-list"
			ns.stats.CryptoBlocked.Add(1)
			ns.stats.BlockedRequests.Add(1)
			return result
		}
	}

	// Check ads (Level 2+)
	if ns.config.Level >= NetShieldLevel2 && ns.config.BlockAds {
		if ns.adDomains.Contains(domain) || ns.adDomains.ContainsSuffix(domain) {
			result.Blocked = true
			result.Category = ThreatAdware
			result.Reason = "advertising domain"
			result.RuleMatched = "ad-list"
			ns.stats.AdsBlocked.Add(1)
			ns.stats.BlockedRequests.Add(1)
			return result
		}

		// Check ad patterns
		for _, pattern := range ns.adPatterns {
			if pattern.MatchString(domain) {
				result.Blocked = true
				result.Category = ThreatAdware
				result.Reason = "advertising pattern match"
				result.RuleMatched = pattern.String()
				ns.stats.AdsBlocked.Add(1)
				ns.stats.BlockedRequests.Add(1)
				return result
			}
		}
	}

	// Check trackers (Level 2+)
	if ns.config.Level >= NetShieldLevel2 && ns.config.BlockTrackers {
		if ns.trackerDomains.Contains(domain) || ns.trackerDomains.ContainsSuffix(domain) {
			result.Blocked = true
			result.Category = ThreatTracker
			result.Reason = "tracking domain"
			result.RuleMatched = "tracker-list"
			ns.stats.TrackersBlocked.Add(1)
			ns.stats.BlockedRequests.Add(1)
			return result
		}

		// Check tracker patterns
		for _, pattern := range ns.trackerPatterns {
			if pattern.MatchString(domain) {
				result.Blocked = true
				result.Category = ThreatTracker
				result.Reason = "tracking pattern match"
				result.RuleMatched = pattern.String()
				ns.stats.TrackersBlocked.Add(1)
				ns.stats.BlockedRequests.Add(1)
				return result
			}
		}
	}

	// Check adult content (Level 3)
	if ns.config.Level >= NetShieldLevel3 && ns.config.BlockAdultContent {
		if ns.adultDomains.Contains(domain) || ns.adultDomains.ContainsSuffix(domain) {
			result.Blocked = true
			result.Category = ThreatAdultContent
			result.Reason = "adult content"
			result.RuleMatched = "adult-list"
			ns.stats.AdultBlocked.Add(1)
			ns.stats.BlockedRequests.Add(1)
			return result
		}
	}

	return result // Not blocked
}

// CheckIP checks if an IP address should be blocked
func (ns *NetShield) CheckIP(ip net.IP) BlockResult {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	ns.stats.TotalRequests.Add(1)

	result := BlockResult{
		Domain:    ip.String(),
		Timestamp: time.Now(),
	}

	// Check malware IPs
	if ns.config.BlockMalware && ns.malwareIPs.Contains(ip) {
		result.Blocked = true
		result.Category = ThreatMalware
		result.Reason = "known malware IP"
		result.RuleMatched = "malware-ip-list"
		ns.stats.MalwareBlocked.Add(1)
		ns.stats.BlockedRequests.Add(1)
		return result
	}

	// Check botnet IPs
	if ns.botnetIPs.Contains(ip) {
		result.Blocked = true
		result.Category = ThreatBotnet
		result.Reason = "known botnet IP"
		result.RuleMatched = "botnet-ip-list"
		ns.stats.MalwareBlocked.Add(1)
		ns.stats.BlockedRequests.Add(1)
		return result
	}

	return result
}

// CheckURL checks a full URL for threats
func (ns *NetShield) CheckURL(rawURL string) BlockResult {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return BlockResult{Blocked: false}
	}

	// First check the domain
	result := ns.CheckDomain(parsed.Host)
	if result.Blocked {
		return result
	}

	// Check URL path for suspicious patterns
	path := strings.ToLower(parsed.Path + "?" + parsed.RawQuery)

	// Check for common phishing patterns in URL
	phishingPatterns := []string{
		"login", "signin", "password", "verify", "account",
		"update", "confirm", "secure", "bank", "paypal",
		".exe", ".scr", ".bat", ".cmd", ".vbs", ".js",
	}

	for _, pattern := range phishingPatterns {
		if strings.Contains(path, pattern) {
			// Additional heuristics needed - this is just a flag
			break
		}
	}

	return result
}

// AddBlockDomain adds a domain to the custom blocklist
func (ns *NetShield) AddBlockDomain(domain string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.customBlock.Add(normalizeDomain(domain))
}

// AddAllowDomain adds a domain to the custom allowlist
func (ns *NetShield) AddAllowDomain(domain string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.customAllow.Add(normalizeDomain(domain))
}

// RemoveBlockDomain removes a domain from the custom blocklist
func (ns *NetShield) RemoveBlockDomain(domain string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.customBlock.Remove(normalizeDomain(domain))
}

// RemoveAllowDomain removes a domain from the custom allowlist
func (ns *NetShield) RemoveAllowDomain(domain string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.customAllow.Remove(normalizeDomain(domain))
}

// LoadBlocklistFromReader loads domains from a reader (one per line)
func (ns *NetShield) LoadBlocklistFromReader(r *bufio.Reader, category ThreatCategory) (int, error) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	var domainSet *DomainSet
	switch category {
	case ThreatMalware:
		domainSet = ns.malwareDomains
	case ThreatPhishing:
		domainSet = ns.phishingDomains
	case ThreatAdware:
		domainSet = ns.adDomains
	case ThreatTracker:
		domainSet = ns.trackerDomains
	case ThreatCryptojacking:
		domainSet = ns.cryptoDomains
	case ThreatAdultContent:
		domainSet = ns.adultDomains
	default:
		domainSet = ns.customBlock
	}

	count := 0
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle hosts file format (IP domain)
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			line = parts[1]
		} else if len(parts) == 1 {
			line = parts[0]
		}

		domain := normalizeDomain(line)
		if domain != "" {
			domainSet.Add(domain)
			count++
		}
	}

	ns.lastUpdate = time.Now()

	return count, scanner.Err()
}

// Stats returns current statistics
func (ns *NetShield) Stats() NetShieldStatsSummary {
	return NetShieldStatsSummary{
		TotalRequests:   ns.stats.TotalRequests.Load(),
		BlockedRequests: ns.stats.BlockedRequests.Load(),
		MalwareBlocked:  ns.stats.MalwareBlocked.Load(),
		PhishingBlocked: ns.stats.PhishingBlocked.Load(),
		AdsBlocked:      ns.stats.AdsBlocked.Load(),
		TrackersBlocked: ns.stats.TrackersBlocked.Load(),
		CryptoBlocked:   ns.stats.CryptoBlocked.Load(),
		AdultBlocked:    ns.stats.AdultBlocked.Load(),
		CustomBlocked:   ns.stats.CustomBlocked.Load(),
		LastUpdate:      ns.lastUpdate,
	}
}

// NetShieldStatsSummary provides a summary of statistics
type NetShieldStatsSummary struct {
	TotalRequests   uint64
	BlockedRequests uint64
	MalwareBlocked  uint64
	PhishingBlocked uint64
	AdsBlocked      uint64
	TrackersBlocked uint64
	CryptoBlocked   uint64
	AdultBlocked    uint64
	CustomBlocked   uint64
	LastUpdate      time.Time
}

// BlockRate returns the percentage of blocked requests
func (s NetShieldStatsSummary) BlockRate() float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.BlockedRequests) / float64(s.TotalRequests) * 100
}

// SetLevel changes the protection level
func (ns *NetShield) SetLevel(level NetShieldLevel) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.config.Level = level

	// Update settings based on level
	switch level {
	case NetShieldOff:
		ns.config.BlockMalware = false
		ns.config.BlockPhishing = false
		ns.config.BlockAds = false
		ns.config.BlockTrackers = false
	case NetShieldLevel1:
		ns.config.BlockMalware = true
		ns.config.BlockPhishing = true
		ns.config.BlockAds = false
		ns.config.BlockTrackers = false
	case NetShieldLevel2:
		ns.config.BlockMalware = true
		ns.config.BlockPhishing = true
		ns.config.BlockAds = true
		ns.config.BlockTrackers = true
	case NetShieldLevel3:
		ns.config.BlockMalware = true
		ns.config.BlockPhishing = true
		ns.config.BlockAds = true
		ns.config.BlockTrackers = true
		ns.config.BlockAdultContent = true
	}
}

// normalizeDomain normalizes a domain name
func normalizeDomain(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "www.")

	// Remove port if present
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove path if present
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}

// DomainSet is a thread-safe set of domains
type DomainSet struct {
	domains map[string]struct{}
	mu      sync.RWMutex
}

// NewDomainSet creates a new domain set
func NewDomainSet() *DomainSet {
	return &DomainSet{
		domains: make(map[string]struct{}),
	}
}

// Add adds a domain to the set
func (ds *DomainSet) Add(domain string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.domains[domain] = struct{}{}
}

// Remove removes a domain from the set
func (ds *DomainSet) Remove(domain string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	delete(ds.domains, domain)
}

// Contains checks if a domain is in the set
func (ds *DomainSet) Contains(domain string) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	_, exists := ds.domains[domain]
	return exists
}

// ContainsSuffix checks if any domain in the set is a suffix of the given domain
func (ds *DomainSet) ContainsSuffix(domain string) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	parts := strings.Split(domain, ".")
	for i := range parts {
		suffix := strings.Join(parts[i:], ".")
		if _, exists := ds.domains[suffix]; exists {
			return true
		}
	}
	return false
}

// Size returns the number of domains in the set
func (ds *DomainSet) Size() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return len(ds.domains)
}

// IPSet is a thread-safe set of IP addresses
type IPSet struct {
	ips   map[string]struct{}
	cidrs []*net.IPNet
	mu    sync.RWMutex
}

// NewIPSet creates a new IP set
func NewIPSet() *IPSet {
	return &IPSet{
		ips:   make(map[string]struct{}),
		cidrs: make([]*net.IPNet, 0),
	}
}

// Add adds an IP address to the set
func (is *IPSet) Add(ip net.IP) {
	is.mu.Lock()
	defer is.mu.Unlock()
	is.ips[ip.String()] = struct{}{}
}

// AddCIDR adds a CIDR range to the set
func (is *IPSet) AddCIDR(cidr string) error {
	is.mu.Lock()
	defer is.mu.Unlock()

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}
	is.cidrs = append(is.cidrs, ipnet)
	return nil
}

// Contains checks if an IP is in the set
func (is *IPSet) Contains(ip net.IP) bool {
	is.mu.RLock()
	defer is.mu.RUnlock()

	// Check exact match
	if _, exists := is.ips[ip.String()]; exists {
		return true
	}

	// Check CIDR ranges
	for _, cidr := range is.cidrs {
		if cidr.Contains(ip) {
			return true
		}
	}

	return false
}

// DNSFilter provides DNS-level filtering with NetShield
type DNSFilter struct {
	netshield *NetShield

	// Cache for performance
	cache     map[string]BlockResult
	cacheSize int
	cacheMu   sync.RWMutex

	// Safe DNS resolver
	resolver *net.Resolver
}

// NewDNSFilter creates a DNS filter with NetShield
func NewDNSFilter(ns *NetShield) *DNSFilter {
	return &DNSFilter{
		netshield: ns,
		cache:     make(map[string]BlockResult),
		cacheSize: 10000,
		resolver:  net.DefaultResolver,
	}
}

// Filter filters a DNS query
func (df *DNSFilter) Filter(domain string) BlockResult {
	domain = normalizeDomain(domain)

	// Check cache
	df.cacheMu.RLock()
	if result, ok := df.cache[domain]; ok {
		df.cacheMu.RUnlock()
		return result
	}
	df.cacheMu.RUnlock()

	// Check with NetShield
	result := df.netshield.CheckDomain(domain)

	// Cache result
	df.cacheMu.Lock()
	if len(df.cache) >= df.cacheSize {
		// Simple cache eviction - clear half
		for k := range df.cache {
			delete(df.cache, k)
			if len(df.cache) < df.cacheSize/2 {
				break
			}
		}
	}
	df.cache[domain] = result
	df.cacheMu.Unlock()

	return result
}

// Resolve performs filtered DNS resolution
func (df *DNSFilter) Resolve(ctx context.Context, domain string) ([]net.IP, error) {
	// Check if domain should be blocked
	result := df.Filter(domain)
	if result.Blocked {
		return nil, &BlockedDomainError{
			Domain:   domain,
			Category: result.Category,
			Reason:   result.Reason,
		}
	}

	// Perform actual DNS resolution
	addrs, err := df.resolver.LookupIPAddr(ctx, domain)
	if err != nil {
		return nil, err
	}

	ips := make([]net.IP, len(addrs))
	for i, addr := range addrs {
		ips[i] = addr.IP
	}

	// Also check resolved IPs
	for _, ip := range ips {
		ipResult := df.netshield.CheckIP(ip)
		if ipResult.Blocked {
			return nil, &BlockedDomainError{
				Domain:   domain,
				Category: ipResult.Category,
				Reason:   "resolved IP is blocked: " + ip.String(),
			}
		}
	}

	return ips, nil
}

// BlockedDomainError is returned when a domain is blocked
type BlockedDomainError struct {
	Domain   string
	Category ThreatCategory
	Reason   string
}

func (e *BlockedDomainError) Error() string {
	return "blocked: " + e.Domain + " (" + e.Category.String() + ": " + e.Reason + ")"
}

// HashDomain creates a privacy-preserving hash of a domain
func HashDomain(domain string) string {
	hash := sha256.Sum256([]byte(normalizeDomain(domain)))
	return hex.EncodeToString(hash[:])
}
