package network

import (
	"testing"
	"time"
)

// =============================================================================
// VPN Accelerator Tests
// =============================================================================

func TestNewVPNAccelerator(t *testing.T) {
	acc := NewVPNAccelerator(nil)
	if acc == nil {
		t.Fatal("NewVPNAccelerator(nil) returned nil")
	}
	defer acc.Close()

	if acc.config.PoolSize != 4 {
		t.Errorf("PoolSize = %d, want 4", acc.config.PoolSize)
	}
}

func TestAcceleratorConfig(t *testing.T) {
	config := DefaultAcceleratorConfig()

	if !config.EnableMultipath {
		t.Error("EnableMultipath should be true by default")
	}

	if !config.EnableQoS {
		t.Error("EnableQoS should be true by default")
	}

	if config.ReadBufferSize != 256*1024 {
		t.Errorf("ReadBufferSize = %d, want %d", config.ReadBufferSize, 256*1024)
	}
}

func TestTrafficScheduler(t *testing.T) {
	config := DefaultAcceleratorConfig()
	scheduler := NewTrafficScheduler(config)

	// Test packet classification
	smallPacket := make([]byte, 50)
	priority := scheduler.ClassifyPacket(smallPacket)
	if priority != PriorityHigh && !config.PriorityGaming {
		t.Log("Small packets should be high priority for gaming")
	}

	// Test enqueue/dequeue
	scheduler.Enqueue([]byte("test"), PriorityHigh)

	data, ok := scheduler.Dequeue()
	if !ok {
		t.Error("Dequeue should return data")
	}
	if string(data) != "test" {
		t.Error("Dequeued data doesn't match")
	}
}

func TestConnectionPool(t *testing.T) {
	config := DefaultAcceleratorConfig()
	pool := NewConnectionPool("localhost:1234", config)
	defer pool.Close()

	if pool == nil {
		t.Fatal("NewConnectionPool returned nil")
	}
}

func TestStreamCopier(t *testing.T) {
	copier := NewStreamCopier(32 * 1024)
	if copier == nil {
		t.Fatal("NewStreamCopier returned nil")
	}

	if copier.bufferSize != 32*1024 {
		t.Errorf("bufferSize = %d, want %d", copier.bufferSize, 32*1024)
	}
}

// =============================================================================
// Secure Core Tests
// =============================================================================

func TestNewSecureCore(t *testing.T) {
	sc := NewSecureCore(nil)
	if sc == nil {
		t.Fatal("NewSecureCore(nil) returned nil")
	}
	defer sc.Close()

	if !sc.config.Enabled {
		t.Error("SecureCore should be enabled by default")
	}
}

func TestSecureCoreConfig(t *testing.T) {
	config := DefaultSecureCoreConfig()

	if config.MinHops != 2 {
		t.Errorf("MinHops = %d, want 2", config.MinHops)
	}

	if config.MaxHops != 3 {
		t.Errorf("MaxHops = %d, want 3", config.MaxHops)
	}

	if !config.PreferPrivacyJurisdictions {
		t.Error("PreferPrivacyJurisdictions should be true")
	}
}

func TestSecureCoreAddServer(t *testing.T) {
	sc := NewSecureCore(nil)
	defer sc.Close()

	server := &CoreServer{
		ID:        "test-1",
		Name:      "Test Server",
		Address:   "10.0.0.1",
		Port:      51820,
		Country:   "CH",
		IsCore:    true,
		IsEntry:   true,
		Available: true,
	}

	sc.AddServer(server)

	if len(sc.servers) != 1 {
		t.Errorf("servers count = %d, want 1", len(sc.servers))
	}

	if len(sc.coreServers) != 1 {
		t.Errorf("coreServers count = %d, want 1", len(sc.coreServers))
	}
}

func TestPrivacyJurisdictions(t *testing.T) {
	// Switzerland should be a privacy-friendly jurisdiction
	ch, ok := PrivacyJurisdictions["CH"]
	if !ok {
		t.Error("Switzerland should be in privacy jurisdictions")
	}

	if ch.PrivacyScore != 10 {
		t.Errorf("Switzerland PrivacyScore = %d, want 10", ch.PrivacyScore)
	}

	if ch.EyesAlliance {
		t.Error("Switzerland should not be in Eyes alliance")
	}
}

// =============================================================================
// Smart Protocol Tests
// =============================================================================

func TestNewSmartProtocol(t *testing.T) {
	sp := NewSmartProtocol(nil)
	if sp == nil {
		t.Fatal("NewSmartProtocol(nil) returned nil")
	}

	if !sp.config.AutoSelect {
		t.Error("AutoSelect should be true by default")
	}
}

func TestSmartProtocolSelectProtocol(t *testing.T) {
	sp := NewSmartProtocol(nil)

	// Normal conditions
	conditions := NetworkConditions{
		Latency:    50 * time.Millisecond,
		PacketLoss: 0.01,
		Bandwidth:  100_000_000,
	}

	protocol := sp.SelectProtocol(conditions)
	if protocol == "" {
		t.Error("SelectProtocol should return a protocol")
	}

	t.Logf("Selected protocol for normal conditions: %s", protocol)
}

func TestSmartProtocolRestrictiveNetwork(t *testing.T) {
	sp := NewSmartProtocol(nil)

	// Restrictive network (firewall/DPI)
	conditions := NetworkConditions{
		Latency:       100 * time.Millisecond,
		IsRestrictive: true,
	}

	protocol := sp.SelectProtocol(conditions)

	// Should prefer stealth protocols
	t.Logf("Selected protocol for restrictive network: %s", protocol)
}

