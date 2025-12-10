// Package main demonstrates manual pagination through issues.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with existing issues
//
// This example shows:
// 1. Fetching the first page of issues
// 2. Using PageInfo to check for more pages
// 3. Using EndCursor to fetch the next page
// 4. Manual cursor-based pagination control
//
// Use this approach when you need fine-grained control over pagination.
// For automatic pagination, see list_issues_iterator.go
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/list_issues_paginated.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
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

	// Configure pagination
	pageSize := int64(10)
	var cursor *string = nil
	totalIssues := 0
	pageNum := 0

	fmt.Printf("Fetching issues (page size: %d)\n", pageSize)
	fmt.Println()

	// Paginate through all issues
	for {
		pageNum++
		fmt.Printf("=== Page %d ===\n", pageNum)

		// Fetch one page of issues
		issues, err := client.Issues(ctx, &pageSize, cursor)
		if err != nil {
			log.Fatalf("Failed to fetch issues: %v", err)
		}

		// Display issues on this page
		for i, issue := range issues.Nodes {
			totalIssues++
			fmt.Printf("%d. [%.0f] %s\n", i+1, issue.Number, issue.Title)
			if issue.State.Name != "" {
				fmt.Printf("   State: %s\n", issue.State.Name)
			}
		}

		fmt.Printf("Issues on this page: %d\n", len(issues.Nodes))

		// Check if there are more pages
		if !issues.PageInfo.HasNextPage {
			fmt.Println("\n✓ No more pages available")
			break
		}

		// Get cursor for next page
		cursor = issues.PageInfo.EndCursor
		fmt.Printf("More pages available (cursor: %s...)\n\n", (*cursor)[:20])

		// Optional: Limit pages for demo
		if pageNum >= 3 {
			fmt.Printf("... (stopping after %d pages for demo)\n", pageNum)
			break
		}
	}

	fmt.Printf("\nTotal issues fetched: %d\n", totalIssues)
	fmt.Printf("Total pages fetched: %d\n", pageNum)

	// Note: Maximum page size is typically 250
	// Recommended page size is 50-100 for best performance
}
