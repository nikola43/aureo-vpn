// Package security provides secure rate limiting with protection against abuse
package security

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Rate limiting errors
var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrTooManyRequests   = errors.New("too many requests")
	ErrClientBlocked     = errors.New("client temporarily blocked")
)

// RateLimitConfig defines rate limiting parameters
type RateLimitConfig struct {
	// Basic rate limiting
	MaxRequests     int           // Maximum requests per window
	WindowSize      time.Duration // Time window for rate limiting

	// Burst handling
	BurstSize       int           // Maximum burst size
	BurstRecovery   time.Duration // Time to recover one burst token

	// Progressive penalties
	PenaltyEnabled  bool          // Enable progressive penalties
	PenaltyFactor   float64       // Multiply window by this on each violation
	MaxPenalty      time.Duration // Maximum penalty duration

	// Blocking
	BlockThreshold  int           // Number of violations before blocking
	BlockDuration   time.Duration // Duration to block after threshold
}

// DefaultRateLimitConfig returns secure defaults
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		MaxRequests:    100,
		WindowSize:     time.Minute,
		BurstSize:      10,
		BurstRecovery:  100 * time.Millisecond,
		PenaltyEnabled: true,
		PenaltyFactor:  2.0,
		MaxPenalty:     time.Hour,
		BlockThreshold: 10,
		BlockDuration:  15 * time.Minute,
	}
}

// AuthRateLimitConfig returns stricter config for auth endpoints
func AuthRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		MaxRequests:    5,
		WindowSize:     time.Minute,
		BurstSize:      3,
		BurstRecovery:  5 * time.Second,
		PenaltyEnabled: true,
		PenaltyFactor:  3.0,
		MaxPenalty:     24 * time.Hour,
		BlockThreshold: 5,
		BlockDuration:  time.Hour,
	}
}

// SecureRateLimiter implements thread-safe rate limiting with advanced features
type SecureRateLimiter struct {
	config  *RateLimitConfig
	clients map[string]*ClientState
	blocked map[string]time.Time
	mu      sync.RWMutex

	// Cleanup
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// ClientState tracks the state of a rate-limited client
type ClientState struct {
	Requests    []time.Time   // Request timestamps in current window
	BurstTokens int           // Available burst tokens
	LastBurst   time.Time     // Last burst token recovery time
	Violations  int           // Number of rate limit violations
	PenaltyEnd  time.Time     // When current penalty ends
	mu          sync.Mutex
}

// NewSecureRateLimiter creates a new secure rate limiter
func NewSecureRateLimiter(config *RateLimitConfig) *SecureRateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	rl := &SecureRateLimiter{
		config:          config,
		clients:         make(map[string]*ClientState),
		blocked:         make(map[string]time.Time),
		cleanupInterval: 5 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}

	// Start background cleanup
	go rl.cleanupLoop()

	return rl
}

// Allow checks if a request should be allowed
func (rl *SecureRateLimiter) Allow(ctx context.Context, identifier string) error {
	// Check if blocked
	if rl.isBlocked(identifier) {
		return ErrClientBlocked
	}

	// Get or create client state
	state := rl.getClientState(identifier)

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()

	// Check if in penalty period
	if now.Before(state.PenaltyEnd) {
		return ErrRateLimitExceeded
	}

	// Calculate effective window (with penalty)
	effectiveWindow := rl.config.WindowSize
	if state.Violations > 0 && rl.config.PenaltyEnabled {
		penalty := float64(effectiveWindow) * pow(rl.config.PenaltyFactor, float64(state.Violations))
		effectiveWindow = time.Duration(penalty)
		if effectiveWindow > rl.config.MaxPenalty {
			effectiveWindow = rl.config.MaxPenalty
		}
	}

	// Remove old requests outside the window
	windowStart := now.Add(-effectiveWindow)
	validRequests := make([]time.Time, 0, len(state.Requests))
	for _, t := range state.Requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}
	state.Requests = validRequests

	// Recover burst tokens
	if rl.config.BurstSize > 0 {
		tokensSinceLastBurst := int(now.Sub(state.LastBurst) / rl.config.BurstRecovery)
		if tokensSinceLastBurst > 0 {
			state.BurstTokens += tokensSinceLastBurst
			if state.BurstTokens > rl.config.BurstSize {
				state.BurstTokens = rl.config.BurstSize
			}
			state.LastBurst = now
		}
	}

	// Check rate limit
	if len(state.Requests) >= rl.config.MaxRequests {
		// Try to use burst token
		if state.BurstTokens > 0 {
			state.BurstTokens--
			state.Requests = append(state.Requests, now)
			return nil
		}

		// Rate limit exceeded
		state.Violations++

		// Check if should block
		if state.Violations >= rl.config.BlockThreshold {
			rl.block(identifier)
			return ErrClientBlocked
		}

		// Apply penalty
		if rl.config.PenaltyEnabled {
			penaltyDuration := time.Duration(float64(rl.config.WindowSize) * pow(rl.config.PenaltyFactor, float64(state.Violations)))
			if penaltyDuration > rl.config.MaxPenalty {
				penaltyDuration = rl.config.MaxPenalty
			}
			state.PenaltyEnd = now.Add(penaltyDuration)
		}

		return ErrRateLimitExceeded
	}

	// Allow request
	state.Requests = append(state.Requests, now)

	// Decay violations over time (reward good behavior)
	if len(state.Requests) < rl.config.MaxRequests/2 && state.Violations > 0 {
		state.Violations--
	}

	return nil
}

