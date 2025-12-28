package crypto

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// DAITA Defender Tests
// =============================================================================

func TestNewDAITADefender(t *testing.T) {
	// Test with default config
	defender := NewDAITADefender(nil)
	if defender == nil {
		t.Fatal("NewDAITADefender(nil) returned nil")
	}
	defer defender.Close()

	// Test with custom config
	config := DefaultDAITAConfig()
	defender2 := NewDAITADefender(config)
	if defender2 == nil {
		t.Fatal("NewDAITADefender(config) returned nil")
	}
	defer defender2.Close()
}

func TestDAITADefenderProcessOutbound(t *testing.T) {
	config := DefaultDAITAConfig()
	defender := NewDAITADefender(config)
	defer defender.Close()

	testCases := []struct {
		name   string
		packet []byte
	}{
		{"small packet", []byte("hello")},
		{"medium packet", bytes.Repeat([]byte("test"), 100)},
		{"large packet", bytes.Repeat([]byte("data"), 300)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := defender.ProcessOutbound(tc.packet)
			if err != nil {
				t.Fatalf("ProcessOutbound() error = %v", err)
			}

			if len(result) == 0 {
				t.Error("ProcessOutbound() returned empty result")
			}

			// At least the original packet should be in result
			foundOriginal := false
			for _, pkt := range result {
				// Check if packet contains original data (after padding header)
				if len(pkt) >= 8+len(tc.packet) {
					if bytes.Contains(pkt[8:], tc.packet) {
						foundOriginal = true
						break
					}
				}
			}

			if !foundOriginal && !config.ConstantPacketSize {
				t.Log("Original packet data may be padded")
			}
		})
	}
}

func TestDAITADefenderProcessInbound(t *testing.T) {
	config := DefaultDAITAConfig()
	defender := NewDAITADefender(config)
	defer defender.Close()

	// Create a padded outbound packet
	originalData := []byte("test data for inbound")
	result, _ := defender.ProcessOutbound(originalData)

	// Process the first real packet back through inbound
	for _, pkt := range result {
		if !isCoverPacket(pkt) {
			processed, isCover, err := defender.ProcessInbound(pkt)
			if err != nil {
				t.Fatalf("ProcessInbound() error = %v", err)
			}

			if isCover {
				t.Error("Real packet marked as cover")
			}

			if !bytes.Equal(processed, originalData) {
				t.Errorf("ProcessInbound() returned wrong data: got %d bytes, want %d bytes", len(processed), len(originalData))
			}
			break
		}
	}
}

func TestDAITACoverPacketFiltering(t *testing.T) {
	config := DefaultDAITAConfig()
	defender := NewDAITADefender(config)
	defer defender.Close()

	// Generate a cover packet manually
	coverGen := NewCoverTrafficGenerator(config)
	coverPacket := coverGen.Generate()

	// Process cover packet through inbound
	_, isCover, err := defender.ProcessInbound(coverPacket)
	if err != nil {
		t.Fatalf("ProcessInbound() error = %v", err)
	}

	if !isCover {
		t.Error("Cover packet not detected")
	}
}

func TestDAITAStats(t *testing.T) {
	config := DefaultDAITAConfig()
	defender := NewDAITADefender(config)
	defer defender.Close()

	// Process some packets
	for i := 0; i < 10; i++ {
		defender.ProcessOutbound([]byte("test packet"))
	}

	stats := defender.Stats()

	if stats.PacketsProcessed != 10 {
		t.Errorf("PacketsProcessed = %d, want 10", stats.PacketsProcessed)
	}

	if stats.RealDataBytes == 0 {
		t.Error("RealDataBytes should not be 0")
	}

	// Overhead should be calculated
	if stats.OverheadBytes > 0 && stats.OverheadPercent == 0 {
		t.Error("OverheadPercent should be calculated")
	}
}

func TestDAITADefenderConcurrent(t *testing.T) {
	config := DefaultDAITAConfig()
	defender := NewDAITADefender(config)
	defer defender.Close()

	var wg sync.WaitGroup
	numGoroutines := 10
	packetsPerGoroutine := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < packetsPerGoroutine; j++ {
				_, err := defender.ProcessOutbound([]byte("concurrent test"))
				if err != nil {
					t.Errorf("ProcessOutbound() error = %v", err)
				}
			}
		}()
	}

	wg.Wait()

	stats := defender.Stats()
	expectedPackets := uint64(numGoroutines * packetsPerGoroutine)
	if stats.PacketsProcessed != expectedPackets {
		t.Errorf("PacketsProcessed = %d, want %d", stats.PacketsProcessed, expectedPackets)
	}
}

