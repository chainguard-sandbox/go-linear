// Package main demonstrates efficient bulk operations.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one team
//
// This example shows:
// 1. Creating multiple issues efficiently
// 2. Rate limit management for bulk operations
// 3. Error handling in batch scenarios
// 4. Progress tracking and logging
//
// Best practices for batch operations:
// - Respect rate limits (~2 req/sec sustained)
// - Use concurrent requests with semaphore
// - Implement exponential backoff on failures
// - Log progress for long-running operations
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/batch_operations.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	// Create client with rate limit monitoring
	var rateLimitMu sync.Mutex
	lastRateLimit := time.Now()

	client, err := linear.NewClient(apiKey,
		linear.WithRateLimitCallback(func(info *linear.RateLimitInfo) {
			rateLimitMu.Lock()
			defer rateLimitMu.Unlock()

			if time.Since(lastRateLimit) > 5*time.Second {
				fmt.Printf("[Rate Limit] %d/%d requests remaining\n",
					info.RequestsRemaining, info.RequestsLimit)
				lastRateLimit = time.Now()
			}
		}),
		linear.WithRetry(5, 1*time.Second, 30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get team for issue creation
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		log.Fatalf("Failed to get teams: %v", err)
	}
	if len(teams.Nodes) == 0 {
		log.Fatal("No teams found")
	}
	teamID := teams.Nodes[0].ID

	// Define batch of issues to create
	batchSize := 10
	fmt.Printf("Creating %d issues in batch...\n\n", batchSize)

	// Use semaphore to limit concurrency (respect rate limits)
	sem := make(chan struct{}, 2) // Max 2 concurrent requests
	var wg sync.WaitGroup
	var resultsMu sync.Mutex
	var created []string
	var failed int

	start := time.Now()

	for i := 1; i <= batchSize; i++ {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(num int) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			title := fmt.Sprintf("Batch issue #%d", num)
			description := fmt.Sprintf("Created as part of batch operation demo (issue %d of %d)", num, batchSize)

			issue, err := client.IssueCreate(ctx, linear.IssueCreateInput{
				TeamID:      teamID,
				Title:       &title,
				Description: &description,
			})

			resultsMu.Lock()
			defer resultsMu.Unlock()

			if err != nil {
				fmt.Printf("  ✗ Issue #%d failed: %v\n", num, err)
				failed++
				return
			}

			created = append(created, issue.ID)
			fmt.Printf("  ✓ Issue #%d created: [%.0f] %s\n", num, issue.Number, issue.Title)
		}(i)

		// Throttle: Wait 500ms between launches to respect rate limits
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for all operations to complete
	wg.Wait()

	duration := time.Since(start)

	// Display results
	fmt.Printf("\n=== Batch Operation Complete ===\n")
	fmt.Printf("Created: %d issues\n", len(created))
	fmt.Printf("Failed: %d issues\n", failed)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Rate: %.2f issues/sec\n", float64(batchSize)/duration.Seconds())

	fmt.Println("\n✓ Batch Operation Best Practices:")
	fmt.Println("  - Use semaphore to limit concurrency (max 2-3 concurrent requests)")
	fmt.Println("  - Throttle between launches (500ms delay)")
	fmt.Println("  - Monitor rate limits with WithRateLimitCallback")
	fmt.Println("  - Configure retry with WithRetry for transient failures")
	fmt.Println("  - Log progress for long-running operations")
	fmt.Println("  - Use sync primitives for result aggregation")

	// Optional: Clean up created issues
	if len(created) > 0 {
		fmt.Printf("\nCreated issue IDs (for cleanup): %v\n", created)
	}
}
