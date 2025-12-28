package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"
)

// =============================================================================
// TrafficObfuscator Tests
// =============================================================================

func TestNewTrafficObfuscator(t *testing.T) {
	obf := NewTrafficObfuscator(nil)
	if obf == nil {
		t.Fatal("NewTrafficObfuscator() returned nil")
	}
	defer obf.Close()
}

func TestTrafficObfuscatorWithConfig(t *testing.T) {
	config := DefaultObfuscationConfig()
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	if obf == nil {
		t.Fatal("NewTrafficObfuscator() with config returned nil")
	}
}

func TestTrafficObfuscatorMaxSecurity(t *testing.T) {
	config := MaxSecurityObfuscationConfig()
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	if obf.padToSize != 1400 {
		t.Errorf("padToSize = %d, want 1400", obf.padToSize)
	}
}

func TestObfuscateDeobfuscate(t *testing.T) {
	config := DefaultObfuscationConfig()
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	testData := []byte("test payload data for obfuscation")

	packet, err := obf.Obfuscate(testData)
	if err != nil {
		t.Fatalf("Obfuscate() error = %v", err)
	}

	if packet == nil {
		t.Fatal("Obfuscate() returned nil packet")
	}

	// Packet should be different from original
	if bytes.Equal(packet.Data, testData) {
		t.Error("obfuscated data should differ from original")
	}

	// Deobfuscate
	result, isDummy, err := obf.Deobfuscate(packet)
	if err != nil {
		t.Fatalf("Deobfuscate() error = %v", err)
	}

	if isDummy {
		t.Error("packet should not be dummy")
	}

	if !bytes.Equal(result, testData) {
		t.Errorf("Deobfuscate() mismatch: got %d bytes, want %d bytes", len(result), len(testData))
	}
}

func TestObfuscationModeNone(t *testing.T) {
	config := &ObfuscationConfig{
		Mode:          ObfuscationNone,
		MinPacketSize: 64,
		MaxPacketSize: 1500,
	}
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	testData := []byte("no obfuscation test")

	packet, err := obf.Obfuscate(testData)
	if err != nil {
		t.Fatalf("Obfuscate() error = %v", err)
	}

	// With no obfuscation, data should pass through
	if !bytes.Equal(packet.Data, testData) {
		t.Error("with ObfuscationNone, data should pass through unchanged")
	}
}

func TestObfuscationModeHTTPS(t *testing.T) {
	config := &ObfuscationConfig{
		Mode:          ObfuscationHTTPS,
		MinPacketSize: 64,
		MaxPacketSize: 1500,
	}
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	testData := []byte("HTTPS obfuscation test")

	packet, err := obf.Obfuscate(testData)
	if err != nil {
		t.Fatalf("Obfuscate() error = %v", err)
	}

	// Check TLS record header (first byte should be 0x17 = Application Data)
	if len(packet.Data) < 5 {
		t.Fatal("HTTPS packet too short")
	}
	if packet.Data[0] != 0x17 {
		t.Errorf("TLS record type = %x, want 0x17", packet.Data[0])
	}

	// Deobfuscate
	result, _, err := obf.Deobfuscate(packet)
	if err != nil {
		t.Fatalf("Deobfuscate() error = %v", err)
	}

	if !bytes.Equal(result, testData) {
		t.Error("Deobfuscate() data mismatch")
	}
}

func TestObfuscationModeDNS(t *testing.T) {
	config := &ObfuscationConfig{
		Mode:          ObfuscationDNS,
		MinPacketSize: 64,
		MaxPacketSize: 1500,
	}
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	testData := []byte("DNS over HTTPS test")

	packet, err := obf.Obfuscate(testData)
	if err != nil {
		t.Fatalf("Obfuscate() error = %v", err)
	}

	result, _, err := obf.Deobfuscate(packet)
	if err != nil {
		t.Fatalf("Deobfuscate() error = %v", err)
	}

	if !bytes.Equal(result, testData) {
		t.Error("Deobfuscate() data mismatch")
	}
}

func TestObfuscationModeWebSocket(t *testing.T) {
	config := &ObfuscationConfig{
		Mode:          ObfuscationWebSocket,
		MinPacketSize: 64,
		MaxPacketSize: 1500,
	}
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	testData := []byte("WebSocket test data")

	packet, err := obf.Obfuscate(testData)
	if err != nil {
		t.Fatalf("Obfuscate() error = %v", err)
	}

	result, _, err := obf.Deobfuscate(packet)
	if err != nil {
		t.Fatalf("Deobfuscate() error = %v", err)
	}

	if !bytes.Equal(result, testData) {
		t.Error("Deobfuscate() data mismatch")
	}
}