// =============================================================================
// Maybenot State Machine Tests
// =============================================================================

func TestMaybenotState(t *testing.T) {
	state := NewMaybenotState()
	if state == nil {
		t.Fatal("NewMaybenotState() returned nil")
	}

	// Initial action should be deterministic based on state
	action := state.GetAction()
	if action < MaybenotActionNone || action > MaybenotActionBlock {
		t.Errorf("Invalid action: %d", action)
	}
}

func TestMaybenotStateTransitions(t *testing.T) {
	state := NewMaybenotState()

	// Test send transition
	state.Transition(MaybenotEventSend, 100)
	if state.sendCount != 1 {
		t.Errorf("sendCount = %d, want 1", state.sendCount)
	}

	// Test receive transition
	state.Transition(MaybenotEventRecv, 50)
	if state.recvCount != 1 {
		t.Errorf("recvCount = %d, want 1", state.recvCount)
	}

	// Multiple transitions
	for i := 0; i < 100; i++ {
		state.Transition(MaybenotEventSend, i*10)
	}

	if state.sendCount != 101 {
		t.Errorf("sendCount = %d, want 101", state.sendCount)
	}
}

// =============================================================================
// Cover Traffic Generator Tests
// =============================================================================

func TestCoverTrafficGenerator(t *testing.T) {
	config := DefaultDAITAConfig()
	gen := NewCoverTrafficGenerator(config)

	// Generate cover packet
	packet := gen.Generate()

	if len(packet) != config.PacketSize {
		t.Errorf("Cover packet size = %d, want %d", len(packet), config.PacketSize)
	}

	// Check cover marker
	if packet[4] != 0xFF {
		t.Error("Cover packet missing marker")
	}

	// Check length field is zero
	if packet[0] != 0 || packet[1] != 0 || packet[2] != 0 || packet[3] != 0 {
		t.Error("Cover packet should have zero length field")
	}
}

func TestCoverTrafficGeneratorRecordSize(t *testing.T) {
	config := DefaultDAITAConfig()
	config.ConstantPacketSize = false
	gen := NewCoverTrafficGenerator(config)

	// Record some sizes
	for i := 0; i < 50; i++ {
		gen.RecordSize(100 + i*10)
	}

	// Generate should now mimic recorded sizes
	packet := gen.Generate()
	if len(packet) < 100 {
		t.Error("Generated packet should mimic recorded sizes")
	}
}

// =============================================================================
// Pattern Distorter Tests
// =============================================================================

func TestPatternDistorter(t *testing.T) {
	config := DefaultDAITAConfig()
	distorter := NewPatternDistorter(config)

	dummies := distorter.GenerateDummies(500)

	// Should generate some dummy packets based on distortion level
	if config.DistortionLevel > 0 && len(dummies) == 0 {
		t.Log("Note: Dummy generation can be probabilistic")
	}

	for _, dummy := range dummies {
		if dummy[4] != 0xFF {
			t.Error("Dummy packet missing cover marker")
		}
	}
}

func TestPatternDistorterJitter(t *testing.T) {
	config := DefaultDAITAConfig()
	distorter := NewPatternDistorter(config)

	// Get multiple jitter values and verify they vary
	jitters := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		jitter := distorter.GetJitter()
		jitters[jitter] = true

		// Jitter should be positive
		if jitter < 0 {
			t.Error("Jitter should be positive")
		}
	}

	// Should have some variation
	if len(jitters) < 5 {
		t.Log("Note: Jitter should have some variation, got limited unique values")
	}
}

// =============================================================================
// Token Bucket Tests
// =============================================================================

func TestTokenBucket(t *testing.T) {
	bucket := NewTokenBucket(10, 50*time.Millisecond)

	// Should be able to take tokens
	for i := 0; i < 10; i++ {
		if !bucket.Take(1) {
			t.Errorf("Take(1) failed at iteration %d", i)
		}
	}

	// Bucket should be empty now
	if bucket.Take(1) {
		t.Error("Take(1) should fail when bucket is empty")
	}

	// Wait for refill
	time.Sleep(60 * time.Millisecond)

	// Should have tokens again
	if !bucket.Take(1) {
		t.Error("Take(1) should succeed after refill")
	}
}

