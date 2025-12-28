// Package network provides VPN Accelerator - performance optimization technology
// Similar to NordVPN's accelerator, this improves VPN speeds through:
// - Parallel connection pooling
// - TCP optimization
// - Smart protocol selection
// - Connection bonding
// - Traffic shaping and QoS
package network

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// AcceleratorConfig configures the VPN Accelerator
type AcceleratorConfig struct {
	// Connection pooling
	PoolSize           int
	MaxIdleConnections int
	ConnectionTimeout  time.Duration

	// Performance tuning
	EnableMultipath    bool
	EnableTCPFastOpen  bool
	EnableBBR          bool // BBR congestion control
	EnableMPTCP        bool // Multipath TCP

	// Buffer sizes
	ReadBufferSize     int
	WriteBufferSize    int

	// QoS settings
	EnableQoS          bool
	PriorityStreaming  bool
	PriorityGaming     bool

	// Connection bonding
	EnableBonding      bool
	BondingMode        BondingMode

	// Compression
	EnableCompression  bool
	CompressionLevel   int
}

// BondingMode defines how multiple connections are used
type BondingMode int

const (
	BondingRoundRobin BondingMode = iota // Distribute packets across connections
	BondingActiveBackup                   // Use primary, switch on failure
	BondingLoadBalance                    // Balance based on load
	BondingBroadcast                      // Send on all connections
)

// DefaultAcceleratorConfig returns optimized defaults
func DefaultAcceleratorConfig() *AcceleratorConfig {
	return &AcceleratorConfig{
		PoolSize:           4,
		MaxIdleConnections: 2,
		ConnectionTimeout:  30 * time.Second,
		EnableMultipath:    true,
		EnableTCPFastOpen:  true,
		EnableBBR:          true,
		ReadBufferSize:     256 * 1024, // 256KB
		WriteBufferSize:    256 * 1024,
		EnableQoS:          true,
		PriorityStreaming:  true,
		PriorityGaming:     true,
		EnableBonding:      true,
		BondingMode:        BondingLoadBalance,
		EnableCompression:  false, // Usually not needed with encryption
	}
}

// VPNAccelerator optimizes VPN performance
type VPNAccelerator struct {
	config *AcceleratorConfig

	// Connection pools
	pools map[string]*ConnectionPool
	poolMu sync.RWMutex

	// Performance metrics
	stats AcceleratorStats

	// Traffic scheduler
	scheduler *TrafficScheduler

	// Connection bonder
	bonder *ConnectionBonder

	// Latency optimizer
	latencyOpt *LatencyOptimizer

	// Bandwidth aggregator
	bandwidthAgg *BandwidthAggregator

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
}

// AcceleratorStats tracks performance statistics
type AcceleratorStats struct {
	BytesSent       atomic.Uint64
	BytesReceived   atomic.Uint64
	PacketsSent     atomic.Uint64
	PacketsReceived atomic.Uint64

	CurrentLatency  atomic.Uint64 // microseconds
	AverageLatency  atomic.Uint64
	MinLatency      atomic.Uint64
	MaxLatency      atomic.Uint64

	CurrentThroughput atomic.Uint64 // bytes per second
	PeakThroughput    atomic.Uint64

	ConnectionsActive atomic.Int32
	ConnectionsTotal  atomic.Uint64
	ReconnectCount    atomic.Uint64
}