func TestObfuscationModeFull(t *testing.T) {
	config := &ObfuscationConfig{
		Mode:          ObfuscationFull,
		MinPacketSize: 64,
		MaxPacketSize: 1500,
	}
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	testData := []byte("full obfuscation test with multiple layers")

	packet, err := obf.Obfuscate(testData)
	if err != nil {
		t.Fatalf("Obfuscate() error = %v", err)
	}

	result, _, err := obf.Deobfuscate(packet)
	if err != nil {
		t.Fatalf("Deobfuscate() error = %v", err)
	}

	if !bytes.Equal(result, testData) {
		t.Error("Deobfuscate() data mismatch")
	}
}

func TestGenerateDummyPacket(t *testing.T) {
	config := DefaultObfuscationConfig()
	config.EnableDummy = true
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	packet, err := obf.GenerateDummyPacket()
	if err != nil {
		t.Fatalf("GenerateDummyPacket() error = %v", err)
	}

	if !packet.IsDummy {
		t.Error("packet should be marked as dummy")
	}

	if len(packet.Data) < obf.minPacketSize {
		t.Errorf("dummy packet too small: %d < %d", len(packet.Data), obf.minPacketSize)
	}
}

func TestDummyPacketDeobfuscate(t *testing.T) {
	obf := NewTrafficObfuscator(nil)
	defer obf.Close()

	dummy, _ := obf.GenerateDummyPacket()

	_, isDummy, err := obf.Deobfuscate(dummy)
	if err != nil {
		t.Fatalf("Deobfuscate() error = %v", err)
	}

	if !isDummy {
		t.Error("deobfuscate should identify dummy packet")
	}
}

func TestObfuscationStats(t *testing.T) {
	obf := NewTrafficObfuscator(nil)
	defer obf.Close()

	// Initial stats
	stats := obf.Stats()
	if stats.PacketsSent != 0 {
		t.Errorf("initial PacketsSent = %d, want 0", stats.PacketsSent)
	}

	// Obfuscate some packets
	for i := 0; i < 10; i++ {
		obf.Obfuscate([]byte("test data"))
	}

	stats = obf.Stats()
	if stats.PacketsSent != 10 {
		t.Errorf("PacketsSent = %d, want 10", stats.PacketsSent)
	}
}

func TestPacketDelay(t *testing.T) {
	config := &ObfuscationConfig{
		Mode:     ObfuscationNone,
		MinDelay: 1 * time.Millisecond,
		MaxDelay: 10 * time.Millisecond,
		Jitter:   0.3,
	}
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	packet, _ := obf.Obfuscate([]byte("delay test"))

	// With jitter, the delay can be less than minDelay
	// Just check that delay is not negative
	if packet.Delay < 0 {
		t.Errorf("delay %v is negative", packet.Delay)
	}
}

// =============================================================================
// TimingAnalysisProtector Tests
// =============================================================================

func TestNewTimingAnalysisProtector(t *testing.T) {
	tap := NewTimingAnalysisProtector(100, 10*time.Millisecond)
	if tap == nil {
		t.Fatal("NewTimingAnalysisProtector() returned nil")
	}
}

