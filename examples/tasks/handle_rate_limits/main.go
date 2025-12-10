// Package main demonstrates how to handle rate limit errors.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace
//
// This example shows:
// 1. Making API requests that might hit rate limits
// 2. Detecting RateLimitError using errors.As
// 3. Respecting the Retry-After header
// 4. Automatic retry behavior of the client
//
// Linear API rate limits (as of 2025):
// - ~2 requests per second sustained
// - Burst allowance for short spikes
// - 429 status code when exceeded
//
// The client automatically retries 429 errors with exponential backoff.
// You can also implement custom retry logic using RateLimitError.
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/handle_rate_limits.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eslerm/go-linear/pkg/linear"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	// Create client with custom retry configuration
	client, err := linear.NewClient(apiKey,
		// Configure aggressive retry for demonstration
		linear.WithRetry(5, 500*time.Millisecond, 60*time.Second),
		linear.WithMaxRetryDuration(2*time.Minute),
		// Add rate limit callback to monitor limits
		linear.WithRateLimitCallback(func(info *linear.RateLimitInfo) {
			fmt.Printf("[Rate Limit] Requests: %d/%d remaining | Complexity: %d/%d remaining\n",
				info.RequestsRemaining, info.RequestsLimit,
				info.ComplexityRemaining, info.ComplexityLimit)

			// Warn if approaching limit
			if info.RequestsRemaining < info.RequestsLimit/10 {
				fmt.Printf("⚠️  WARNING: Approaching rate limit (%d requests left)\n", info.RequestsRemaining)
			}
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Example: Make requests and handle rate limits
	fmt.Println("Making API requests (watching for rate limits)...")
	fmt.Println()

	for i := 1; i <= 5; i++ {
		fmt.Printf("Request %d: ", i)

		_, err := client.Viewer(ctx)
		if err != nil {
			// Check if it's a rate limit error
			var rateLimitErr *linear.RateLimitError
			if errors.As(err, &rateLimitErr) {
				fmt.Printf("Rate limited! Retry after %d seconds\n", rateLimitErr.RetryAfter)
				fmt.Printf("   Client already retried automatically.\n")
				fmt.Printf("   For custom retry: time.Sleep(%d * time.Second)\n", rateLimitErr.RetryAfter)
				continue
			}

			log.Fatalf("Request failed: %v", err)
		}

		fmt.Printf("Success\n")

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n✓ All requests completed")
	fmt.Println("\nRate Limit Best Practices:")
	fmt.Println("1. Client automatically retries 429 errors with exponential backoff")
	fmt.Println("2. Use WithRateLimitCallback to monitor remaining quota")
	fmt.Println("3. Implement request throttling (max 2 req/sec sustained)")
	fmt.Println("4. Use batch operations when available")
	fmt.Println("5. Cache frequently accessed data (teams, users, states)")
}
