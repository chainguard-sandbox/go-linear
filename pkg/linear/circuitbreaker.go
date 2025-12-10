package linear

import (
	"errors"
	"sync"
	"time"
)

// CircuitBreaker prevents cascading failures by stopping requests when error rate is high.
//
// States:
//   - Closed: Normal operation, requests allowed
//   - Open: Circuit breaker tripped, requests fail fast
//   - HalfOpen: Testing if service recovered
//
// Example:
//
//	cb := &linear.CircuitBreaker{
//	    MaxFailures:  5,
//	    ResetTimeout: 60 * time.Second,
//	}
//	client, _ := linear.NewClient(apiKey, linear.WithCircuitBreaker(cb))
type CircuitBreaker struct {
	// MaxFailures is the number of failures before opening the circuit
	MaxFailures int

	// ResetTimeout is how long to wait before attempting to close the circuit
	ResetTimeout time.Duration

	mu            sync.RWMutex
	failures      int
	lastFailTime  time.Time
	state         circuitState
	nextAttemptAt time.Time
}

type circuitState int

const (
	stateClosed circuitState = iota
	stateOpen
	stateHalfOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

// RecordSuccess records a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = stateClosed
}

// RecordFailure records a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.MaxFailures {
		cb.state = stateOpen
		cb.nextAttemptAt = time.Now().Add(cb.ResetTimeout)
	}
}

// Allow checks if a request should be allowed.
// Returns ErrCircuitOpen if circuit is open.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.RLock()
	state := cb.state
	nextAttempt := cb.nextAttemptAt
	cb.mu.RUnlock()

	switch state {
	case stateClosed:
		return nil
	case stateOpen:
		// Check if reset timeout has passed
		if time.Now().After(nextAttempt) {
			cb.mu.Lock()
			cb.state = stateHalfOpen
			cb.mu.Unlock()
			return nil
		}
		return ErrCircuitOpen
	case stateHalfOpen:
		// Allow one request to test if service recovered
		return nil
	}

	return nil
}

// State returns the current circuit breaker state (for monitoring).
func (cb *CircuitBreaker) State() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case stateClosed:
		return "closed"
	case stateOpen:
		return "open"
	case stateHalfOpen:
		return "half-open"
	}
	return "unknown"
}