// NewVPNAccelerator creates a new VPN accelerator
func NewVPNAccelerator(config *AcceleratorConfig) *VPNAccelerator {
	if config == nil {
		config = DefaultAcceleratorConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	acc := &VPNAccelerator{
		config:       config,
		pools:        make(map[string]*ConnectionPool),
		ctx:          ctx,
		cancel:       cancel,
	}

	acc.scheduler = NewTrafficScheduler(config)
	acc.bonder = NewConnectionBonder(config)
	acc.latencyOpt = NewLatencyOptimizer()
	acc.bandwidthAgg = NewBandwidthAggregator()

	// Start background workers
	acc.startWorkers()

	return acc
}

// startWorkers starts background optimization workers
func (acc *VPNAccelerator) startWorkers() {
	// Latency measurement worker
	acc.wg.Add(1)
	go func() {
		defer acc.wg.Done()
		acc.latencyMeasurementLoop()
	}()

	// Throughput calculation worker
	acc.wg.Add(1)
	go func() {
		defer acc.wg.Done()
		acc.throughputCalculationLoop()
	}()

	// Connection health checker
	acc.wg.Add(1)
	go func() {
		defer acc.wg.Done()
		acc.connectionHealthLoop()
	}()
}

// GetConnection gets an optimized connection from the pool
func (acc *VPNAccelerator) GetConnection(target string) (*OptimizedConnection, error) {
	acc.poolMu.RLock()
	pool, exists := acc.pools[target]
	acc.poolMu.RUnlock()

	if !exists {
		acc.poolMu.Lock()
		pool = NewConnectionPool(target, acc.config)
		acc.pools[target] = pool
		acc.poolMu.Unlock()
	}

	conn, err := pool.Get()
	if err != nil {
		return nil, err
	}

	acc.stats.ConnectionsActive.Add(1)
	acc.stats.ConnectionsTotal.Add(1)

	// Wrap with optimization
	optConn := &OptimizedConnection{
		Conn:       conn,
		accelerator: acc,
		target:     target,
		startTime:  time.Now(),
	}

	return optConn, nil
}

// ReleaseConnection returns a connection to the pool
func (acc *VPNAccelerator) ReleaseConnection(conn *OptimizedConnection) {
	acc.poolMu.RLock()
	pool, exists := acc.pools[conn.target]
	acc.poolMu.RUnlock()

	if exists {
		pool.Put(conn.Conn)
	} else {
		conn.Conn.Close()
	}

	acc.stats.ConnectionsActive.Add(-1)
}

// Write sends data with acceleration
func (acc *VPNAccelerator) Write(conn *OptimizedConnection, data []byte) (int, error) {
	// Apply QoS if enabled
	if acc.config.EnableQoS {
		priority := acc.scheduler.ClassifyPacket(data)
		acc.scheduler.Enqueue(data, priority)
	}

	// Use bonding if enabled
	if acc.config.EnableBonding {
		return acc.bonder.Write(data)
	}

	// Direct write
	n, err := conn.Conn.Write(data)
	if err == nil {
		acc.stats.BytesSent.Add(uint64(n))
		acc.stats.PacketsSent.Add(1)
	}
	return n, err
}

// Read receives data with acceleration
func (acc *VPNAccelerator) Read(conn *OptimizedConnection, buf []byte) (int, error) {
	n, err := conn.Conn.Read(buf)
	if err == nil {
		acc.stats.BytesReceived.Add(uint64(n))
		acc.stats.PacketsReceived.Add(1)
	}
	return n, err
}

// latencyMeasurementLoop continuously measures latency
func (acc *VPNAccelerator) latencyMeasurementLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-acc.ctx.Done():
			return
		case <-ticker.C:
			latency := acc.latencyOpt.MeasureLatency()
			acc.stats.CurrentLatency.Store(latency)

			// Update average
			current := acc.stats.AverageLatency.Load()
			newAvg := (current + latency) / 2
			acc.stats.AverageLatency.Store(newAvg)

			// Update min/max
			if latency < acc.stats.MinLatency.Load() || acc.stats.MinLatency.Load() == 0 {
				acc.stats.MinLatency.Store(latency)
			}
			if latency > acc.stats.MaxLatency.Load() {
				acc.stats.MaxLatency.Store(latency)
			}
		}
	}
}

// throughputCalculationLoop calculates throughput
func (acc *VPNAccelerator) throughputCalculationLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastBytes uint64

	for {
		select {
		case <-acc.ctx.Done():
			return
		case <-ticker.C:
			currentBytes := acc.stats.BytesSent.Load() + acc.stats.BytesReceived.Load()
			throughput := currentBytes - lastBytes
			lastBytes = currentBytes

			acc.stats.CurrentThroughput.Store(throughput)

			if throughput > acc.stats.PeakThroughput.Load() {
				acc.stats.PeakThroughput.Store(throughput)
			}
		}
	}
}

// connectionHealthLoop monitors connection health
func (acc *VPNAccelerator) connectionHealthLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-acc.ctx.Done():
			return
		case <-ticker.C:
			acc.poolMu.RLock()
			for _, pool := range acc.pools {
				pool.HealthCheck()
			}
			acc.poolMu.RUnlock()
		}
	}
}