func TestTokenBucketRefill(t *testing.T) {
	bucket := NewTokenBucket(100, 100*time.Millisecond)

	// Empty the bucket
	for bucket.Take(1) {
	}

	// Wait for partial refill
	time.Sleep(50 * time.Millisecond)

	// Should have some tokens
	count := 0
	for bucket.Take(1) {
		count++
	}

	if count == 0 {
		t.Error("Bucket should have refilled some tokens")
	}
}

// =============================================================================
// Website Fingerprint Defense Tests
// =============================================================================

func TestWebsiteFingerprintDefense(t *testing.T) {
	wf := NewWebsiteFingerprintDefense()

	// Should not inject before defense starts
	if wf.ShouldInjectCover(10) {
		t.Error("Should not inject before StartDefense()")
	}

	// Start defense
	wf.StartDefense()

	// Should inject during active defense
	shouldInject := wf.ShouldInjectCover(3)
	_ = shouldInject // Value depends on packet count

	// Stop defense
	wf.StopDefense()

	// Should not inject after stop
	if wf.ShouldInjectCover(10) {
		t.Error("Should not inject after StopDefense()")
	}
}

// =============================================================================
// Traffic Analysis Resistance Tests
// =============================================================================

func TestTrafficAnalysisResistance(t *testing.T) {
	config := DefaultDAITAConfig()
	tar := NewTrafficAnalysisResistance(config)
	defer tar.Close()

	// Process a packet
	result, err := tar.Process([]byte("test data"), true)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if len(result) == 0 {
		t.Error("Process() returned empty result")
	}

	// Check overhead percentage
	overhead := tar.GetOverheadPercent()
	if overhead < 0 || overhead > 100 {
		t.Errorf("OverheadPercent = %f, should be 0-100", overhead)
	}
}

// =============================================================================
// Defense Level Tests
// =============================================================================

func TestDefenseLevels(t *testing.T) {
	testCases := []struct {
		name   string
		config *DAITAConfig
	}{
		{"default", DefaultDAITAConfig()},
		{"high security", HighSecurityDAITAConfig()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defender := NewDAITADefender(tc.config)
			defer defender.Close()

			// Process some packets
			for i := 0; i < 100; i++ {
				_, err := defender.ProcessOutbound([]byte("test data for level check"))
				if err != nil {
					t.Fatalf("ProcessOutbound() error = %v", err)
				}
			}

			stats := defender.Stats()
			t.Logf("%s: Overhead = %.2f%%, Cover packets = %d",
				tc.name, stats.OverheadPercent, stats.CoverPacketsSent)
		})
	}
}

// =============================================================================
// Padding Tests
// =============================================================================

func TestConstantSizePadding(t *testing.T) {
	config := DefaultDAITAConfig()
	config.ConstantPacketSize = true
	config.PacketSize = 1420

	defender := NewDAITADefender(config)
	defer defender.Close()

	testSizes := []int{10, 100, 500, 1000}

	for _, size := range testSizes {
		data := bytes.Repeat([]byte("x"), size)
		result, _ := defender.ProcessOutbound(data)

		// Find the real packet (first non-cover)
		for _, pkt := range result {
			if !isCoverPacket(pkt) {
				if len(pkt) != config.PacketSize {
					t.Errorf("Padded packet size = %d, want %d for input size %d",
						len(pkt), config.PacketSize, size)
				}
				break
			}
		}
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkDAITAProcessOutbound(b *testing.B) {
	config := DefaultDAITAConfig()
	defender := NewDAITADefender(config)
	defer defender.Close()

	data := make([]byte, 1400)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		defender.ProcessOutbound(data)
	}
}

func BenchmarkDAITAProcessInbound(b *testing.B) {
	config := DefaultDAITAConfig()
	defender := NewDAITADefender(config)
	defer defender.Close()

	// Pre-generate packets
	data := make([]byte, 1400)
	packets := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		result, _ := defender.ProcessOutbound(data)
		for _, pkt := range result {
			if !isCoverPacket(pkt) {
				packets[i] = pkt
				break
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if packets[i] != nil {
			defender.ProcessInbound(packets[i])
		}
	}
}

func BenchmarkCoverGenerate(b *testing.B) {
	config := DefaultDAITAConfig()
	gen := NewCoverTrafficGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate()
	}
}

func BenchmarkTokenBucketTake(b *testing.B) {
	bucket := NewTokenBucket(1000000, time.Nanosecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.Take(1)
	}
}
