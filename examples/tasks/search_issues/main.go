// Package main demonstrates how to search for issues in Linear.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with existing issues
//
// This example shows:
// 1. Basic text search across all issues
// 2. Using search operators (priority:, state:, assignee:)
// 3. Pagination through search results
// 4. Filtering by multiple criteria
//
// Search operators supported:
// - "bug" - Contains word "bug"
// - "\"exact phrase\"" - Exact phrase match
// - "title:bug" - Search in title field only
// - "state:todo" - Filter by workflow state
// - "priority:high" - Filter by priority
// - "assignee:me" - Filter by assignee
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/search_issues.go
package main

import (
	"context"
	"fmt"
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

	// Example 1: Simple text search
	fmt.Println("=== Example 1: Simple text search ===")
	searchAndDisplay(ctx, client, "bug")

	// Example 2: Search with priority filter
	fmt.Println("\n=== Example 2: High priority issues ===")
	searchAndDisplay(ctx, client, "priority:1")

	// Example 3: Search with state filter
	fmt.Println("\n=== Example 3: Todo issues ===")
	searchAndDisplay(ctx, client, "state:todo")

	// Example 4: Combined filters
	fmt.Println("\n=== Example 4: High priority bugs in todo state ===")
	searchAndDisplay(ctx, client, "priority:1 state:todo bug")

	// Example 5: Search in title only
	fmt.Println("\n=== Example 5: Issues with 'API' in title ===")
	searchAndDisplay(ctx, client, "title:API")
}

func searchAndDisplay(ctx context.Context, client *linear.Client, query string) {
	first := int64(10)
	issues, err := client.IssueSearch(ctx, query, &first, nil)
	if err != nil {
		log.Printf("Search failed for query '%s': %v", query, err)
		return
	}

	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Found %d results\n", len(issues.Nodes))

	if len(issues.Nodes) == 0 {
		fmt.Println("No results found.")
		return
	}

	for i, issue := range issues.Nodes {
		fmt.Printf("%d. %s\n", i+1, issue.Title)
		fmt.Printf("   ID: %s\n", issue.ID)
		if issue.State.Name != "" {
			fmt.Printf("   State: %s", issue.State.Name)
		}
		priorities := []string{"None", "Urgent", "High", "Normal", "Low"}
		if int(issue.Priority) < len(priorities) {
			fmt.Printf(" | Priority: %s", priorities[int(issue.Priority)])
		}
		fmt.Println()
	}

	// Show pagination info
	if issues.PageInfo.HasNextPage {
		fmt.Printf("\n... and more results available (use cursor: %s)\n", *issues.PageInfo.EndCursor)
	}
}