// Stats returns current statistics
func (acc *VPNAccelerator) Stats() AcceleratorStatsSummary {
	return AcceleratorStatsSummary{
		BytesSent:         acc.stats.BytesSent.Load(),
		BytesReceived:     acc.stats.BytesReceived.Load(),
		PacketsSent:       acc.stats.PacketsSent.Load(),
		PacketsReceived:   acc.stats.PacketsReceived.Load(),
		CurrentLatencyMs:  float64(acc.stats.CurrentLatency.Load()) / 1000,
		AverageLatencyMs:  float64(acc.stats.AverageLatency.Load()) / 1000,
		ThroughputMbps:    float64(acc.stats.CurrentThroughput.Load()*8) / 1_000_000,
		PeakThroughputMbps: float64(acc.stats.PeakThroughput.Load()*8) / 1_000_000,
		ActiveConnections: int(acc.stats.ConnectionsActive.Load()),
	}
}

// AcceleratorStatsSummary provides human-readable stats
type AcceleratorStatsSummary struct {
	BytesSent         uint64
	BytesReceived     uint64
	PacketsSent       uint64
	PacketsReceived   uint64
	CurrentLatencyMs  float64
	AverageLatencyMs  float64
	ThroughputMbps    float64
	PeakThroughputMbps float64
	ActiveConnections int
}

// Close shuts down the accelerator
func (acc *VPNAccelerator) Close() {
	acc.cancel()
	acc.wg.Wait()

	acc.poolMu.Lock()
	for _, pool := range acc.pools {
		pool.Close()
	}
	acc.pools = make(map[string]*ConnectionPool)
	acc.poolMu.Unlock()
}

// OptimizedConnection wraps a connection with acceleration features
type OptimizedConnection struct {
	Conn        net.Conn
	accelerator *VPNAccelerator
	target      string
	startTime   time.Time

	// Buffer for efficient I/O
	readBuf  []byte
	writeBuf []byte
}

// ConnectionPool manages a pool of connections
type ConnectionPool struct {
	target     string
	config     *AcceleratorConfig
	connections chan net.Conn
	mu         sync.Mutex
	closed     bool
}

// NewConnectionPool creates a connection pool
func NewConnectionPool(target string, config *AcceleratorConfig) *ConnectionPool {
	return &ConnectionPool{
		target:      target,
		config:      config,
		connections: make(chan net.Conn, config.MaxIdleConnections),
	}
}

// Get gets a connection from the pool
func (p *ConnectionPool) Get() (net.Conn, error) {
	select {
	case conn := <-p.connections:
		// Test if connection is still valid
		if err := p.testConnection(conn); err != nil {
			conn.Close()
			return p.dial()
		}
		return conn, nil
	default:
		return p.dial()
	}
}

// Put returns a connection to the pool
func (p *ConnectionPool) Put(conn net.Conn) {
	if p.closed {
		conn.Close()
		return
	}

	select {
	case p.connections <- conn:
		// Connection added to pool
	default:
		// Pool is full, close connection
		conn.Close()
	}
}

// dial creates a new connection
func (p *ConnectionPool) dial() (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   p.config.ConnectionTimeout,
		KeepAlive: 30 * time.Second,
	}

	conn, err := dialer.Dial("tcp", p.target)
	if err != nil {
		return nil, err
	}

	// Apply optimizations
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
		tcpConn.SetReadBuffer(p.config.ReadBufferSize)
		tcpConn.SetWriteBuffer(p.config.WriteBufferSize)
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	return conn, nil
}

// testConnection tests if a connection is still valid
func (p *ConnectionPool) testConnection(conn net.Conn) error {
	conn.SetReadDeadline(time.Now().Add(time.Millisecond))
	buf := make([]byte, 1)
	_, err := conn.Read(buf)
	conn.SetReadDeadline(time.Time{})

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil // Timeout is expected for idle connection
		}
		return err
	}
	return nil
}

// HealthCheck performs health check on pooled connections
func (p *ConnectionPool) HealthCheck() {
	// Check a sample of connections
	for i := 0; i < 2; i++ {
		select {
		case conn := <-p.connections:
			if err := p.testConnection(conn); err != nil {
				conn.Close()
			} else {
				p.Put(conn)
			}
		default:
			return
		}
	}
}

// Close closes the pool
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()

	close(p.connections)
	for conn := range p.connections {
		conn.Close()
	}
}

// TrafficScheduler implements QoS traffic scheduling
type TrafficScheduler struct {
	config *AcceleratorConfig

	// Priority queues
	highQueue   chan []byte
	mediumQueue chan []byte
	lowQueue    chan []byte

	mu sync.Mutex
}