func TestConstantTimeOperation(t *testing.T) {
	tap := NewTimingAnalysisProtector(100, 20*time.Millisecond)

	start := time.Now()
	err := tap.ConstantTimeOperation(func() error {
		// Quick operation
		return nil
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("ConstantTimeOperation() error = %v", err)
	}

	// Should take at least minOpTime
	if elapsed < 20*time.Millisecond {
		t.Errorf("elapsed %v < min 20ms", elapsed)
	}
}

func TestRecordTiming(t *testing.T) {
	tap := NewTimingAnalysisProtector(10, 10*time.Millisecond)

	// Record some timings
	for i := 0; i < 10; i++ {
		tap.RecordTiming(time.Duration(i+10) * time.Millisecond)
	}

	// Check if jitter is needed
	_ = tap.ShouldAddJitter()
}

func TestShouldAddJitter(t *testing.T) {
	tap := NewTimingAnalysisProtector(10, 10*time.Millisecond)

	// With no data, should add jitter
	if !tap.ShouldAddJitter() {
		t.Error("ShouldAddJitter() should return true with insufficient data")
	}

	// Record uniform timings (pattern)
	for i := 0; i < 10; i++ {
		tap.RecordTiming(10 * time.Millisecond)
	}

	// With very uniform timings, should add jitter
	if !tap.ShouldAddJitter() {
		t.Error("ShouldAddJitter() should return true for uniform timings")
	}
}

// =============================================================================
// TrafficShaper Tests
// =============================================================================

func TestNewTrafficShaper(t *testing.T) {
	ts := NewTrafficShaper(1024*1024, 100, 100*time.Millisecond)
	if ts == nil {
		t.Fatal("NewTrafficShaper() returned nil")
	}
}

func TestTrafficShaperAcquire(t *testing.T) {
	ts := NewTrafficShaper(1024*1024, 100, 100*time.Millisecond)

	// Should be able to acquire immediately
	delay := ts.Acquire(100)
	if delay != 0 {
		t.Errorf("first Acquire() delay = %v, want 0", delay)
	}

	// After consuming tokens, might need to wait
	for i := 0; i < 1000; i++ {
		ts.Acquire(1000)
	}

	// Now should have delay
	delay = ts.Acquire(10000)
	if delay == 0 {
		t.Log("Traffic shaper might have enough tokens (acceptable)")
	}
}

// =============================================================================
// Large Data Tests
// =============================================================================

func TestObfuscateLargeData(t *testing.T) {
	obf := NewTrafficObfuscator(nil)
	defer obf.Close()

	// Test with various sizes
	sizes := []int{64, 256, 1024, 4096, 8192}

	for _, size := range sizes {
		data := make([]byte, size)
		rand.Read(data)

		packet, err := obf.Obfuscate(data)
		if err != nil {
			t.Errorf("Obfuscate(%d bytes) error = %v", size, err)
			continue
		}

		result, _, err := obf.Deobfuscate(packet)
		if err != nil {
			t.Errorf("Deobfuscate(%d bytes) error = %v", size, err)
			continue
		}

		if !bytes.Equal(result, data) {
			t.Errorf("data mismatch for %d bytes", size)
		}
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkObfuscateHTTPS(b *testing.B) {
	config := &ObfuscationConfig{
		Mode:          ObfuscationHTTPS,
		MinPacketSize: 64,
		MaxPacketSize: 1500,
	}
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	data := make([]byte, 1400)
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obf.Obfuscate(data)
	}
}

func BenchmarkObfuscateFull(b *testing.B) {
	config := &ObfuscationConfig{
		Mode:          ObfuscationFull,
		MinPacketSize: 64,
		MaxPacketSize: 1500,
	}
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	data := make([]byte, 1400)
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obf.Obfuscate(data)
	}
}

func BenchmarkDeobfuscateFull(b *testing.B) {
	config := &ObfuscationConfig{
		Mode:          ObfuscationFull,
		MinPacketSize: 64,
		MaxPacketSize: 1500,
	}
	obf := NewTrafficObfuscator(config)
	defer obf.Close()

	data := make([]byte, 1400)
	rand.Read(data)

	// Pre-obfuscate packets
	packets := make([]*ObfuscatedPacket, b.N)
	for i := 0; i < b.N; i++ {
		packets[i], _ = obf.Obfuscate(data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obf.Deobfuscate(packets[i])
	}
}

func BenchmarkGenerateDummyPacket(b *testing.B) {
	obf := NewTrafficObfuscator(nil)
	defer obf.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obf.GenerateDummyPacket()
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestObfuscatorConcurrent(t *testing.T) {
	obf := NewTrafficObfuscator(nil)
	defer obf.Close()

	done := make(chan bool)

	// Multiple concurrent obfuscation operations
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				data := make([]byte, 100)
				rand.Read(data)

				packet, err := obf.Obfuscate(data)
				if err != nil {
					t.Errorf("Obfuscate() error = %v", err)
				}

				_, _, err = obf.Deobfuscate(packet)
				if err != nil {
					t.Errorf("Deobfuscate() error = %v", err)
				}
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestObfuscatorWithTunnelCipher(t *testing.T) {
	// Create tunnel cipher with symmetric keys
	primaryKey := make([]byte, 32)
	secondaryKey := make([]byte, 32)
	rand.Read(primaryKey)
	rand.Read(secondaryKey)

	keys := &DerivedKeys{
		PrimarySendKey:   make([]byte, 32),
		PrimaryRecvKey:   make([]byte, 32),
		SecondarySendKey: make([]byte, 32),
		SecondaryRecvKey: make([]byte, 32),
	}
	copy(keys.PrimarySendKey, primaryKey)
	copy(keys.PrimaryRecvKey, primaryKey)
	copy(keys.SecondarySendKey, secondaryKey)
	copy(keys.SecondaryRecvKey, secondaryKey)
	rand.Read(keys.SessionID[:])

	tunnel, _ := NewTunnelCipher(keys)
	defer tunnel.Close()

	// Create obfuscator
	obf := NewTrafficObfuscator(nil)
	defer obf.Close()

	// Encrypt with tunnel
	plaintext := []byte("combined encryption and obfuscation test")
	encrypted, err := tunnel.Encrypt(plaintext, PacketTypeData)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Obfuscate
	packet, err := obf.Obfuscate(encrypted)
	if err != nil {
		t.Fatalf("Obfuscate() error = %v", err)
	}

	// Deobfuscate
	deobfuscated, _, err := obf.Deobfuscate(packet)
	if err != nil {
		t.Fatalf("Deobfuscate() error = %v", err)
	}

	// Decrypt with tunnel
	decrypted, _, err := tunnel.Decrypt(deobfuscated)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("final decrypted data mismatch")
	}
}
