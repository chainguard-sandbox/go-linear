// Package main demonstrates how to use the circuit breaker pattern.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
//
// This example shows:
// 1. Configuring a circuit breaker for fail-fast behavior
// 2. How the circuit opens after consecutive failures
// 3. How the circuit recovers in half-open state
// 4. Handling ErrCircuitOpen errors
//
// Circuit Breaker States:
// - Closed: Normal operation, requests allowed
// - Open: Circuit tripped, requests fail immediately with ErrCircuitOpen
// - HalfOpen: Testing recovery, single request allowed
//
// Usage:
//
//	export LINEAR_API_KEY=invalid_key_for_testing
//	go run examples/tasks/handle_circuit_breaker.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	// Create circuit breaker with low thresholds for demonstration
	cb := &linear.CircuitBreaker{
		MaxFailures:  3,                // Open after 3 consecutive failures
		ResetTimeout: 10 * time.Second, // Try recovery after 10 seconds
	}

	// Create client with circuit breaker
	client, err := linear.NewClient(apiKey,
		linear.WithCircuitBreaker(cb),
		linear.WithRetry(1, 100*time.Millisecond, 1*time.Second), // Fast retries for demo
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	fmt.Println("Circuit Breaker Demo")
	fmt.Printf("Configuration: MaxFailures=%d, ResetTimeout=%v\n", cb.MaxFailures, cb.ResetTimeout)
	fmt.Println()

	// Make requests to demonstrate circuit breaker behavior
	for i := 1; i <= 10; i++ {
		fmt.Printf("Request %d: ", i)

		_, err := client.Viewer(ctx)
		if err != nil {
			// Check if circuit breaker opened
			if errors.Is(err, linear.ErrCircuitOpen) {
				fmt.Printf("⚠️  Circuit breaker is OPEN (fail-fast)\n")
				fmt.Printf("   API is experiencing issues. Request blocked to prevent cascading failures.\n")
				fmt.Printf("   Circuit will attempt recovery in %v\n", cb.ResetTimeout)

				// Wait for circuit to enter half-open state
				if i == 5 {
					fmt.Printf("\n   Waiting %v for circuit breaker to reset...\n\n", cb.ResetTimeout)
					time.Sleep(cb.ResetTimeout + 1*time.Second)
				}
				continue
			}

			// Other errors (auth, network, etc.)
			fmt.Printf("Failed: %v\n", err)
			continue
		}

		fmt.Printf("Success\n")
	}

	fmt.Println("\n✓ Circuit Breaker Behavior:")
	fmt.Println("  - After 3 failures: Circuit opens")
	fmt.Println("  - While open: Requests fail immediately (no API calls made)")
	fmt.Println("  - After timeout: Circuit enters half-open state")
	fmt.Println("  - On success: Circuit closes and resumes normal operation")
	fmt.Println("  - On failure: Circuit opens again")

	fmt.Println("\nUse Case:")
	fmt.Println("  - Prevents cascading failures during API outages")
	fmt.Println("  - Reduces load on failing systems")
	fmt.Println("  - Automatic recovery when service is healthy")
	fmt.Println("  - Essential for production deployments")
}