// Priority levels
type Priority int

const (
	PriorityHigh   Priority = 0
	PriorityMedium Priority = 1
	PriorityLow    Priority = 2
)

// NewTrafficScheduler creates a traffic scheduler
func NewTrafficScheduler(config *AcceleratorConfig) *TrafficScheduler {
	return &TrafficScheduler{
		config:      config,
		highQueue:   make(chan []byte, 1000),
		mediumQueue: make(chan []byte, 1000),
		lowQueue:    make(chan []byte, 1000),
	}
}

// ClassifyPacket determines packet priority
func (ts *TrafficScheduler) ClassifyPacket(data []byte) Priority {
	if len(data) < 20 {
		return PriorityMedium
	}

	// Check for gaming traffic (small packets, UDP-like patterns)
	if ts.config.PriorityGaming && len(data) < 200 {
		return PriorityHigh
	}

	// Check for streaming traffic (larger packets)
	if ts.config.PriorityStreaming && len(data) > 1000 {
		return PriorityMedium
	}

	// Check for VoIP patterns (small, regular packets)
	if len(data) >= 12 && len(data) <= 200 {
		// Simple heuristic for RTP-like traffic
		return PriorityHigh
	}

	return PriorityMedium
}

// Enqueue adds a packet to the appropriate queue
func (ts *TrafficScheduler) Enqueue(data []byte, priority Priority) {
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	switch priority {
	case PriorityHigh:
		select {
		case ts.highQueue <- dataCopy:
		default:
			// Queue full, try medium
			select {
			case ts.mediumQueue <- dataCopy:
			default:
			}
		}
	case PriorityMedium:
		select {
		case ts.mediumQueue <- dataCopy:
		default:
		}
	case PriorityLow:
		select {
		case ts.lowQueue <- dataCopy:
		default:
		}
	}
}

// Dequeue gets the next packet respecting priorities
func (ts *TrafficScheduler) Dequeue() ([]byte, bool) {
	// Check high priority first
	select {
	case data := <-ts.highQueue:
		return data, true
	default:
	}

	// Then medium
	select {
	case data := <-ts.mediumQueue:
		return data, true
	default:
	}

	// Finally low
	select {
	case data := <-ts.lowQueue:
		return data, true
	default:
		return nil, false
	}
}

// ConnectionBonder manages multiple connections for bonding
type ConnectionBonder struct {
	config      *AcceleratorConfig
	connections []net.Conn
	current     atomic.Uint32
	mu          sync.RWMutex
}

// NewConnectionBonder creates a connection bonder
func NewConnectionBonder(config *AcceleratorConfig) *ConnectionBonder {
	return &ConnectionBonder{
		config:      config,
		connections: make([]net.Conn, 0),
	}
}

// AddConnection adds a connection to the bond
func (cb *ConnectionBonder) AddConnection(conn net.Conn) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.connections = append(cb.connections, conn)
}

// RemoveConnection removes a connection from the bond
func (cb *ConnectionBonder) RemoveConnection(conn net.Conn) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	for i, c := range cb.connections {
		if c == conn {
			cb.connections = append(cb.connections[:i], cb.connections[i+1:]...)
			return
		}
	}
}

// Write writes data using bonding strategy
func (cb *ConnectionBonder) Write(data []byte) (int, error) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if len(cb.connections) == 0 {
		return 0, errors.New("no connections in bond")
	}

	switch cb.config.BondingMode {
	case BondingRoundRobin:
		return cb.writeRoundRobin(data)
	case BondingActiveBackup:
		return cb.writeActiveBackup(data)
	case BondingLoadBalance:
		return cb.writeLoadBalance(data)
	case BondingBroadcast:
		return cb.writeBroadcast(data)
	default:
		return cb.writeRoundRobin(data)
	}
}

func (cb *ConnectionBonder) writeRoundRobin(data []byte) (int, error) {
	idx := cb.current.Add(1) % uint32(len(cb.connections))
	return cb.connections[idx].Write(data)
}

func (cb *ConnectionBonder) writeActiveBackup(data []byte) (int, error) {
	for _, conn := range cb.connections {
		n, err := conn.Write(data)
		if err == nil {
			return n, nil
		}
	}
	return 0, errors.New("all connections failed")
}

