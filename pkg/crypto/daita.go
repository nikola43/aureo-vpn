// Package crypto provides DAITA - Defense Against AI-guided Traffic Analysis
// Inspired by Mullvad VPN's implementation, enhanced with additional protections
// This counters machine learning-based traffic fingerprinting attacks
package crypto

import (
	"crypto/rand"
	"encoding/binary"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// DAITA configuration constants
const (
	// Packet size normalization
	DAITAPacketSize     = 1420 // Fixed MTU-safe packet size
	DAITAMinPadding     = 64
	DAITAMaxPadding     = 512

	// Cover traffic parameters
	DAITACoverMinInterval = 10 * time.Millisecond
	DAITACoverMaxInterval = 100 * time.Millisecond
	DAITACoverBurstSize   = 5

	// Traffic shaping
	DAITABucketSize    = 100
	DAITABucketRefill  = 50 * time.Millisecond

	// Maybenot-style defense framework
	MaybenotStateSize = 16
)

// DAITAConfig holds DAITA configuration
type DAITAConfig struct {
	// Enable constant packet sizes
	ConstantPacketSize bool
	PacketSize         int

	// Enable cover traffic injection
	CoverTraffic       bool
	CoverMinInterval   time.Duration
	CoverMaxInterval   time.Duration
	CoverBurstSize     int

	// Enable traffic pattern distortion
	PatternDistortion  bool
	DistortionLevel    float64 // 0.0-1.0

	// Bandwidth overhead limit (percentage)
	MaxOverhead        float64

	// Machine learning defense level
	DefenseLevel       DAITADefenseLevel
}

// DAITADefenseLevel defines defense intensity
type DAITADefenseLevel int

const (
	DAITALevelNone DAITADefenseLevel = iota
	DAITALevelLight   // ~10% overhead, basic protection
	DAITALevelMedium  // ~30% overhead, strong protection
	DAITALevelStrong  // ~50% overhead, maximum protection
)

// DefaultDAITAConfig returns secure defaults
func DefaultDAITAConfig() *DAITAConfig {
	return &DAITAConfig{
		ConstantPacketSize: true,
		PacketSize:         DAITAPacketSize,
		CoverTraffic:       true,
		CoverMinInterval:   DAITACoverMinInterval,
		CoverMaxInterval:   DAITACoverMaxInterval,
		CoverBurstSize:     DAITACoverBurstSize,
		PatternDistortion:  true,
		DistortionLevel:    0.5,
		MaxOverhead:        0.30, // 30% max bandwidth overhead
		DefenseLevel:       DAITALevelMedium,
	}
}

// HighSecurityDAITAConfig returns maximum protection settings
func HighSecurityDAITAConfig() *DAITAConfig {
	return &DAITAConfig{
		ConstantPacketSize: true,
		PacketSize:         DAITAPacketSize,
		CoverTraffic:       true,
		CoverMinInterval:   5 * time.Millisecond,
		CoverMaxInterval:   50 * time.Millisecond,
		CoverBurstSize:     10,
		PatternDistortion:  true,
		DistortionLevel:    0.8,
		MaxOverhead:        0.50, // 50% max overhead for maximum protection
		DefenseLevel:       DAITALevelStrong,
	}
}

// DAITADefender implements Defense Against AI-guided Traffic Analysis
type DAITADefender struct {
	config *DAITAConfig

	// Maybenot-style state machine for defense
	state *MaybenotState

	// Cover traffic generation
	coverGenerator *CoverTrafficGenerator

	// Pattern distortion
	patternDistorter *PatternDistorter

	// Token bucket for rate limiting overhead
	tokenBucket *TokenBucket

	// Statistics
	stats DAITAStats

	// Control
	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

// DAITAStats holds defense statistics
type DAITAStats struct {
	PacketsProcessed   atomic.Uint64
	CoverPacketsSent   atomic.Uint64
	BytesPadded        atomic.Uint64
	PatternsDistorted  atomic.Uint64
	OverheadBytes      atomic.Uint64
	RealDataBytes      atomic.Uint64
}

// NewDAITADefender creates a new DAITA defender
func NewDAITADefender(config *DAITAConfig) *DAITADefender {
	if config == nil {
		config = DefaultDAITAConfig()
	}

	d := &DAITADefender{
		config:           config,
		state:            NewMaybenotState(),
		coverGenerator:   NewCoverTrafficGenerator(config),
		patternDistorter: NewPatternDistorter(config),
		tokenBucket:      NewTokenBucket(DAITABucketSize, DAITABucketRefill),
		stopChan:         make(chan struct{}),
	}

	return d
}

// ProcessOutbound processes an outbound packet with DAITA protections
func (d *DAITADefender) ProcessOutbound(packet []byte) ([][]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.stats.PacketsProcessed.Add(1)
	d.stats.RealDataBytes.Add(uint64(len(packet)))

	var result [][]byte

	// Step 1: Constant packet size padding
	if d.config.ConstantPacketSize {
		packet = d.padToConstantSize(packet)
	}

	// Step 2: Pattern distortion (inject dummy packets)
	if d.config.PatternDistortion {
		dummyPackets := d.patternDistorter.GenerateDummies(len(packet))
		for _, dummy := range dummyPackets {
			if d.tokenBucket.Take(1) {
				result = append(result, dummy)
				d.stats.CoverPacketsSent.Add(1)
				d.stats.OverheadBytes.Add(uint64(len(dummy)))
			}
		}
	}

	// Step 3: Add the real packet
	result = append(result, packet)

	// Step 4: Maybenot state transition
	d.state.Transition(MaybenotEventSend, len(packet))

	// Step 5: Generate cover traffic based on state
	if d.config.CoverTraffic {
		coverPackets := d.generateStateCover()
		for _, cover := range coverPackets {
			if d.tokenBucket.Take(1) {
				result = append(result, cover)
				d.stats.CoverPacketsSent.Add(1)
				d.stats.OverheadBytes.Add(uint64(len(cover)))
			}
		}
	}

	d.stats.PatternsDistorted.Add(1)
	return result, nil
}

// ProcessInbound processes an inbound packet, filtering cover traffic
func (d *DAITADefender) ProcessInbound(packet []byte) ([]byte, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if this is a cover packet (marked with special header)
	if len(packet) >= 4 && isCoverPacket(packet) {
		return nil, true, nil // Is cover packet, discard
	}

	// Remove padding if constant size was used
	if d.config.ConstantPacketSize {
		packet = d.removePadding(packet)
	}

	// Update state
	d.state.Transition(MaybenotEventRecv, len(packet))

	return packet, false, nil
}

// padToConstantSize pads packet to constant size
func (d *DAITADefender) padToConstantSize(packet []byte) []byte {
	targetSize := d.config.PacketSize
	if len(packet) >= targetSize-8 { // Reserve space for length header
		// Packet too large, fragment or pass through
		return d.addLengthHeader(packet)
	}

	// Create padded packet
	result := make([]byte, targetSize)

	// Length header (4 bytes)
	binary.BigEndian.PutUint32(result[:4], uint32(len(packet)))

	// Marker byte (1 byte) - 0x00 for real data
	result[4] = 0x00

	// Random padding type marker (3 bytes for alignment)
	rand.Read(result[5:8])

	// Copy actual data
	copy(result[8:], packet)

	// Fill remaining with random padding
	rand.Read(result[8+len(packet):])

	d.stats.BytesPadded.Add(uint64(targetSize - len(packet) - 8))

	return result
}

// addLengthHeader adds a length header to oversized packets
func (d *DAITADefender) addLengthHeader(packet []byte) []byte {
	result := make([]byte, 8+len(packet))
	binary.BigEndian.PutUint32(result[:4], uint32(len(packet)))
	result[4] = 0x00 // Real data marker
	rand.Read(result[5:8])
	copy(result[8:], packet)
	return result
}

// removePadding removes constant size padding
func (d *DAITADefender) removePadding(packet []byte) []byte {
	if len(packet) < 8 {
		return packet
	}

	length := binary.BigEndian.Uint32(packet[:4])
	marker := packet[4]

	if marker != 0x00 { // Not a real data packet
		return packet
	}

	if length > uint32(len(packet)-8) {
		return packet // Invalid length
	}

	return packet[8 : 8+length]
}

// generateStateCover generates cover traffic based on Maybenot state
func (d *DAITADefender) generateStateCover() [][]byte {
	action := d.state.GetAction()

	var packets [][]byte

	switch action {
	case MaybenotActionSendCover:
		// Generate single cover packet
		cover := d.coverGenerator.Generate()
		packets = append(packets, cover)

	case MaybenotActionSendBurst:
		// Generate burst of cover packets
		burstSize := d.config.CoverBurstSize
		for i := 0; i < burstSize; i++ {
			cover := d.coverGenerator.Generate()
			packets = append(packets, cover)
		}

	case MaybenotActionDelay:
		// Add timing jitter
		jitter := d.patternDistorter.GetJitter()
		time.Sleep(jitter)
	}

	return packets
}

// StartCoverTrafficLoop starts continuous cover traffic generation
func (d *DAITADefender) StartCoverTrafficLoop(sendFunc func([]byte) error) {
	if !d.config.CoverTraffic {
		return
	}

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()

		for {
			select {
			case <-d.stopChan:
				return
			default:
				// Random interval between cover packets
				interval := d.randomInterval()
				time.Sleep(interval)

				if d.tokenBucket.Take(1) {
					cover := d.coverGenerator.Generate()
					if sendFunc != nil {
						sendFunc(cover)
						d.stats.CoverPacketsSent.Add(1)
					}
				}
			}
		}
	}()
}

func (d *DAITADefender) randomInterval() time.Duration {
	min := d.config.CoverMinInterval
	max := d.config.CoverMaxInterval

	randBytes := make([]byte, 8)
	rand.Read(randBytes)
	randVal := binary.BigEndian.Uint64(randBytes)

	rangeVal := uint64(max - min)
	interval := min + time.Duration(randVal%rangeVal)

	return interval
}

// Stats returns DAITA statistics
func (d *DAITADefender) Stats() DAITAStatsSummary {
	realBytes := d.stats.RealDataBytes.Load()
	overheadBytes := d.stats.OverheadBytes.Load()

	var overheadPercent float64
	if realBytes > 0 {
		overheadPercent = float64(overheadBytes) / float64(realBytes) * 100
	}

	return DAITAStatsSummary{
		PacketsProcessed:  d.stats.PacketsProcessed.Load(),
		CoverPacketsSent:  d.stats.CoverPacketsSent.Load(),
		BytesPadded:       d.stats.BytesPadded.Load(),
		PatternsDistorted: d.stats.PatternsDistorted.Load(),
		OverheadBytes:     overheadBytes,
		RealDataBytes:     realBytes,
		OverheadPercent:   overheadPercent,
	}
}

// DAITAStatsSummary provides a summary of DAITA statistics
type DAITAStatsSummary struct {
	PacketsProcessed  uint64
	CoverPacketsSent  uint64
	BytesPadded       uint64
	PatternsDistorted uint64
	OverheadBytes     uint64
	RealDataBytes     uint64
	OverheadPercent   float64
}

// Close stops the DAITA defender
func (d *DAITADefender) Close() {
	close(d.stopChan)
	d.wg.Wait()
}

// isCoverPacket checks if a packet is a cover packet
func isCoverPacket(packet []byte) bool {
	if len(packet) < 8 {
		return false
	}
	// Cover packets have marker byte 0xFF
	return packet[4] == 0xFF
}

// MaybenotState implements the Maybenot defense framework state machine
type MaybenotState struct {
	state      [MaybenotStateSize]byte
	lastAction MaybenotAction
	lastEvent  MaybenotEvent
	sendCount  uint64
	recvCount  uint64
	mu         sync.Mutex
}

// MaybenotEvent represents network events
type MaybenotEvent int

const (
	MaybenotEventNone MaybenotEvent = iota
	MaybenotEventSend
	MaybenotEventRecv
	MaybenotEventTimer
)

// MaybenotAction represents defense actions
type MaybenotAction int

const (
	MaybenotActionNone MaybenotAction = iota
	MaybenotActionSendCover
	MaybenotActionSendBurst
	MaybenotActionDelay
	MaybenotActionBlock
)

// NewMaybenotState creates a new Maybenot state machine
func NewMaybenotState() *MaybenotState {
	s := &MaybenotState{}
	rand.Read(s.state[:])
	return s
}

// Transition updates state based on event
func (s *MaybenotState) Transition(event MaybenotEvent, dataSize int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastEvent = event

	switch event {
	case MaybenotEventSend:
		s.sendCount++
		// Update state based on send pattern
		s.updateStateOnSend(dataSize)

	case MaybenotEventRecv:
		s.recvCount++
		// Update state based on receive pattern
		s.updateStateOnRecv(dataSize)
	}
}

func (s *MaybenotState) updateStateOnSend(dataSize int) {
	// Rotate state based on data characteristics
	rotation := dataSize % 8
	for i := 0; i < rotation; i++ {
		s.rotateState()
	}

	// XOR with size-derived value
	sizeHash := byte(dataSize ^ (dataSize >> 8))
	for i := range s.state {
		s.state[i] ^= sizeHash
	}
}

func (s *MaybenotState) updateStateOnRecv(dataSize int) {
	// Similar to send but different transformation
	rotation := (dataSize + 3) % 8
	for i := 0; i < rotation; i++ {
		s.rotateState()
	}
}

func (s *MaybenotState) rotateState() {
	last := s.state[MaybenotStateSize-1]
	for i := MaybenotStateSize - 1; i > 0; i-- {
		s.state[i] = s.state[i-1]
	}
	s.state[0] = last
}

// GetAction returns the current recommended action
func (s *MaybenotState) GetAction() MaybenotAction {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Decision based on state hash
	stateHash := s.hashState()

	// Different actions based on state probability
	switch {
	case stateHash < 0x40:
		return MaybenotActionSendCover
	case stateHash < 0x60:
		return MaybenotActionSendBurst
	case stateHash < 0x80:
		return MaybenotActionDelay
	default:
		return MaybenotActionNone
	}
}

func (s *MaybenotState) hashState() byte {
	var hash byte
	for _, b := range s.state {
		hash ^= b
	}
	return hash
}

// CoverTrafficGenerator generates realistic cover traffic
type CoverTrafficGenerator struct {
	config     *DAITAConfig
	sizeHist   []int // Historical packet sizes for mimicry
	histIdx    int
	mu         sync.Mutex
}

// NewCoverTrafficGenerator creates a cover traffic generator
func NewCoverTrafficGenerator(config *DAITAConfig) *CoverTrafficGenerator {
	return &CoverTrafficGenerator{
		config:   config,
		sizeHist: make([]int, 100), // Track last 100 packet sizes
	}
}

// Generate creates a cover packet that mimics real traffic
func (g *CoverTrafficGenerator) Generate() []byte {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Use constant packet size if configured
	size := g.config.PacketSize
	if !g.config.ConstantPacketSize {
		// Mimic historical traffic pattern
		size = g.mimicSize()
	}

	packet := make([]byte, size)

	// Header
	binary.BigEndian.PutUint32(packet[:4], 0) // Zero length = cover packet
	packet[4] = 0xFF // Cover packet marker

	// Random padding indicator
	rand.Read(packet[5:8])

	// Random content
	rand.Read(packet[8:])

	return packet
}

func (g *CoverTrafficGenerator) mimicSize() int {
	// Choose a size from recent history
	if g.histIdx == 0 {
		return g.config.PacketSize
	}

	randBytes := make([]byte, 2)
	rand.Read(randBytes)
	idx := int(binary.BigEndian.Uint16(randBytes)) % g.histIdx

	return g.sizeHist[idx]
}

// RecordSize records a packet size for mimicry
func (g *CoverTrafficGenerator) RecordSize(size int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.sizeHist[g.histIdx%len(g.sizeHist)] = size
	g.histIdx++
}

// PatternDistorter distorts traffic patterns to defeat fingerprinting
type PatternDistorter struct {
	config *DAITAConfig

	// Timing distortion
	baseDelay   time.Duration
	jitterRange time.Duration

	// Burst detection
	burstThreshold int
	burstWindow    time.Duration
	lastBurst      time.Time
	burstCount     int

	mu sync.Mutex
}

// NewPatternDistorter creates a pattern distorter
func NewPatternDistorter(config *DAITAConfig) *PatternDistorter {
	return &PatternDistorter{
		config:         config,
		baseDelay:      5 * time.Millisecond,
		jitterRange:    20 * time.Millisecond,
		burstThreshold: 10,
		burstWindow:    100 * time.Millisecond,
	}
}

// GenerateDummies generates dummy packets to distort pattern
func (p *PatternDistorter) GenerateDummies(realPacketSize int) [][]byte {
	p.mu.Lock()
	defer p.mu.Unlock()

	var dummies [][]byte

	// Calculate number of dummies based on distortion level
	numDummies := int(p.config.DistortionLevel * 3) // 0-3 dummies

	// Random variation
	randBytes := make([]byte, 1)
	rand.Read(randBytes)
	if randBytes[0] > 128 {
		numDummies++
	}

	for i := 0; i < numDummies; i++ {
		// Create dummy with similar characteristics
		size := realPacketSize
		if !p.config.ConstantPacketSize {
			// Vary size slightly
			variation := int(randBytes[0]) % 100 - 50
			size += variation
			if size < DAITAMinPadding {
				size = DAITAMinPadding
			}
		} else {
			size = p.config.PacketSize
		}

		dummy := make([]byte, size)
		binary.BigEndian.PutUint32(dummy[:4], 0)
		dummy[4] = 0xFF // Cover marker
		rand.Read(dummy[5:])

		dummies = append(dummies, dummy)
	}

	return dummies
}

// GetJitter returns a random jitter duration
func (p *PatternDistorter) GetJitter() time.Duration {
	randBytes := make([]byte, 8)
	rand.Read(randBytes)
	randVal := binary.BigEndian.Uint64(randBytes)

	jitter := time.Duration(randVal % uint64(p.jitterRange))

	// Gaussian-like distribution (sum of uniform randoms)
	for i := 0; i < 3; i++ {
		rand.Read(randBytes)
		jitter += time.Duration(binary.BigEndian.Uint64(randBytes) % uint64(p.jitterRange/4))
	}
	jitter /= 4

	return p.baseDelay + jitter
}

// TokenBucket implements rate limiting for overhead control
type TokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a token bucket
func NewTokenBucket(maxTokens int, refillInterval time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:     float64(maxTokens),
		maxTokens:  float64(maxTokens),
		refillRate: float64(maxTokens) / refillInterval.Seconds(),
		lastRefill: time.Now(),
	}
}

