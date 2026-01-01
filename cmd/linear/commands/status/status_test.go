package status

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func TestNewStatusCommand(t *testing.T) {
	cmd := NewStatusCommand()

	if cmd.Use != "status" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status")
	}
	if cmd.Flags().Lookup("output") == nil {
		t.Error("Expected output flag")
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

func TestDisplayTable(t *testing.T) {
	cmd := NewStatusCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	testInfo := &linear.RateLimitInfo{
		RequestsLimit:       1000,
		RequestsRemaining:   950,
		RequestsReset:       time.Now().Add(time.Hour),
		ComplexityLimit:     10000,
		ComplexityRemaining: 9500,
		ComplexityReset:     time.Now().Add(time.Hour),
	}
	timestamp := time.Now()

	displayTable(cmd, testInfo, timestamp)

	output := buf.String()

	// Check headers
	if !strings.Contains(output, "LINEAR API RATE LIMIT STATUS") {
		t.Error("Expected header 'LINEAR API RATE LIMIT STATUS'")
	}
	if !strings.Contains(output, "REQUEST-BASED LIMITS") {
		t.Error("Expected 'REQUEST-BASED LIMITS' section")
	}
	if !strings.Contains(output, "COMPLEXITY-BASED LIMITS") {
		t.Error("Expected 'COMPLEXITY-BASED LIMITS' section")
	}

	// Check values
	if !strings.Contains(output, "950 / 1000") {
		t.Error("Expected request values '950 / 1000'")
	}
	if !strings.Contains(output, "9500 / 10000") {
		t.Error("Expected complexity values '9500 / 10000'")
	}
	if !strings.Contains(output, "95.0%") {
		t.Error("Expected percentage '95.0%'")
	}
	if !strings.Contains(output, "Last updated:") {
		t.Error("Expected 'Last updated:' timestamp")
	}
}

func TestDisplayTableZeroResetTimes(t *testing.T) {
	cmd := NewStatusCommand()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Test with zero reset times (should not show "Resets in")
	testInfo := &linear.RateLimitInfo{
		RequestsLimit:       1000,
		RequestsRemaining:   950,
		RequestsReset:       time.Time{}, // Zero time
		ComplexityLimit:     10000,
		ComplexityRemaining: 9500,
		ComplexityReset:     time.Time{}, // Zero time
	}
	timestamp := time.Now()

	displayTable(cmd, testInfo, timestamp)

	output := buf.String()

	// Should not contain "Resets in" when times are zero
	if strings.Contains(output, "Resets in:") {
		t.Error("Should not show 'Resets in:' for zero reset times")
	}
}

func TestRunStatusMissingAPIKey(t *testing.T) {
	// Save and clear the API key
	origKey := os.Getenv("LINEAR_API_KEY")
	os.Setenv("LINEAR_API_KEY", "")
	defer func() {
		if origKey != "" {
			os.Setenv("LINEAR_API_KEY", origKey)
		} else {
			os.Unsetenv("LINEAR_API_KEY")
		}
	}()

	cmd := NewStatusCommand()
	var buf bytes.Buffer
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
	for i := 0; i < 10; i++ {
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
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = capture.get()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Final get should succeed
	info, _ := capture.get()
	if info == nil {
		t.Error("Expected non-nil info after concurrent access")
	}
}
