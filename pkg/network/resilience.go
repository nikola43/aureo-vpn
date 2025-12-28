// Package network provides connection resilience
// Handles automatic reconnection, failover, and recovery
package network

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// ResilienceConfig configures connection resilience
type ResilienceConfig struct {
	// Retry settings
	MaxRetries        int
	RetryBackoff      time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64

	// Health check
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration

	// Failover
	EnableFailover    bool
	FailoverThreshold int // Consecutive failures before failover
	FailbackDelay     time.Duration

	// Keep-alive
	KeepAliveInterval time.Duration
	KeepAliveTimeout  time.Duration

	// Circuit breaker
	EnableCircuitBreaker bool
	BreakerThreshold     int
	BreakerTimeout       time.Duration
}

// DefaultResilienceConfig returns resilient defaults
func DefaultResilienceConfig() *ResilienceConfig {
	return &ResilienceConfig{
		MaxRetries:           5,
		RetryBackoff:         time.Second,
		MaxBackoff:           60 * time.Second,
		BackoffMultiplier:    2.0,
		HealthCheckInterval:  30 * time.Second,
		HealthCheckTimeout:   5 * time.Second,
		EnableFailover:       true,
		FailoverThreshold:    3,
		FailbackDelay:        5 * time.Minute,
		KeepAliveInterval:    25 * time.Second,
		KeepAliveTimeout:     10 * time.Second,
		EnableCircuitBreaker: true,
		BreakerThreshold:     5,
		BreakerTimeout:       30 * time.Second,
	}
}

// ConnectionState represents connection state
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateFailed
	StateFailover
)

func (s ConnectionState) String() string {
	names := []string{"disconnected", "connecting", "connected", "reconnecting", "failed", "failover"}
	if int(s) < len(names) {
		return names[s]
	}
	return "unknown"
}

