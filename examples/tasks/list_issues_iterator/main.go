// Package main demonstrates automatic pagination using IssueIterator.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with existing issues
//
// This example shows:
// 1. Creating an IssueIterator for automatic pagination
// 2. Iterating through all issues without manual cursor management
// 3. Handling io.EOF to detect end of results
// 4. Thread-safety guarantees of iterators
//
// IssueIterator automatically:
// - Fetches pages as needed
// - Manages pagination cursors
// - Is safe for concurrent use (mutex-protected)
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/list_issues_iterator.go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/eslerm/go-linear/pkg/linear"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		log.Fatal("LINEAR_API_KEY environment variable not set")
	}

	// Create client
	client, err := linear.NewClient(apiKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Create iterator with page size of 50
	// The iterator will automatically fetch more pages as needed
	iter := linear.NewIssueIterator(client, 50)

	fmt.Println("Fetching all issues (automatic pagination)...")
	fmt.Println()

	count := 0
	for {
		// Get next issue
		// This automatically fetches the next page when needed
		issue, err := iter.Next(ctx)

		// Check for end of results
		if errors.Is(err, io.EOF) {
			break
		}

		// Handle other errors
		if err != nil {
			log.Fatalf("Error fetching issue: %v", err)
		}

		// Process the issue
		count++
		fmt.Printf("%d. [%.0f] %s\n", count, issue.Number, issue.Title)

		// Show state if available
		if issue.State.Name != "" {
			fmt.Printf("   State: %s\n", issue.State.Name)
		}

		// Optional: Limit for demo purposes
		if count >= 20 {
			fmt.Printf("\n... (showing first 20 issues only)\n")
			break
		}
	}

	fmt.Printf("\nTotal issues fetched: %d\n", count)

	// Note: IssueIterator is thread-safe
	// Multiple goroutines can safely call Next() concurrently
	// The iterator handles synchronization internally
}