// Take attempts to take tokens from the bucket
func (b *TokenBucket) Take(n int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	// Check if we have enough tokens
	if b.tokens >= float64(n) {
		b.tokens -= float64(n)
		return true
	}

	return false
}

// WebsiteFingerprinting provides additional defense against website fingerprinting
type WebsiteFingerprintDefense struct {
	// Page load detection
	pageLoadThreshold int
	pageLoadWindow    time.Duration

	// Traffic pattern database (for mimicry)
	patternDB map[string][]TrafficPattern

	// Defense state
	active    bool
	startTime time.Time

	mu sync.Mutex
}

// TrafficPattern represents a traffic pattern for mimicry
type TrafficPattern struct {
	Direction    bool // true = outbound
	Size         int
	InterArrival time.Duration
}

// NewWebsiteFingerprintDefense creates website fingerprinting defense
func NewWebsiteFingerprintDefense() *WebsiteFingerprintDefense {
	return &WebsiteFingerprintDefense{
		pageLoadThreshold: 50,                   // packets
		pageLoadWindow:    500 * time.Millisecond,
		patternDB:         make(map[string][]TrafficPattern),
	}
}

// StartDefense begins active defense for a page load
func (w *WebsiteFingerprintDefense) StartDefense() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.active = true
	w.startTime = time.Now()
}