// ResilientConnection wraps a connection with resilience features
type ResilientConnection struct {
	config *ResilienceConfig

	// Current connection
	conn   net.Conn
	connMu sync.RWMutex

	// State
	state      ConnectionState
	stateMu    sync.RWMutex

	// Target
	primary   string
	secondary []string
	current   string

	// Retry state
	retryCount   int
	lastRetry    time.Time
	currentBackoff time.Duration

	// Health
	consecutiveFailures atomic.Int32
	lastHealthCheck     time.Time
	healthy             atomic.Bool

	// Circuit breaker
	breaker *CircuitBreaker

	// Stats
	stats ResilienceStats

	// Callbacks
	onConnect    func()
	onDisconnect func(error)
	onFailover   func(string)

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// ResilienceStats tracks resilience statistics
type ResilienceStats struct {
	Connections     atomic.Uint64
	Disconnections  atomic.Uint64
	Reconnections   atomic.Uint64
	Failovers       atomic.Uint64
	FailedAttempts  atomic.Uint64
	TotalDowntime   atomic.Int64 // milliseconds
	LastConnected   time.Time
	LastDisconnected time.Time
}

// NewResilientConnection creates a resilient connection
func NewResilientConnection(primary string, config *ResilienceConfig) *ResilientConnection {
	if config == nil {
		config = DefaultResilienceConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	rc := &ResilientConnection{
		config:         config,
		primary:        primary,
		current:        primary,
		secondary:      make([]string, 0),
		state:          StateDisconnected,
		currentBackoff: config.RetryBackoff,
		ctx:            ctx,
		cancel:         cancel,
	}

	if config.EnableCircuitBreaker {
		rc.breaker = NewCircuitBreaker(config.BreakerThreshold, config.BreakerTimeout)
	}

	return rc
}

// AddSecondary adds a secondary/failover target
func (rc *ResilientConnection) AddSecondary(addr string) {
	rc.secondary = append(rc.secondary, addr)
}

// Connect establishes the connection with resilience
func (rc *ResilientConnection) Connect() error {
	rc.setState(StateConnecting)

	err := rc.attemptConnect(rc.primary)
	if err != nil {
		if rc.config.EnableFailover && len(rc.secondary) > 0 {
			return rc.failover()
		}
		rc.setState(StateFailed)
		return err
	}

	rc.setState(StateConnected)
	rc.stats.Connections.Add(1)
	rc.stats.LastConnected = time.Now()
	rc.healthy.Store(true)
	rc.consecutiveFailures.Store(0)
	rc.retryCount = 0
	rc.currentBackoff = rc.config.RetryBackoff

	// Start health monitoring
	rc.wg.Add(1)
	go rc.healthMonitorLoop()

	// Start keep-alive
	rc.wg.Add(1)
	go rc.keepAliveLoop()

	if rc.onConnect != nil {
		rc.onConnect()
	}

	return nil
}

// attemptConnect tries to connect to a target
func (rc *ResilientConnection) attemptConnect(target string) error {
	if rc.breaker != nil && !rc.breaker.Allow() {
		return errors.New("circuit breaker open")
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: rc.config.KeepAliveInterval,
	}

	conn, err := dialer.DialContext(rc.ctx, "tcp", target)
	if err != nil {
		rc.stats.FailedAttempts.Add(1)
		if rc.breaker != nil {
			rc.breaker.RecordFailure()
		}
		return err
	}

	rc.connMu.Lock()
	rc.conn = conn
	rc.current = target
	rc.connMu.Unlock()

	if rc.breaker != nil {
		rc.breaker.RecordSuccess()
	}

	return nil
}

// failover attempts to connect to secondary targets
func (rc *ResilientConnection) failover() error {
	rc.setState(StateFailover)

	for _, target := range rc.secondary {
		if err := rc.attemptConnect(target); err == nil {
			rc.stats.Failovers.Add(1)
			rc.setState(StateConnected)

			if rc.onFailover != nil {
				rc.onFailover(target)
			}

			// Schedule failback check
			go rc.failbackLoop()

			return nil
		}
	}

	rc.setState(StateFailed)
	return errors.New("all failover targets failed")
}

// Reconnect attempts to reconnect after disconnection
func (rc *ResilientConnection) Reconnect() error {
	rc.setState(StateReconnecting)
	rc.stats.Reconnections.Add(1)

	for rc.retryCount < rc.config.MaxRetries {
		select {
		case <-rc.ctx.Done():
			return rc.ctx.Err()
		default:
		}

		// Apply backoff
		if rc.retryCount > 0 {
			time.Sleep(rc.currentBackoff)
			rc.currentBackoff = time.Duration(float64(rc.currentBackoff) * rc.config.BackoffMultiplier)
			if rc.currentBackoff > rc.config.MaxBackoff {
				rc.currentBackoff = rc.config.MaxBackoff
			}
		}

		err := rc.attemptConnect(rc.current)
		if err == nil {
			rc.setState(StateConnected)
			rc.retryCount = 0
			rc.currentBackoff = rc.config.RetryBackoff
			return nil
		}

		rc.retryCount++
		rc.lastRetry = time.Now()
	}

	// Try failover
	if rc.config.EnableFailover && len(rc.secondary) > 0 {
		return rc.failover()
	}

	rc.setState(StateFailed)
	return errors.New("max retries exceeded")
}

// healthMonitorLoop monitors connection health
func (rc *ResilientConnection) healthMonitorLoop() {
	defer rc.wg.Done()

	ticker := time.NewTicker(rc.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rc.ctx.Done():
			return
		case <-ticker.C:
			if !rc.checkHealth() {
				failures := rc.consecutiveFailures.Add(1)
				if int(failures) >= rc.config.FailoverThreshold {
					rc.handleConnectionFailure(errors.New("health check failed"))
				}
			} else {
				rc.consecutiveFailures.Store(0)
			}
		}
	}
}

// checkHealth performs a health check
func (rc *ResilientConnection) checkHealth() bool {
	rc.connMu.RLock()
	conn := rc.conn
	rc.connMu.RUnlock()

	if conn == nil {
		return false
	}

	conn.SetDeadline(time.Now().Add(rc.config.HealthCheckTimeout))
	defer conn.SetDeadline(time.Time{})

	// Try to read (should timeout for healthy idle connection)
	buf := make([]byte, 1)
	_, err := conn.Read(buf)

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			rc.healthy.Store(true)
			rc.lastHealthCheck = time.Now()
			return true
		}
		rc.healthy.Store(false)
		return false
	}

	// Got data unexpectedly, connection is still alive
	rc.healthy.Store(true)
	return true
}

