package status

import (
	"strings"
	"testing"
	"time"

	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

func TestNewStatusCommand(t *testing.T) {
	cmd := NewStatusCommand()

	if cmd.Use != "status" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status")
	}
}

func TestRateLimitCapture(t *testing.T) {
	capture := &rateLimitCapture{}

	// Test initial state
	info, timestamp := capture.get()
	if info != nil {
		t.Error("Expected nil info initially")
	}
	if !timestamp.IsZero() {
		t.Error("Expected zero timestamp initially")
	}

	// Test update
	testInfo := &linear.RateLimitInfo{
		RequestsLimit:       1000,
		RequestsRemaining:   950,
		RequestsReset:       time.Now().Add(time.Hour),
		ComplexityLimit:     10000,
		ComplexityRemaining: 9500,
		ComplexityReset:     time.Now().Add(time.Hour),
	}
	capture.update(testInfo)

	// Test get after update
	info, timestamp = capture.get()
	if info == nil {
		t.Fatal("Expected non-nil info after update")
	}
	if info.RequestsLimit != 1000 {
		t.Errorf("RequestsLimit = %d, want 1000", info.RequestsLimit)
	}
	if info.RequestsRemaining != 950 {
		t.Errorf("RequestsRemaining = %d, want 950", info.RequestsRemaining)
	}
	if timestamp.IsZero() {
		t.Error("Expected non-zero timestamp after update")
	}
}

func TestRunStatusMissingAPIKey(t *testing.T) {
	// Set the API key to empty (t.Setenv automatically restores after test)
	t.Setenv("LINEAR_API_KEY", "")

	cmd := NewStatusCommand()
	var buf strings.Builder
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error when LINEAR_API_KEY is missing")
	}
	if !strings.Contains(err.Error(), "LINEAR_API_KEY") {
		t.Errorf("Expected error about LINEAR_API_KEY, got: %v", err)
	}
}

func TestRateLimitCaptureConcurrency(t *testing.T) {
	capture := &rateLimitCapture{}
	done := make(chan bool)

	// Concurrent updates
	for i := range 10 {
		go func(val int) {
			testInfo := &linear.RateLimitInfo{
				RequestsLimit:     1000,
				RequestsRemaining: val,
			}
			capture.update(testInfo)
			done <- true
		}(i)
	}

	// Concurrent reads
	for range 10 {
		go func() {
			_, _ = capture.get()
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 20 {
		<-done
	}

	// Final get should succeed
	info, _ := capture.get()
	if info == nil {
		t.Error("Expected non-nil info after concurrent access")
	}
}