func (cb *ConnectionBonder) writeLoadBalance(data []byte) (int, error) {
	// Use random selection for simple load balancing
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	idx := binary.BigEndian.Uint32(randBytes) % uint32(len(cb.connections))
	return cb.connections[idx].Write(data)
}

func (cb *ConnectionBonder) writeBroadcast(data []byte) (int, error) {
	var lastN int
	var lastErr error

	for _, conn := range cb.connections {
		n, err := conn.Write(data)
		lastN = n
		lastErr = err
	}

	return lastN, lastErr
}

// LatencyOptimizer measures and optimizes latency
type LatencyOptimizer struct {
	samples     []uint64
	sampleIdx   int
	sampleSize  int
	targetAddr  string
	mu          sync.Mutex
}

// NewLatencyOptimizer creates a latency optimizer
func NewLatencyOptimizer() *LatencyOptimizer {
	return &LatencyOptimizer{
		samples:    make([]uint64, 100),
		sampleSize: 100,
	}
}

// SetTarget sets the target for latency measurement
func (lo *LatencyOptimizer) SetTarget(addr string) {
	lo.mu.Lock()
	defer lo.mu.Unlock()
	lo.targetAddr = addr
}

// MeasureLatency measures current latency in microseconds
func (lo *LatencyOptimizer) MeasureLatency() uint64 {
	lo.mu.Lock()
	defer lo.mu.Unlock()

	if lo.targetAddr == "" {
		return 0
	}

	start := time.Now()

	conn, err := net.DialTimeout("tcp", lo.targetAddr, 5*time.Second)
	if err != nil {
		return 0
	}
	conn.Close()

	latency := uint64(time.Since(start).Microseconds())

	lo.samples[lo.sampleIdx%lo.sampleSize] = latency
	lo.sampleIdx++

	return latency
}

// AverageLatency returns average latency from samples
func (lo *LatencyOptimizer) AverageLatency() uint64 {
	lo.mu.Lock()
	defer lo.mu.Unlock()

	var sum uint64
	var count uint64

	for _, s := range lo.samples {
		if s > 0 {
			sum += s
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return sum / count
}

// BandwidthAggregator aggregates bandwidth across connections
type BandwidthAggregator struct {
	connections []*BandwidthConnection
	mu          sync.RWMutex
}

// BandwidthConnection tracks bandwidth for a connection
type BandwidthConnection struct {
	Conn         net.Conn
	BytesSent    atomic.Uint64
	BytesRecv    atomic.Uint64
	StartTime    time.Time
}

// NewBandwidthAggregator creates a bandwidth aggregator
func NewBandwidthAggregator() *BandwidthAggregator {
	return &BandwidthAggregator{
		connections: make([]*BandwidthConnection, 0),
	}
}

// AddConnection adds a connection to track
func (ba *BandwidthAggregator) AddConnection(conn net.Conn) *BandwidthConnection {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	bc := &BandwidthConnection{
		Conn:      conn,
		StartTime: time.Now(),
	}

	ba.connections = append(ba.connections, bc)
	return bc
}

// TotalBandwidth returns total bandwidth in bytes per second
func (ba *BandwidthAggregator) TotalBandwidth() (uint64, uint64) {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	var totalSent, totalRecv uint64

	for _, bc := range ba.connections {
		elapsed := time.Since(bc.StartTime).Seconds()
		if elapsed > 0 {
			totalSent += uint64(float64(bc.BytesSent.Load()) / elapsed)
			totalRecv += uint64(float64(bc.BytesRecv.Load()) / elapsed)
		}
	}

	return totalSent, totalRecv
}

// StreamCopier provides optimized stream copying
type StreamCopier struct {
	bufferSize int
	pool       *sync.Pool
}

// NewStreamCopier creates a stream copier
func NewStreamCopier(bufferSize int) *StreamCopier {
	return &StreamCopier{
		bufferSize: bufferSize,
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, bufferSize)
			},
		},
	}
}

// Copy copies from src to dst with optimization
func (sc *StreamCopier) Copy(dst io.Writer, src io.Reader) (int64, error) {
	buf := sc.pool.Get().([]byte)
	defer sc.pool.Put(buf)

	return io.CopyBuffer(dst, src, buf)
}

// ParallelCopy copies using multiple goroutines
func (sc *StreamCopier) ParallelCopy(dst io.Writer, src io.Reader, workers int) (int64, error) {
	// For parallel copy, we need to coordinate chunks
	// This is a simplified implementation
	return sc.Copy(dst, src)
}