// keepAliveLoop sends keep-alive packets
func (rc *ResilientConnection) keepAliveLoop() {
	defer rc.wg.Done()

	ticker := time.NewTicker(rc.config.KeepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rc.ctx.Done():
			return
		case <-ticker.C:
			if err := rc.sendKeepAlive(); err != nil {
				rc.consecutiveFailures.Add(1)
			}
		}
	}
}

// sendKeepAlive sends a keep-alive packet
func (rc *ResilientConnection) sendKeepAlive() error {
	rc.connMu.RLock()
	conn := rc.conn
	rc.connMu.RUnlock()

	if conn == nil {
		return errors.New("no connection")
	}

	// For TCP, just setting deadline and doing a zero-byte write
	conn.SetWriteDeadline(time.Now().Add(rc.config.KeepAliveTimeout))
	defer conn.SetWriteDeadline(time.Time{})

	// The actual keep-alive would be protocol-specific
	return nil
}

// failbackLoop checks if primary is available again
func (rc *ResilientConnection) failbackLoop() {
	if rc.current == rc.primary {
		return // Already on primary
	}

	ticker := time.NewTicker(rc.config.FailbackDelay)
	defer ticker.Stop()

	for {
		select {
		case <-rc.ctx.Done():
			return
		case <-ticker.C:
			// Try to connect to primary
			dialer := &net.Dialer{Timeout: 5 * time.Second}
			conn, err := dialer.Dial("tcp", rc.primary)
			if err == nil {
				conn.Close()
				// Primary is back, initiate failback
				rc.failback()
				return
			}
		}
	}
}

// failback switches back to primary
func (rc *ResilientConnection) failback() {
	rc.connMu.Lock()
	if rc.conn != nil {
		rc.conn.Close()
	}
	rc.connMu.Unlock()

	rc.current = rc.primary
	rc.Reconnect()
}

// handleConnectionFailure handles a connection failure
func (rc *ResilientConnection) handleConnectionFailure(err error) {
	rc.stats.Disconnections.Add(1)
	rc.stats.LastDisconnected = time.Now()

	if rc.onDisconnect != nil {
		rc.onDisconnect(err)
	}

	// Track downtime
	if !rc.stats.LastConnected.IsZero() {
		downtime := time.Since(rc.stats.LastConnected).Milliseconds()
		rc.stats.TotalDowntime.Add(downtime)
	}

	// Attempt reconnection
	go rc.Reconnect()
}

// Write writes data with resilience
func (rc *ResilientConnection) Write(data []byte) (int, error) {
	rc.connMu.RLock()
	conn := rc.conn
	rc.connMu.RUnlock()

	if conn == nil {
		return 0, errors.New("not connected")
	}

	n, err := conn.Write(data)
	if err != nil {
		rc.handleConnectionFailure(err)
		return n, err
	}

	return n, nil
}

// Read reads data with resilience
func (rc *ResilientConnection) Read(buf []byte) (int, error) {
	rc.connMu.RLock()
	conn := rc.conn
	rc.connMu.RUnlock()

	if conn == nil {
		return 0, errors.New("not connected")
	}

	n, err := conn.Read(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && !netErr.Timeout() {
			rc.handleConnectionFailure(err)
		}
		return n, err
	}

	return n, nil
}