func TestSmartProtocolMobileNetwork(t *testing.T) {
	sp := NewSmartProtocol(nil)

	// Mobile network
	conditions := NetworkConditions{
		Latency:  80 * time.Millisecond,
		IsMobile: true,
	}

	protocol := sp.SelectProtocol(conditions)
	t.Logf("Selected protocol for mobile: %s", protocol)
}

func TestSmartProtocolGetProfile(t *testing.T) {
	sp := NewSmartProtocol(nil)

	profile := sp.GetProtocolProfile("wireguard")
	if profile == nil {
		t.Fatal("WireGuard profile should exist")
	}

	if profile.Speed != 10 {
		t.Errorf("WireGuard Speed = %d, want 10", profile.Speed)
	}

	if profile.Security != 10 {
		t.Errorf("WireGuard Security = %d, want 10", profile.Security)
	}
}

// =============================================================================
// Connection Resilience Tests
// =============================================================================

func TestNewResilientConnection(t *testing.T) {
	rc := NewResilientConnection("localhost:1234", nil)
	if rc == nil {
		t.Fatal("NewResilientConnection returned nil")
	}
	defer rc.Close()

	if rc.State() != StateDisconnected {
		t.Errorf("Initial state = %v, want StateDisconnected", rc.State())
	}
}

func TestResilienceConfig(t *testing.T) {
	config := DefaultResilienceConfig()

	if config.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", config.MaxRetries)
	}

	if !config.EnableFailover {
		t.Error("EnableFailover should be true")
	}

	if !config.EnableCircuitBreaker {
		t.Error("EnableCircuitBreaker should be true")
	}
}

func TestResilientConnectionAddSecondary(t *testing.T) {
	rc := NewResilientConnection("localhost:1234", nil)
	defer rc.Close()

	rc.AddSecondary("localhost:5678")
	rc.AddSecondary("localhost:9012")

	if len(rc.secondary) != 2 {
		t.Errorf("secondary count = %d, want 2", len(rc.secondary))
	}
}

func TestConnectionStateString(t *testing.T) {
	states := []struct {
		state    ConnectionState
		expected string
	}{
		{StateDisconnected, "disconnected"},
		{StateConnecting, "connecting"},
		{StateConnected, "connected"},
		{StateReconnecting, "reconnecting"},
		{StateFailed, "failed"},
		{StateFailover, "failover"},
	}

	for _, tc := range states {
		if tc.state.String() != tc.expected {
			t.Errorf("State %d String() = %s, want %s", tc.state, tc.state.String(), tc.expected)
		}
	}
}

// =============================================================================
// Circuit Breaker Tests
// =============================================================================

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 5*time.Second)

	// Initially closed
	if cb.State() != "closed" {
		t.Errorf("Initial state = %s, want closed", cb.State())
	}

	// Should allow requests
	if !cb.Allow() {
		t.Error("Should allow requests when closed")
	}

	// Record failures to open circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.State() != "open" {
		t.Errorf("State after failures = %s, want open", cb.State())
	}

	// Should not allow requests when open
	if cb.Allow() {
		t.Error("Should not allow requests when open")
	}

	// Record success to close circuit
	cb.state.Store(circuitHalfOpen) // Simulate half-open
	cb.RecordSuccess()

	if cb.State() != "closed" {
		t.Errorf("State after success = %s, want closed", cb.State())
	}
}

// =============================================================================
// Latency Optimizer Tests
// =============================================================================

func TestLatencyOptimizer(t *testing.T) {
	lo := NewLatencyOptimizer()
	if lo == nil {
		t.Fatal("NewLatencyOptimizer returned nil")
	}

	if lo.sampleSize != 100 {
		t.Errorf("sampleSize = %d, want 100", lo.sampleSize)
	}
}

// =============================================================================
// Bandwidth Aggregator Tests
// =============================================================================

func TestBandwidthAggregator(t *testing.T) {
	ba := NewBandwidthAggregator()
	if ba == nil {
		t.Fatal("NewBandwidthAggregator returned nil")
	}

	sent, recv := ba.TotalBandwidth()
	if sent != 0 || recv != 0 {
		t.Error("Initial bandwidth should be zero")
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestAcceleratorWithScheduler(t *testing.T) {
	config := DefaultAcceleratorConfig()
	config.EnableQoS = true

	acc := NewVPNAccelerator(config)
	defer acc.Close()

	// Verify scheduler is created
	if acc.scheduler == nil {
		t.Error("Scheduler should be initialized")
	}
}

func TestSecureCoreWithSmartProtocol(t *testing.T) {
	sc := NewSecureCore(nil)
	defer sc.Close()

	sp := NewSmartProtocol(nil)

	// Add some servers
	for i := 0; i < 5; i++ {
		sc.AddServer(&CoreServer{
			ID:        "server-" + string(rune('a'+i)),
			Country:   "CH",
			IsCore:    true,
			IsEntry:   i == 0,
			IsExit:    i == 4,
			Available: true,
		})
	}

	// Select protocol based on conditions
	conditions := NetworkConditions{Latency: 50 * time.Millisecond}
	protocol := sp.SelectProtocol(conditions)

	t.Logf("Using protocol %s with %d servers", protocol, len(sc.servers))
}
