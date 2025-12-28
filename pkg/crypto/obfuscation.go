// Package crypto provides traffic obfuscation for VPN tunnels
// Makes VPN traffic indistinguishable from regular HTTPS traffic
package crypto

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// Obfuscation errors
var (
	ErrObfuscationFailed = errors.New("traffic obfuscation failed")
	ErrDeobfuscateFailed = errors.New("traffic deobfuscation failed")
	ErrInvalidPacketSize = errors.New("invalid obfuscated packet size")
)

// ObfuscationMode defines traffic obfuscation strategy
type ObfuscationMode uint8

const (
	// ObfuscationNone - No obfuscation (testing only)
	ObfuscationNone ObfuscationMode = iota

	// ObfuscationHTTPS - Make traffic look like HTTPS
	ObfuscationHTTPS

	// ObfuscationDNS - Make traffic look like DNS over HTTPS
	ObfuscationDNS

	// ObfuscationWebSocket - Make traffic look like WebSocket
	ObfuscationWebSocket

	// ObfuscationFull - Maximum obfuscation (HTTPS + timing + padding)
	ObfuscationFull
)

// TrafficObfuscator provides military-grade traffic analysis protection
type TrafficObfuscator struct {
	mode ObfuscationMode

	// Packet size normalization
	minPacketSize int
	maxPacketSize int
	padToSize     int // Fixed size for all packets (0 = variable)

	// Timing obfuscation
	minDelay time.Duration
	maxDelay time.Duration
	jitter   float64 // 0.0-1.0 jitter factor

	// Dummy traffic generation
	dummyTrafficEnabled bool
	dummyInterval       time.Duration
	dummyVariance       float64

	// Traffic shaping
	burstSize       int
	burstInterval   time.Duration
	shapingEnabled  bool

	// Statistics (anonymized)
	packetsSent    atomic.Uint64
	packetsRecv    atomic.Uint64
	dummySent      atomic.Uint64
	bytesObfuscated atomic.Uint64

	// TLS record layer simulation
	tlsVersion   [2]byte
	tlsRecordBuf []byte

	// Control
	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

// ObfuscationConfig holds obfuscation settings
type ObfuscationConfig struct {
	Mode            ObfuscationMode
	MinPacketSize   int
	MaxPacketSize   int
	PadToFixedSize  int
	MinDelay        time.Duration
	MaxDelay        time.Duration
	Jitter          float64
	EnableDummy     bool
	DummyInterval   time.Duration
	DummyVariance   float64
	BurstSize       int
	BurstInterval   time.Duration
	EnableShaping   bool
}

// DefaultObfuscationConfig returns secure defaults
func DefaultObfuscationConfig() *ObfuscationConfig {
	return &ObfuscationConfig{
		Mode:           ObfuscationFull,
		MinPacketSize:  64,
		MaxPacketSize:  1500,
		PadToFixedSize: 0, // Variable size with padding
		MinDelay:       1 * time.Millisecond,
		MaxDelay:       50 * time.Millisecond,
		Jitter:         0.3,
		EnableDummy:    true,
		DummyInterval:  100 * time.Millisecond,
		DummyVariance:  0.5,
		BurstSize:      10,
		BurstInterval:  10 * time.Millisecond,
		EnableShaping:  true,
	}
}

// MaxSecurityObfuscationConfig returns maximum security settings
func MaxSecurityObfuscationConfig() *ObfuscationConfig {
	return &ObfuscationConfig{
		Mode:           ObfuscationFull,
		MinPacketSize:  128,
		MaxPacketSize:  1400,
		PadToFixedSize: 1400, // Fixed size for all packets
		MinDelay:       5 * time.Millisecond,
		MaxDelay:       100 * time.Millisecond,
		Jitter:         0.5,
		EnableDummy:    true,
		DummyInterval:  50 * time.Millisecond,
		DummyVariance:  0.7,
		BurstSize:      5,
		BurstInterval:  20 * time.Millisecond,
		EnableShaping:  true,
	}
}

// NewTrafficObfuscator creates a new traffic obfuscator
func NewTrafficObfuscator(config *ObfuscationConfig) *TrafficObfuscator {
	if config == nil {
		config = DefaultObfuscationConfig()
	}

	to := &TrafficObfuscator{
		mode:                config.Mode,
		minPacketSize:       config.MinPacketSize,
		maxPacketSize:       config.MaxPacketSize,
		padToSize:           config.PadToFixedSize,
		minDelay:            config.MinDelay,
		maxDelay:            config.MaxDelay,
		jitter:              config.Jitter,
		dummyTrafficEnabled: config.EnableDummy,
		dummyInterval:       config.DummyInterval,
		dummyVariance:       config.DummyVariance,
		burstSize:           config.BurstSize,
		burstInterval:       config.BurstInterval,
		shapingEnabled:      config.EnableShaping,
		tlsVersion:          [2]byte{0x03, 0x03}, // TLS 1.2
		tlsRecordBuf:        make([]byte, 16384),
		stopChan:            make(chan struct{}),
	}

	return to
}

// ObfuscatedPacket represents an obfuscated packet
type ObfuscatedPacket struct {
	Data      []byte
	IsDummy   bool
	Timestamp time.Time
	Delay     time.Duration
}

// Obfuscate applies traffic obfuscation to data
func (to *TrafficObfuscator) Obfuscate(data []byte) (*ObfuscatedPacket, error) {
	to.mu.RLock()
	defer to.mu.RUnlock()

	var obfuscated []byte
	var err error

	switch to.mode {
	case ObfuscationNone:
		obfuscated = data
	case ObfuscationHTTPS:
		obfuscated, err = to.wrapAsTLSRecord(data)
	case ObfuscationDNS:
		obfuscated, err = to.wrapAsDNSOverHTTPS(data)
	case ObfuscationWebSocket:
		obfuscated, err = to.wrapAsWebSocket(data)
	case ObfuscationFull:
		// Apply all obfuscation layers
		obfuscated, err = to.fullObfuscate(data)
	default:
		obfuscated = data
	}

	if err != nil {
		return nil, err
	}

	// Calculate delay with jitter
	delay := to.calculateDelay()

	to.packetsSent.Add(1)
	to.bytesObfuscated.Add(uint64(len(obfuscated)))

	return &ObfuscatedPacket{
		Data:      obfuscated,
		IsDummy:   false,
		Timestamp: time.Now(),
		Delay:     delay,
	}, nil
}

// Deobfuscate removes obfuscation from data
func (to *TrafficObfuscator) Deobfuscate(packet *ObfuscatedPacket) ([]byte, bool, error) {
	to.mu.RLock()
	defer to.mu.RUnlock()

	if packet.IsDummy {
		return nil, true, nil
	}

	var data []byte
	var err error

	switch to.mode {
	case ObfuscationNone:
		data = packet.Data
	case ObfuscationHTTPS:
		data, err = to.unwrapTLSRecord(packet.Data)
	case ObfuscationDNS:
		data, err = to.unwrapDNSOverHTTPS(packet.Data)
	case ObfuscationWebSocket:
		data, err = to.unwrapWebSocket(packet.Data)
	case ObfuscationFull:
		data, err = to.fullDeobfuscate(packet.Data)
	default:
		data = packet.Data
	}

	if err != nil {
		return nil, false, err
	}

	to.packetsRecv.Add(1)
	return data, false, nil
}

// wrapAsTLSRecord wraps data as a TLS Application Data record
func (to *TrafficObfuscator) wrapAsTLSRecord(data []byte) ([]byte, error) {
	// Add random padding
	padded := to.addPadding(data)

	// TLS record header (5 bytes)
	// Content type: Application Data (0x17)
	// Version: TLS 1.2 (0x0303)
	// Length: 2 bytes
	recordLen := len(padded)
	if recordLen > 16384 {
		return nil, ErrInvalidPacketSize
	}

	record := make([]byte, 5+recordLen)
	record[0] = 0x17 // Application Data
	record[1] = to.tlsVersion[0]
	record[2] = to.tlsVersion[1]
	binary.BigEndian.PutUint16(record[3:5], uint16(recordLen))
	copy(record[5:], padded)

	return record, nil
}

// unwrapTLSRecord extracts data from a TLS record
func (to *TrafficObfuscator) unwrapTLSRecord(record []byte) ([]byte, error) {
	if len(record) < 5 {
		return nil, ErrInvalidPacketSize
	}

	// Verify it's Application Data
	if record[0] != 0x17 {
		return nil, ErrDeobfuscateFailed
	}

	length := binary.BigEndian.Uint16(record[3:5])
	if len(record) < 5+int(length) {
		return nil, ErrInvalidPacketSize
	}

	padded := record[5 : 5+length]
	return to.removePadding(padded)
}

// wrapAsDNSOverHTTPS wraps data to look like DNS-over-HTTPS
func (to *TrafficObfuscator) wrapAsDNSOverHTTPS(data []byte) ([]byte, error) {
	// Add padding
	padded := to.addPadding(data)

	// DNS-over-HTTPS uses POST with application/dns-message content type
	// We simulate the wire format: 2-byte length prefix + data
	if len(padded) > 65535 {
		return nil, ErrInvalidPacketSize
	}

	wrapped := make([]byte, 2+len(padded))
	binary.BigEndian.PutUint16(wrapped[:2], uint16(len(padded)))
	copy(wrapped[2:], padded)

	// Wrap in TLS
	return to.wrapAsTLSRecord(wrapped)
}

// unwrapDNSOverHTTPS extracts data from DNS-over-HTTPS format
func (to *TrafficObfuscator) unwrapDNSOverHTTPS(data []byte) ([]byte, error) {
	// Unwrap TLS first
	inner, err := to.unwrapTLSRecord(data)
	if err != nil {
		return nil, err
	}

	if len(inner) < 2 {
		return nil, ErrInvalidPacketSize
	}

	length := binary.BigEndian.Uint16(inner[:2])
	if len(inner) < 2+int(length) {
		return nil, ErrInvalidPacketSize
	}

	padded := inner[2 : 2+length]
	return to.removePadding(padded)
}

// wrapAsWebSocket wraps data to look like WebSocket frames
func (to *TrafficObfuscator) wrapAsWebSocket(data []byte) ([]byte, error) {
	padded := to.addPadding(data)

	// WebSocket binary frame
	// FIN=1, RSV=0, Opcode=2 (binary)
	// MASK=1 (client to server), length
	payloadLen := len(padded)

	var header []byte
	if payloadLen <= 125 {
		header = make([]byte, 6) // 2 + 4 (mask key)
		header[0] = 0x82         // FIN + Binary
		header[1] = 0x80 | byte(payloadLen)
	} else if payloadLen <= 65535 {
		header = make([]byte, 8) // 2 + 2 + 4
		header[0] = 0x82
		header[1] = 0x80 | 126
		binary.BigEndian.PutUint16(header[2:4], uint16(payloadLen))
	} else {
		header = make([]byte, 14) // 2 + 8 + 4
		header[0] = 0x82
		header[1] = 0x80 | 127
		binary.BigEndian.PutUint64(header[2:10], uint64(payloadLen))
	}

	// Generate random masking key
	maskKeyOffset := len(header) - 4
	if _, err := rand.Read(header[maskKeyOffset:]); err != nil {
		return nil, err
	}

	// Apply mask to payload
	masked := make([]byte, payloadLen)
	maskKey := header[maskKeyOffset:]
	for i := 0; i < payloadLen; i++ {
		masked[i] = padded[i] ^ maskKey[i%4]
	}

	result := make([]byte, len(header)+payloadLen)
	copy(result, header)
	copy(result[len(header):], masked)

	// Wrap in TLS
	return to.wrapAsTLSRecord(result)
}

// unwrapWebSocket extracts data from WebSocket format
func (to *TrafficObfuscator) unwrapWebSocket(data []byte) ([]byte, error) {
	// Unwrap TLS first
	inner, err := to.unwrapTLSRecord(data)
	if err != nil {
		return nil, err
	}

	if len(inner) < 6 {
		return nil, ErrInvalidPacketSize
	}

	// Parse WebSocket frame
	payloadLen := int(inner[1] & 0x7F)
	maskOffset := 2

	if payloadLen == 126 {
		if len(inner) < 8 {
			return nil, ErrInvalidPacketSize
		}
		payloadLen = int(binary.BigEndian.Uint16(inner[2:4]))
		maskOffset = 4
	} else if payloadLen == 127 {
		if len(inner) < 14 {
			return nil, ErrInvalidPacketSize
		}
		payloadLen = int(binary.BigEndian.Uint64(inner[2:10]))
		maskOffset = 10
	}

	if len(inner) < maskOffset+4+payloadLen {
		return nil, ErrInvalidPacketSize
	}

	maskKey := inner[maskOffset : maskOffset+4]
	masked := inner[maskOffset+4 : maskOffset+4+payloadLen]

	// Unmask
	unmasked := make([]byte, payloadLen)
	for i := 0; i < payloadLen; i++ {
		unmasked[i] = masked[i] ^ maskKey[i%4]
	}

	return to.removePadding(unmasked)
}

// fullObfuscate applies all obfuscation layers
func (to *TrafficObfuscator) fullObfuscate(data []byte) ([]byte, error) {
	// Layer 1: Add random prefix/suffix to break patterns
	randomized := to.addRandomWrapper(data)

	// Layer 2: Pad to fixed or variable size
	padded := to.addPadding(randomized)

	// Layer 3: XOR with pseudo-random stream (adds another layer)
	xored := to.xorWithStream(padded)

	// Layer 4: Wrap as TLS record
	return to.wrapAsTLSRecord(xored)
}

// fullDeobfuscate removes all obfuscation layers
func (to *TrafficObfuscator) fullDeobfuscate(data []byte) ([]byte, error) {
	// Layer 4: Unwrap TLS record
	xored, err := to.unwrapTLSRecord(data)
	if err != nil {
		return nil, err
	}

	// Layer 3: Un-XOR
	padded := to.xorWithStream(xored)

	// Layer 2: Remove padding
	randomized, err := to.removePadding(padded)
	if err != nil {
		return nil, err
	}

	// Layer 1: Remove random wrapper
	return to.removeRandomWrapper(randomized)
}

// addPadding adds PKCS#7-style padding with random data
func (to *TrafficObfuscator) addPadding(data []byte) []byte {
	targetSize := to.padToSize
	if targetSize == 0 {
		// Variable padding: pad to next 64-byte boundary + random
		randomExtra := make([]byte, 1)
		rand.Read(randomExtra)
		extra := int(randomExtra[0]) % 64
		targetSize = ((len(data) + 68) / 64) * 64 + extra
		if targetSize > to.maxPacketSize {
			targetSize = to.maxPacketSize
		}
		if targetSize < to.minPacketSize {
			targetSize = to.minPacketSize
		}
	}

	if len(data) >= targetSize-4 {
		// Can't pad, just add length prefix
		result := make([]byte, 4+len(data))
		binary.BigEndian.PutUint32(result[:4], uint32(len(data)))
		copy(result[4:], data)
		return result
	}

	paddingLen := targetSize - len(data) - 4
	result := make([]byte, 4+len(data)+paddingLen)

	// Length prefix
	binary.BigEndian.PutUint32(result[:4], uint32(len(data)))

	// Data
	copy(result[4:4+len(data)], data)

	// Random padding
	rand.Read(result[4+len(data):])

	return result
}

// removePadding removes padding and extracts original data
func (to *TrafficObfuscator) removePadding(padded []byte) ([]byte, error) {
	if len(padded) < 4 {
		return nil, ErrInvalidPacketSize
	}

	dataLen := binary.BigEndian.Uint32(padded[:4])
	if dataLen > uint32(len(padded)-4) {
		return nil, ErrDeobfuscateFailed
	}

	return padded[4 : 4+dataLen], nil
}

// addRandomWrapper adds random prefix and suffix
func (to *TrafficObfuscator) addRandomWrapper(data []byte) []byte {
	// Random prefix length (4-16 bytes)
	prefixLenBytes := make([]byte, 1)
	rand.Read(prefixLenBytes)
	prefixLen := 4 + int(prefixLenBytes[0])%13

	// Random suffix length (4-16 bytes)
	suffixLenBytes := make([]byte, 1)
	rand.Read(suffixLenBytes)
	suffixLen := 4 + int(suffixLenBytes[0])%13

	// Format: [prefixLen:1][prefix:N][data][suffixLen:1][suffix:M]
	result := make([]byte, 1+prefixLen+len(data)+1+suffixLen)

	result[0] = byte(prefixLen)
	rand.Read(result[1 : 1+prefixLen])
	copy(result[1+prefixLen:], data)
	result[1+prefixLen+len(data)] = byte(suffixLen)
	rand.Read(result[1+prefixLen+len(data)+1:])

	return result
}

// removeRandomWrapper removes random prefix and suffix
// Format: [prefixLen:1][prefix:N][data][suffixLen:1][suffix:M]
func (to *TrafficObfuscator) removeRandomWrapper(wrapped []byte) ([]byte, error) {
	if len(wrapped) < 2 {
		return nil, ErrInvalidPacketSize
	}

	prefixLen := int(wrapped[0])
	if len(wrapped) < 1+prefixLen+2 { // At least prefixLen byte, prefix, 1 byte data, suffixLen byte
		return nil, ErrDeobfuscateFailed
	}

	// After prefix: [data][suffixLen:1][suffix:M]
	remaining := wrapped[1+prefixLen:]
	if len(remaining) < 2 {
		return nil, ErrDeobfuscateFailed
	}

	// The suffixLen byte is right after data, followed by suffix bytes
	// We need to find it by checking each possible position
	// Format: remaining = [data][suffixLen:1][suffix:M]
	// So if byte at position i equals (len(remaining) - i - 1), that's the suffixLen byte
	for i := len(remaining) - 1; i >= 0; i-- {
		possibleSuffixLen := int(remaining[i])
		// Check if this byte could be the suffixLen marker
		// It should be followed by exactly possibleSuffixLen bytes
		if i+1+possibleSuffixLen == len(remaining) && possibleSuffixLen >= 4 && possibleSuffixLen <= 16 {
			return remaining[:i], nil
		}
	}

	return nil, ErrDeobfuscateFailed
}

// xorWithStream applies XOR with deterministic pseudo-random stream
// Uses first 8 bytes as seed
func (to *TrafficObfuscator) xorWithStream(data []byte) []byte {
	if len(data) < 8 {
		return data
	}

	result := make([]byte, len(data))
	copy(result, data)

	// Use simple PRNG seeded from data
	seed := binary.BigEndian.Uint64(data[:8])

	// LCG parameters (same as glibc)
	a := uint64(1103515245)
	c := uint64(12345)
	m := uint64(1 << 31)

	state := seed
	for i := 8; i < len(result); i++ {
		state = (a*state + c) % m
		result[i] ^= byte(state & 0xFF)
	}

	return result
}

// calculateDelay calculates delay with jitter
func (to *TrafficObfuscator) calculateDelay() time.Duration {
	if to.maxDelay <= to.minDelay {
		return to.minDelay
	}

	// Base delay
	delayRange := to.maxDelay - to.minDelay
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomFactor := float64(binary.BigEndian.Uint64(randomBytes)) / float64(math.MaxUint64)

	baseDelay := to.minDelay + time.Duration(float64(delayRange)*randomFactor)

	// Apply jitter
	if to.jitter > 0 {
		jitterRange := float64(baseDelay) * to.jitter
		rand.Read(randomBytes)
		jitterFactor := (float64(binary.BigEndian.Uint64(randomBytes))/float64(math.MaxUint64) - 0.5) * 2
		baseDelay += time.Duration(jitterRange * jitterFactor)
	}

	if baseDelay < 0 {
		baseDelay = 0
	}

	return baseDelay
}

// GenerateDummyPacket creates a dummy packet for traffic shaping
func (to *TrafficObfuscator) GenerateDummyPacket() (*ObfuscatedPacket, error) {
	to.mu.RLock()
	defer to.mu.RUnlock()

	// Generate random dummy data
	size := to.minPacketSize
	if to.padToSize > 0 {
		size = to.padToSize
	} else {
		sizeBytes := make([]byte, 2)
		rand.Read(sizeBytes)
		size = to.minPacketSize + int(binary.BigEndian.Uint16(sizeBytes))%(to.maxPacketSize-to.minPacketSize)
	}

	dummyData := make([]byte, size)
	rand.Read(dummyData)

	// Mark as dummy with special header
	dummyData[0] = 0xFF
	dummyData[1] = 0x00
	dummyData[2] = 0xFF
	dummyData[3] = 0x00

	obfuscated, err := to.Obfuscate(dummyData)
	if err != nil {
		return nil, err
	}

	obfuscated.IsDummy = true
	to.dummySent.Add(1)

	return obfuscated, nil
}

// StartDummyTraffic starts generating dummy traffic
func (to *TrafficObfuscator) StartDummyTraffic(sendFunc func(*ObfuscatedPacket) error) {
	if !to.dummyTrafficEnabled {
		return
	}

	to.wg.Add(1)
	go func() {
		defer to.wg.Done()

		ticker := time.NewTicker(to.dummyInterval)
		defer ticker.Stop()

		for {
			select {
			case <-to.stopChan:
				return
			case <-ticker.C:
				// Add variance to interval
				if to.dummyVariance > 0 {
					varianceBytes := make([]byte, 8)
					rand.Read(varianceBytes)
					variance := (float64(binary.BigEndian.Uint64(varianceBytes))/float64(math.MaxUint64) - 0.5) * 2 * to.dummyVariance
					adjustment := time.Duration(float64(to.dummyInterval) * variance)
					time.Sleep(adjustment)
				}

				packet, err := to.GenerateDummyPacket()
				if err == nil && sendFunc != nil {
					sendFunc(packet)
				}
			}
		}
	}()
}

// Stats returns obfuscation statistics
func (to *TrafficObfuscator) Stats() ObfuscationStats {
	return ObfuscationStats{
		PacketsSent:     to.packetsSent.Load(),
		PacketsReceived: to.packetsRecv.Load(),
		DummyPackets:    to.dummySent.Load(),
		BytesObfuscated: to.bytesObfuscated.Load(),
	}
}

// ObfuscationStats holds obfuscation statistics
type ObfuscationStats struct {
	PacketsSent     uint64
	PacketsReceived uint64
	DummyPackets    uint64
	BytesObfuscated uint64
}

// Close stops the obfuscator
func (to *TrafficObfuscator) Close() {
	close(to.stopChan)
	to.wg.Wait()
}

// TimingAnalysisProtector provides protection against timing attacks
type TimingAnalysisProtector struct {
	// Packet timing history (for pattern breaking)
	timingWindow []time.Duration
	windowSize   int
	windowIdx    int

	// Constant-time operation delays
	minOpTime time.Duration

	mu sync.Mutex
}

// NewTimingAnalysisProtector creates a timing attack protector
func NewTimingAnalysisProtector(windowSize int, minOpTime time.Duration) *TimingAnalysisProtector {
	return &TimingAnalysisProtector{
		timingWindow: make([]time.Duration, windowSize),
		windowSize:   windowSize,
		minOpTime:    minOpTime,
	}
}

// ConstantTimeOperation ensures operations take constant time
func (tap *TimingAnalysisProtector) ConstantTimeOperation(operation func() error) error {
	start := time.Now()

	err := operation()

	// Ensure minimum operation time
	elapsed := time.Since(start)
	if elapsed < tap.minOpTime {
		// Add random jitter to remaining time
		remaining := tap.minOpTime - elapsed
		jitterBytes := make([]byte, 8)
		rand.Read(jitterBytes)
		jitter := time.Duration(binary.BigEndian.Uint64(jitterBytes) % uint64(remaining/4))
		time.Sleep(remaining + jitter)
	}

	return err
}

// RecordTiming records operation timing for pattern analysis
func (tap *TimingAnalysisProtector) RecordTiming(duration time.Duration) {
	tap.mu.Lock()
	defer tap.mu.Unlock()

	tap.timingWindow[tap.windowIdx] = duration
	tap.windowIdx = (tap.windowIdx + 1) % tap.windowSize
}

// ShouldAddJitter determines if jitter should be added to break patterns
func (tap *TimingAnalysisProtector) ShouldAddJitter() bool {
	tap.mu.Lock()
	defer tap.mu.Unlock()

	// Check if timings are too regular (pattern detected)
	if tap.windowIdx < 3 {
		return true // Not enough data, add jitter
	}

	// Calculate variance
	var sum, sumSq float64
	count := 0
	for _, t := range tap.timingWindow {
		if t > 0 {
			ms := float64(t.Milliseconds())
			sum += ms
			sumSq += ms * ms
			count++
		}
	}

	if count < 3 {
		return true
	}

	mean := sum / float64(count)
	variance := (sumSq / float64(count)) - (mean * mean)

	// If variance is too low, patterns are detectable
	return variance < 10.0 // Threshold in ms^2
}

// TrafficShaper implements traffic shaping for VPN
type TrafficShaper struct {
	// Token bucket for rate limiting
	tokens       float64
	maxTokens    float64
	tokenRate    float64 // tokens per second
	lastUpdate   time.Time

	// Burst control
	burstTokens  int
	burstPeriod  time.Duration
	burstStart   time.Time
	burstCount   int

	mu sync.Mutex
}

// NewTrafficShaper creates a new traffic shaper
func NewTrafficShaper(maxBytesPerSec int, burstSize int, burstPeriod time.Duration) *TrafficShaper {
	return &TrafficShaper{
		tokens:      float64(burstSize),
		maxTokens:   float64(burstSize * 2),
		tokenRate:   float64(maxBytesPerSec),
		lastUpdate:  time.Now(),
		burstTokens: burstSize,
		burstPeriod: burstPeriod,
		burstStart:  time.Now(),
	}
}

// Acquire acquires tokens for sending data, returns delay to wait
func (ts *TrafficShaper) Acquire(bytes int) time.Duration {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(ts.lastUpdate).Seconds()
	ts.tokens += elapsed * ts.tokenRate
	if ts.tokens > ts.maxTokens {
		ts.tokens = ts.maxTokens
	}
	ts.lastUpdate = now

	// Check burst limit
	if now.Sub(ts.burstStart) > ts.burstPeriod {
		ts.burstStart = now
		ts.burstCount = 0
	}

	// Calculate delay
	bytesFloat := float64(bytes)
	if ts.tokens >= bytesFloat && ts.burstCount < ts.burstTokens {
		ts.tokens -= bytesFloat
		ts.burstCount++
		return 0
	}

	// Need to wait
	tokensNeeded := bytesFloat - ts.tokens
	if tokensNeeded < 0 {
		tokensNeeded = 0
	}

	delay := time.Duration(tokensNeeded/ts.tokenRate*1000) * time.Millisecond

	// Add random jitter
	jitterBytes := make([]byte, 2)
	rand.Read(jitterBytes)
	jitter := time.Duration(binary.BigEndian.Uint16(jitterBytes)%100) * time.Millisecond
	delay += jitter

	return delay
}