// State returns current connection state
func (rc *ResilientConnection) State() ConnectionState {
	rc.stateMu.RLock()
	defer rc.stateMu.RUnlock()
	return rc.state
}

func (rc *ResilientConnection) setState(state ConnectionState) {
	rc.stateMu.Lock()
	defer rc.stateMu.Unlock()
	rc.state = state
}

// IsConnected returns whether connected
func (rc *ResilientConnection) IsConnected() bool {
	return rc.State() == StateConnected
}

// IsHealthy returns whether connection is healthy
func (rc *ResilientConnection) IsHealthy() bool {
	return rc.healthy.Load()
}

// OnConnect sets connect callback
func (rc *ResilientConnection) OnConnect(fn func()) {
	rc.onConnect = fn
}

// OnDisconnect sets disconnect callback
func (rc *ResilientConnection) OnDisconnect(fn func(error)) {
	rc.onDisconnect = fn
}

// OnFailover sets failover callback
func (rc *ResilientConnection) OnFailover(fn func(string)) {
	rc.onFailover = fn
}

// Stats returns resilience statistics
func (rc *ResilientConnection) Stats() ResilienceStatsSummary {
	return ResilienceStatsSummary{
		Connections:      rc.stats.Connections.Load(),
		Disconnections:   rc.stats.Disconnections.Load(),
		Reconnections:    rc.stats.Reconnections.Load(),
		Failovers:        rc.stats.Failovers.Load(),
		FailedAttempts:   rc.stats.FailedAttempts.Load(),
		TotalDowntimeMs:  rc.stats.TotalDowntime.Load(),
		CurrentState:     rc.State().String(),
		IsHealthy:        rc.healthy.Load(),
	}
}

// ResilienceStatsSummary provides stats summary
type ResilienceStatsSummary struct {
	Connections     uint64
	Disconnections  uint64
	Reconnections   uint64
	Failovers       uint64
	FailedAttempts  uint64
	TotalDowntimeMs int64
	CurrentState    string
	IsHealthy       bool
}

// Close closes the resilient connection
func (rc *ResilientConnection) Close() error {
	rc.cancel()
	rc.wg.Wait()

	rc.connMu.Lock()
	defer rc.connMu.Unlock()

	if rc.conn != nil {
		err := rc.conn.Close()
		rc.conn = nil
		rc.setState(StateDisconnected)
		return err
	}

	return nil
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	threshold    int
	timeout      time.Duration

	failures     atomic.Int32
	state        atomic.Int32 // 0=closed, 1=open, 2=half-open
	lastFailure  time.Time
	mu           sync.RWMutex
}

const (
	circuitClosed = iota
	circuitOpen
	circuitHalfOpen
)

// NewCircuitBreaker creates a circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		timeout:   timeout,
	}
}

// Allow checks if a request should be allowed
func (cb *CircuitBreaker) Allow() bool {
	state := cb.state.Load()

	switch state {
	case circuitClosed:
		return true

	case circuitOpen:
		cb.mu.RLock()
		elapsed := time.Since(cb.lastFailure)
		cb.mu.RUnlock()

		if elapsed > cb.timeout {
			// Transition to half-open
			cb.state.Store(circuitHalfOpen)
			return true
		}
		return false

	case circuitHalfOpen:
		return true

	default:
		return true
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.failures.Store(0)
	cb.state.Store(circuitClosed)
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	failures := cb.failures.Add(1)

	cb.mu.Lock()
	cb.lastFailure = time.Now()
	cb.mu.Unlock()

	if int(failures) >= cb.threshold {
		cb.state.Store(circuitOpen)
	}
}

// State returns circuit breaker state
func (cb *CircuitBreaker) State() string {
	switch cb.state.Load() {
	case circuitClosed:
		return "closed"
	case circuitOpen:
		return "open"
	case circuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