// StopDefense ends active defense
func (w *WebsiteFingerprintDefense) StopDefense() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.active = false
}

// ShouldInjectCover returns whether cover traffic should be injected
func (w *WebsiteFingerprintDefense) ShouldInjectCover(packetCount int) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.active {
		return false
	}

	// During page loads, inject more cover traffic
	elapsed := time.Since(w.startTime)
	if elapsed < w.pageLoadWindow {
		// More aggressive cover during initial page load
		return packetCount%3 == 0
	}

	return false
}

// TrafficAnalysisResistance provides comprehensive traffic analysis resistance
type TrafficAnalysisResistance struct {
	daita    *DAITADefender
	wfDefense *WebsiteFingerprintDefense

	// Bandwidth analysis resistance
	bandwidthPadding bool
	targetBandwidth  int64 // bytes per second

	// Timing analysis resistance
	constantRate bool
	rateWindow   time.Duration

	mu sync.RWMutex
}

// NewTrafficAnalysisResistance creates comprehensive TAR
func NewTrafficAnalysisResistance(config *DAITAConfig) *TrafficAnalysisResistance {
	return &TrafficAnalysisResistance{
		daita:            NewDAITADefender(config),
		wfDefense:        NewWebsiteFingerprintDefense(),
		bandwidthPadding: true,
		targetBandwidth:  1024 * 1024, // 1 MB/s baseline
		constantRate:     false,
		rateWindow:       100 * time.Millisecond,
	}
}

// Process applies all traffic analysis resistance techniques
func (t *TrafficAnalysisResistance) Process(packet []byte, direction bool) ([][]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Apply DAITA protections
	result, err := t.daita.ProcessOutbound(packet)
	if err != nil {
		return nil, err
	}

	// Check if website fingerprinting defense should inject more
	if t.wfDefense.ShouldInjectCover(len(result)) {
		cover := t.daita.coverGenerator.Generate()
		result = append(result, cover)
	}

	return result, nil
}

// GetOverheadPercent returns current overhead percentage
func (t *TrafficAnalysisResistance) GetOverheadPercent() float64 {
	stats := t.daita.Stats()
	return math.Min(stats.OverheadPercent, 100.0)
}

// Close stops all traffic analysis resistance
func (t *TrafficAnalysisResistance) Close() {
	t.daita.Close()
}
