// Package main demonstrates how to search for issues in Linear.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with existing issues
//
// This example shows:
// 1. Basic full-text search across all issues
// 2. Using structured filters (future: priority, state, team)
// 3. Pagination through search results
// 4. Including archived issues
//
// The searchIssues API provides:
// - Full-text search across titles, descriptions, and optionally comments
// - Structured filtering (cleaner than text operators)
// - Total count of results
// - Better performance than the deprecated issueSearch
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/search_issues
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
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
	searchSimple(ctx, client, "bug")

	// Example 2: Search for recent issues
	fmt.Println("\n=== Example 2: Search for 'example' ===")
	searchSimple(ctx, client, "example")

	// Example 3: Include archived issues
	fmt.Println("\n=== Example 3: Including archived issues ===")
	includeArchived := true
	searchWithOptions(ctx, client, "urgent", &includeArchived)
}

// searchSimple demonstrates basic text search
func searchSimple(ctx context.Context, client *linear.Client, term string) {
	first := int64(10)
	issues, err := client.SearchIssues(ctx, term, &first, nil, nil, nil)
	if err != nil {
		log.Printf("Search failed for term '%s': %v", term, err)
		return
	}

	fmt.Printf("Term: '%s'\n", term)
	fmt.Printf("Total results: %.0f\n", issues.TotalCount)
	fmt.Printf("Showing: %d results\n", len(issues.Nodes))

	if len(issues.Nodes) == 0 {
		fmt.Println("No results found.")
		return
	}

	for i, issue := range issues.Nodes {
		fmt.Printf("%d. [%.0f] %s\n", i+1, issue.Number, issue.Title)
		if issue.State.Name != "" {
			fmt.Printf("   State: %s", issue.State.Name)
		}
		if issue.Assignee != nil {
			fmt.Printf(" | Assigned: %s", issue.Assignee.Name)
		}
		fmt.Println()
	}

	if issues.PageInfo.HasNextPage {
		fmt.Printf("\n... more results available (cursor: %s)\n", (*issues.PageInfo.EndCursor)[:20]+"...")
	}
}

// searchWithOptions demonstrates additional options
func searchWithOptions(ctx context.Context, client *linear.Client, term string, includeArchived *bool) {
	first := int64(5)
	issues, err := client.SearchIssues(ctx, term, &first, nil, nil, includeArchived)
	if err != nil {
		log.Printf("Search failed: %v", err)
		return
	}

	archived := "false"
	if includeArchived != nil && *includeArchived {
		archived = "true"
	}

	fmt.Printf("Term: '%s' (includeArchived: %s)\n", term, archived)
	fmt.Printf("Total results: %.0f\n", issues.TotalCount)
	fmt.Printf("Showing: %d results\n", len(issues.Nodes))

	for i, issue := range issues.Nodes {
		fmt.Printf("%d. [%.0f] %s\n", i+1, issue.Number, issue.Title)
		if issue.Team.Name != "" {
			fmt.Printf("   Team: %s\n", issue.Team.Name)
		}
	}

	fmt.Println("\nNote: Structured filters (IssueFilter) will be exposed in future API.")
	fmt.Println("For now, use full-text search with descriptive terms.")
}
