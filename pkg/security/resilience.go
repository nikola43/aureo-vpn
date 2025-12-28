// Package security provides resilience patterns for production systems
package security

import (
	"context"
	cryptoRand "crypto/rand"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var secureRand = cryptoRand.Reader

// Circuit breaker states
const (
	StateClosed   = iota // Normal operation
	StateOpen            // Circuit open, failing fast
	StateHalfOpen        // Testing if service recovered
)

// Circuit breaker errors
var (
	ErrCircuitOpen         = errors.New("circuit breaker is open")
	ErrHalfOpenOverflow    = errors.New("too many requests in half-open state")
	ErrOperationTimeout    = errors.New("operation timed out")
)

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxFailures     int           // Number of failures before opening
	Timeout         time.Duration // Time to wait before half-open
	HalfOpenMax     int           // Max requests in half-open state
	SuccessThreshold int          // Successes needed to close from half-open
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:      5,
		Timeout:          30 * time.Second,
		HalfOpenMax:      3,
		SuccessThreshold: 2,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config           CircuitBreakerConfig
	state            int32
	failures         int32
	successes        int32
	halfOpenRequests int32
	lastFailure      time.Time
	mu               sync.RWMutex
	onStateChange    func(from, to int)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// OnStateChange sets a callback for state changes
func (cb *CircuitBreaker) OnStateChange(fn func(from, to int)) {
	cb.onStateChange = fn
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	err := fn()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}

	return err
}

// ExecuteWithContext runs the given function with circuit breaker and context
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	if !cb.allowRequest() {
		return ErrCircuitOpen
	}

	// Create a channel for the result
	done := make(chan error, 1)

	go func() {
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			cb.recordFailure()
		} else {
			cb.recordSuccess()
		}
		return err
	case <-ctx.Done():
		cb.recordFailure()
		return ErrOperationTimeout
	}
}

func (cb *CircuitBreaker) allowRequest() bool {
	state := atomic.LoadInt32(&cb.state)

	switch state {
	case StateClosed:
		return true

	case StateOpen:
		cb.mu.RLock()
		timeout := time.Since(cb.lastFailure) >= cb.config.Timeout
		cb.mu.RUnlock()

		if timeout {
			cb.transitionTo(StateHalfOpen)
			return true
		}
		return false

	case StateHalfOpen:
		if atomic.LoadInt32(&cb.halfOpenRequests) >= int32(cb.config.HalfOpenMax) {
			return false
		}
		atomic.AddInt32(&cb.halfOpenRequests, 1)
		return true
	}

	return false
}

func (cb *CircuitBreaker) recordFailure() {
	state := atomic.LoadInt32(&cb.state)

	switch state {
	case StateClosed:
		failures := atomic.AddInt32(&cb.failures, 1)
		if int(failures) >= cb.config.MaxFailures {
			cb.transitionTo(StateOpen)
		}

	case StateHalfOpen:
		// Any failure in half-open goes back to open
		cb.transitionTo(StateOpen)
	}

	cb.mu.Lock()
	cb.lastFailure = time.Now()
	cb.mu.Unlock()
}

func (cb *CircuitBreaker) recordSuccess() {
	state := atomic.LoadInt32(&cb.state)

	switch state {
	case StateClosed:
		// Reset failure count on success
		atomic.StoreInt32(&cb.failures, 0)

	case StateHalfOpen:
		successes := atomic.AddInt32(&cb.successes, 1)
		if int(successes) >= cb.config.SuccessThreshold {
			cb.transitionTo(StateClosed)
		}
	}
}

func (cb *CircuitBreaker) transitionTo(newState int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	oldState := atomic.SwapInt32(&cb.state, int32(newState))

	// Reset counters on state change
	atomic.StoreInt32(&cb.failures, 0)
	atomic.StoreInt32(&cb.successes, 0)
	atomic.StoreInt32(&cb.halfOpenRequests, 0)

	if cb.onStateChange != nil && int(oldState) != newState {
		cb.onStateChange(int(oldState), newState)
	}
}