// RemainingRequests returns the number of remaining requests in the current window
func (rl *SecureRateLimiter) RemainingRequests(identifier string) int {
	state := rl.getClientState(identifier)

	state.mu.Lock()
	defer state.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.config.WindowSize)

	count := 0
	for _, t := range state.Requests {
		if t.After(windowStart) {
			count++
		}
	}

	remaining := rl.config.MaxRequests - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining
}

// ResetTime returns when the rate limit resets
func (rl *SecureRateLimiter) ResetTime(identifier string) time.Time {
	state := rl.getClientState(identifier)

	state.mu.Lock()
	defer state.mu.Unlock()

	if len(state.Requests) == 0 {
		return time.Now()
	}

	// Find oldest request in window
	oldest := state.Requests[0]
	for _, t := range state.Requests {
		if t.Before(oldest) {
			oldest = t
		}
	}

	return oldest.Add(rl.config.WindowSize)
}

func (rl *SecureRateLimiter) getClientState(identifier string) *ClientState {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	state, exists := rl.clients[identifier]
	if !exists {
		state = &ClientState{
			Requests:    make([]time.Time, 0, rl.config.MaxRequests),
			BurstTokens: rl.config.BurstSize,
			LastBurst:   time.Now(),
		}
		rl.clients[identifier] = state
	}

	return state
}

func (rl *SecureRateLimiter) isBlocked(identifier string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	unblockTime, exists := rl.blocked[identifier]
	if !exists {
		return false
	}

	return time.Now().Before(unblockTime)
}

func (rl *SecureRateLimiter) block(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.blocked[identifier] = time.Now().Add(rl.config.BlockDuration)
}

// Unblock manually unblocks a client
func (rl *SecureRateLimiter) Unblock(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.blocked, identifier)
	if state, exists := rl.clients[identifier]; exists {
		state.mu.Lock()
		state.Violations = 0
		state.PenaltyEnd = time.Time{}
		state.mu.Unlock()
	}
}

func (rl *SecureRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			return
		}
	}
}

func (rl *SecureRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Clean up blocked clients
	for id, unblockTime := range rl.blocked {
		if now.After(unblockTime) {
			delete(rl.blocked, id)
		}
	}

	// Clean up inactive clients
	for id, state := range rl.clients {
		state.mu.Lock()
		if len(state.Requests) == 0 && now.After(state.PenaltyEnd) {
			delete(rl.clients, id)
		}
		state.mu.Unlock()
	}
}

// Stop stops the rate limiter's background cleanup
func (rl *SecureRateLimiter) Stop() {
	close(rl.stopCleanup)
}

// Stats returns rate limiter statistics
func (rl *SecureRateLimiter) Stats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"active_clients": len(rl.clients),
		"blocked_clients": len(rl.blocked),
	}
}

// Helper function for power calculation
func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	result := base
	for i := 1; i < int(exp); i++ {
		result *= base
	}
	return result
}

// SlidingWindowLimiter implements a more precise sliding window rate limiter
type SlidingWindowLimiter struct {
	windowSize    time.Duration
	maxRequests   int
	windows       map[string]*SlidingWindow
	mu            sync.RWMutex
	stopCleanup   chan struct{}
}

