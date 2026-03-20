// Package main demonstrates safe concurrent API requests.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with issues
//
// This example shows:
// 1. Making concurrent API requests safely
// 2. Using goroutines with proper error handling
// 3. Synchronizing results with sync.WaitGroup
// 4. Thread-safety guarantees of the client and iterators
//
// Thread Safety:
// - Client is safe for concurrent use
// - Iterators (IssueIterator, TeamIterator) are mutex-protected
// - Multiple goroutines can share one client
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/concurrent_requests.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	// Create client (thread-safe, can be shared across goroutines)
	client, err := linear.NewClient(apiKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	fmt.Println("=== Concurrent Requests Demo ===")
	fmt.Println()

	// Example 1: Concurrent queries
	fmt.Println("Example 1: Fetching multiple resources concurrently")
	var wg sync.WaitGroup
	start := time.Now()

	// Launch 3 concurrent requests
	wg.Add(3)

	go func() {
		defer wg.Done()
		viewer, err := client.Viewer(ctx)
		if err != nil {
			log.Printf("Viewer query failed: %v", err)
			return
		}
		fmt.Printf("  ✓ Viewer: %s\n", viewer.Email)
	}()

	go func() {
		defer wg.Done()
		first := int64(5)
		issues, err := client.Issues(ctx, &first, nil)
		if err != nil {
			log.Printf("Issues query failed: %v", err)
			return
		}
		fmt.Printf("  ✓ Issues: %d retrieved\n", len(issues.Nodes))
	}()

	go func() {
		defer wg.Done()
		teams, err := client.Teams(ctx, nil, nil)
		if err != nil {
			log.Printf("Teams query failed: %v", err)
			return
		}
		fmt.Printf("  ✓ Teams: %d retrieved\n", len(teams.Nodes))
	}()

	wg.Wait()
	fmt.Printf("Completed in: %v\n\n", time.Since(start))

	// Example 2: Concurrent iterator usage
	fmt.Println("Example 2: Concurrent iteration (thread-safe)")

	iter := linear.NewIssueIterator(client, 10)
	var results sync.Map
	wg.Add(3)

	// Three goroutines reading from the same iterator
	for i := range 3 {
		workerID := i + 1
		go func() {
			defer wg.Done()

			count := 0
			for range 5 {
				_, err := iter.Next(ctx)
				if err != nil {
					return
				}
				count++
			}

			results.Store(workerID, count)
			fmt.Printf("  ✓ Worker %d: fetched %d issues\n", workerID, count)
		}()
	}

	wg.Wait()

	total := 0
	results.Range(func(key, value any) bool {
		total += value.(int)
		return true
	})
	fmt.Printf("Total issues fetched: %d\n\n", total)

	fmt.Println("✓ Thread Safety Guarantees:")
	fmt.Println("  - Client: Safe for concurrent use across goroutines")
	fmt.Println("  - Iterators: Mutex-protected, safe for concurrent Next() calls")
	fmt.Println("  - HTTP Client: Connection pooling handles concurrent requests")
	fmt.Println("  - No external synchronization needed")
}