// State returns the current state
func (cb *CircuitBreaker) State() int {
	return int(atomic.LoadInt32(&cb.state))
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.transitionTo(StateClosed)
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts     int
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	BackoffMultiplier float64
	Jitter          bool
}

// DefaultRetryConfig returns sensible retry defaults
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        10 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
	}
}

// Retry executes a function with exponential backoff retry
func Retry(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Don't wait after the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate wait time with optional jitter
		wait := backoff
		if config.Jitter {
			// Add up to 25% jitter
			jitterBytes := make([]byte, 1)
			if _, err := secureRand.Read(jitterBytes); err == nil {
				jitterPercent := float64(jitterBytes[0]) / 256.0 * 0.25
				wait = time.Duration(float64(wait) * (1 + jitterPercent))
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}

		// Increase backoff
		backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}
	}

	return lastErr
}

// GracefulShutdown manages graceful shutdown of services
type GracefulShutdown struct {
	timeout    time.Duration
	wg         sync.WaitGroup
	shutdownCh chan struct{}
	mu         sync.RWMutex
	started    bool
	services   []func(context.Context) error
}

// NewGracefulShutdown creates a new graceful shutdown manager
func NewGracefulShutdown(timeout time.Duration) *GracefulShutdown {
	return &GracefulShutdown{
		timeout:    timeout,
		shutdownCh: make(chan struct{}),
		services:   make([]func(context.Context) error, 0),
	}
}

// Register registers a shutdown function
func (gs *GracefulShutdown) Register(fn func(context.Context) error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.services = append(gs.services, fn)
}

// RegisterService wraps a service with WaitGroup tracking
func (gs *GracefulShutdown) RegisterService(fn func(context.Context) error) func(context.Context) {
	return func(ctx context.Context) {
		gs.wg.Add(1)
		defer gs.wg.Done()
		fn(ctx)
	}
}

// Shutdown initiates graceful shutdown
func (gs *GracefulShutdown) Shutdown(ctx context.Context) error {
	gs.mu.Lock()
	if gs.started {
		gs.mu.Unlock()
		return nil
	}
	gs.started = true
	close(gs.shutdownCh)
	services := make([]func(context.Context) error, len(gs.services))
	copy(services, gs.services)
	gs.mu.Unlock()

	// Create timeout context
	shutdownCtx, cancel := context.WithTimeout(ctx, gs.timeout)
	defer cancel()

	// Run all shutdown functions
	errCh := make(chan error, len(services))
	for _, fn := range services {
		go func(f func(context.Context) error) {
			errCh <- f(shutdownCtx)
		}(fn)
	}

	// Collect errors
	var errs []error
	for i := 0; i < len(services); i++ {
		if err := <-errCh; err != nil {
			errs = append(errs, err)
		}
	}

	// Wait for all goroutines with timeout
	done := make(chan struct{})
	go func() {
		gs.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-shutdownCtx.Done():
		return errors.New("shutdown timed out")
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// ShutdownCh returns the shutdown channel
func (gs *GracefulShutdown) ShutdownCh() <-chan struct{} {
	return gs.shutdownCh
}

// Bulkhead limits concurrent access to a resource
type Bulkhead struct {
	sem     chan struct{}
	timeout time.Duration
}

// NewBulkhead creates a new bulkhead with the given concurrency limit
func NewBulkhead(maxConcurrent int, timeout time.Duration) *Bulkhead {
	return &Bulkhead{
		sem:     make(chan struct{}, maxConcurrent),
		timeout: timeout,
	}
}

// Execute runs the function within the bulkhead
func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
	select {
	case b.sem <- struct{}{}:
		defer func() { <-b.sem }()
		return fn()
	case <-time.After(b.timeout):
		return ErrHalfOpenOverflow
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Available returns the number of available slots
func (b *Bulkhead) Available() int {
	return cap(b.sem) - len(b.sem)
}
