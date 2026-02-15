package producer

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open.
	ErrCircuitOpen = errors.New("circuit breaker is OPEN")
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	// StateClosed - Normal operation, requests pass through
	StateClosed CircuitState = iota

	// StateOpen - Circuit is open, all requests immediately fail
	StateOpen

	// StateHalfOpen - Circuit is testing if the service has recovered
	StateHalfOpen
)

// String returns the string representation of the circuit state.
func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern to prevent cascading failures.
//
// State transitions:
// - CLOSED → OPEN: After N consecutive failures
// - OPEN → HALF_OPEN: After timeout period
// - HALF_OPEN → CLOSED: After M consecutive successes
// - HALF_OPEN → OPEN: On any failure
type CircuitBreaker struct {
	state            CircuitState
	failureCount     int32
	successCount     int32
	failureThreshold int32        // Open circuit after this many failures (default: 5)
	successThreshold int32        // Close circuit after this many successes (default: 2)
	timeout          time.Duration // How long to wait before half-open (default: 60s)
	lastFailureTime  time.Time
	mu               sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration.
func NewCircuitBreaker(failureThreshold, successThreshold int32, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

// Call executes the given function if the circuit breaker allows it.
// Returns ErrCircuitOpen if the circuit is open.
func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.AllowRequest() {
		return ErrCircuitOpen
	}

	err := fn()
	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

// AllowRequest checks if a request should be allowed based on the circuit state.
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		// Normal operation - allow request
		return true

	case StateOpen:
		// Check if timeout has elapsed to transition to half-open
		if time.Since(cb.lastFailureTime) > cb.timeout {
			// Transition to half-open (will be done in write lock)
			cb.mu.RUnlock()
			cb.mu.Lock()
			defer cb.mu.Unlock()
			defer cb.mu.RLock()

			// Double-check state (another goroutine might have changed it)
			if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.timeout {
				cb.state = StateHalfOpen
				cb.successCount = 0
				cb.failureCount = 0
				return true // Allow one test request
			}
		}
		return false

	case StateHalfOpen:
		// Allow limited requests to test if service recovered
		return true

	default:
		return false
	}
}

// RecordSuccess records a successful operation.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// Already closed, reset failure count
		cb.failureCount = 0

	case StateHalfOpen:
		// Count successes in half-open state
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			// Enough successes - transition to closed
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
	}
}

// RecordFailure records a failed operation.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		// Count failures in closed state
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			// Too many failures - open the circuit
			cb.state = StateOpen
			cb.successCount = 0
		}

	case StateHalfOpen:
		// Any failure in half-open state reopens the circuit
		cb.state = StateOpen
		cb.failureCount = 0
		cb.successCount = 0
	}
}

// GetState returns the current state of the circuit breaker.
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailureCount returns the current failure count.
func (cb *CircuitBreaker) GetFailureCount() int32 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failureCount
}

// GetSuccessCount returns the current success count (in half-open state).
func (cb *CircuitBreaker) GetSuccessCount() int32 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.successCount
}

// Reset manually resets the circuit breaker to closed state.
// This should only be used for testing or manual intervention.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
}
