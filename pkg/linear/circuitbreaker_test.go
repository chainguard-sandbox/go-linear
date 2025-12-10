package linear

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_Closed(t *testing.T) {
	cb := &CircuitBreaker{
		MaxFailures:  3,
		ResetTimeout: 1 * time.Second,
	}

	// Should allow requests when closed
	if err := cb.Allow(); err != nil {
		t.Errorf("Allow() error = %v, want nil", err)
	}

	if cb.State() != "closed" {
		t.Errorf("State() = %q, want %q", cb.State(), "closed")
	}
}

func TestCircuitBreaker_Opens(t *testing.T) {
	cb := &CircuitBreaker{
		MaxFailures:  3,
		ResetTimeout: 1 * time.Second,
	}

	// Record failures
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	// Circuit should be open
	if cb.State() != "open" {
		t.Errorf("State() = %q, want %q after 3 failures", cb.State(), "open")
	}

	// Should reject requests
	if err := cb.Allow(); !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Allow() error = %v, want %v", err, ErrCircuitOpen)
	}
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	cb := &CircuitBreaker{
		MaxFailures:  2,
		ResetTimeout: 100 * time.Millisecond,
	}

	// Open circuit
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != "open" {
		t.Fatalf("State() = %q, want %q", cb.State(), "open")
	}

	// Wait for reset timeout
	time.Sleep(150 * time.Millisecond)

	// Should transition to half-open
	if err := cb.Allow(); err != nil {
		t.Errorf("Allow() after timeout error = %v, want nil", err)
	}

	if cb.State() != "half-open" {
		t.Errorf("State() = %q, want %q", cb.State(), "half-open")
	}
}

func TestCircuitBreaker_Recovery(t *testing.T) {
	cb := &CircuitBreaker{
		MaxFailures:  2,
		ResetTimeout: 50 * time.Millisecond,
	}

	// Open circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait and transition to half-open
	time.Sleep(60 * time.Millisecond)
	_ = cb.Allow()

	// Record success - should close
	cb.RecordSuccess()

	if cb.State() != "closed" {
		t.Errorf("State() = %q, want %q after success", cb.State(), "closed")
	}

	// Should allow requests
	if err := cb.Allow(); err != nil {
		t.Errorf("Allow() after recovery error = %v, want nil", err)
	}
}