// SlidingWindow represents a sliding window counter
type SlidingWindow struct {
	currentCount  int
	previousCount int
	currentStart  time.Time
	mu            sync.Mutex
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter
func NewSlidingWindowLimiter(windowSize time.Duration, maxRequests int) *SlidingWindowLimiter {
	swl := &SlidingWindowLimiter{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		windows:     make(map[string]*SlidingWindow),
		stopCleanup: make(chan struct{}),
	}

	go swl.cleanupLoop()
	return swl
}

// Allow checks if a request should be allowed
func (swl *SlidingWindowLimiter) Allow(identifier string) bool {
	swl.mu.Lock()
	window, exists := swl.windows[identifier]
	if !exists {
		window = &SlidingWindow{
			currentStart: time.Now().Truncate(swl.windowSize),
		}
		swl.windows[identifier] = window
	}
	swl.mu.Unlock()

	window.mu.Lock()
	defer window.mu.Unlock()

	now := time.Now()
	currentWindow := now.Truncate(swl.windowSize)

	// Check if we need to advance the window
	if currentWindow.After(window.currentStart) {
		if currentWindow.Sub(window.currentStart) >= swl.windowSize {
			// More than one window has passed
			window.previousCount = 0
		} else {
			window.previousCount = window.currentCount
		}
		window.currentCount = 0
		window.currentStart = currentWindow
	}

	// Calculate weighted count using sliding window algorithm
	elapsed := now.Sub(currentWindow)
	weight := float64(swl.windowSize-elapsed) / float64(swl.windowSize)
	weightedCount := float64(window.previousCount)*weight + float64(window.currentCount)

	if int(weightedCount) >= swl.maxRequests {
		return false
	}

	window.currentCount++
	return true
}

func (swl *SlidingWindowLimiter) cleanupLoop() {
	ticker := time.NewTicker(swl.windowSize * 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			swl.cleanup()
		case <-swl.stopCleanup:
			return
		}
	}
}

func (swl *SlidingWindowLimiter) cleanup() {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	threshold := time.Now().Add(-swl.windowSize * 2)
	for id, window := range swl.windows {
		window.mu.Lock()
		if window.currentStart.Before(threshold) {
			delete(swl.windows, id)
		}
		window.mu.Unlock()
	}
}

// Stop stops the rate limiter
func (swl *SlidingWindowLimiter) Stop() {
	close(swl.stopCleanup)
}

// IPBasedRateLimiter provides IP-based rate limiting with subnet detection
type IPBasedRateLimiter struct {
	limiter       *SecureRateLimiter
	subnetLimiter *SecureRateLimiter
	subnetPrefix  int // /24 for IPv4, /48 for IPv6
}

// NewIPBasedRateLimiter creates a new IP-based rate limiter
func NewIPBasedRateLimiter(perIPConfig, perSubnetConfig *RateLimitConfig) *IPBasedRateLimiter {
	return &IPBasedRateLimiter{
		limiter:       NewSecureRateLimiter(perIPConfig),
		subnetLimiter: NewSecureRateLimiter(perSubnetConfig),
		subnetPrefix:  24,
	}
}

// Allow checks if a request from the given IP should be allowed
func (ipl *IPBasedRateLimiter) Allow(ctx context.Context, ip string) error {
	// Check per-IP limit
	if err := ipl.limiter.Allow(ctx, ip); err != nil {
		return err
	}

	// Check per-subnet limit
	subnet := extractSubnet(ip, ipl.subnetPrefix)
	if subnet != "" {
		if err := ipl.subnetLimiter.Allow(ctx, fmt.Sprintf("subnet:%s", subnet)); err != nil {
			return fmt.Errorf("subnet rate limit: %w", err)
		}
	}

	return nil
}

// extractSubnet extracts the subnet prefix from an IP address
func extractSubnet(ip string, prefix int) string {
	// Simplified implementation - in production, use net.ParseIP
	// For IPv4, just take the first 3 octets for /24
	parts := make([]byte, 0)
	for i, c := range ip {
		if c == '.' {
			if len(parts) >= 2 {
				return ip[:i]
			}
			parts = append(parts, 1)
		}
	}
	return ""
}
