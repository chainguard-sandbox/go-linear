// Package main demonstrates how to update an existing issue in Linear.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one existing issue
//
// This example shows:
// 1. Finding an existing issue
// 2. Updating the issue title and description
// 3. Changing priority and state
// 4. Verifying the updates
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/update_issue.go
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

	// Step 1: Get an existing issue (fetch the first one)
	first := int64(1)
	issues, err := client.Issues(ctx, &first, nil)
	if err != nil {
		log.Fatalf("Failed to get issues: %v", err)
	}
	if len(issues.Nodes) == 0 {
		log.Fatal("No issues found. Create an issue first.")
	}

	issue := issues.Nodes[0]
	fmt.Printf("Found issue to update:\n")
	fmt.Printf("  ID: %s\n", issue.ID)
	fmt.Printf("  Number: %.0f\n", issue.Number)
	fmt.Printf("  Current title: %s\n", issue.Title)
	fmt.Printf("  Current priority: %.0f\n", issue.Priority)
	fmt.Println()

	// Step 2: Update the issue
	updatedTitle := fmt.Sprintf("%s (Updated)", issue.Title)
	updatedDescription := "This issue was updated programmatically using go-linear SDK.\n\n" +
		"Updates include:\n" +
		"- Modified title\n" +
		"- Added description\n" +
		"- Changed priority to High"
	priority := int64(2) // High priority

	updated, err := client.IssueUpdate(ctx, issue.ID, linear.IssueUpdateInput{
		Title:       &updatedTitle,
		Description: &updatedDescription,
		Priority:    &priority,
	})
	if err != nil {
		log.Fatalf("Failed to update issue: %v", err)
	}

	// Step 3: Display the updated issue
	fmt.Printf("✓ Issue updated successfully!\n")
	fmt.Printf("  ID: %s\n", updated.ID)
	fmt.Printf("  New title: %s\n", updated.Title)
	priorities := []string{"None", "Urgent", "High", "Normal", "Low"}
	if int(updated.Priority) < len(priorities) {
		fmt.Printf("  New priority: %s (%.0f)\n", priorities[int(updated.Priority)], updated.Priority)
	}

	fmt.Println("\nNote: Only fields provided in IssueUpdateInput are modified.")
	fmt.Println("Fields set to nil are left unchanged.")
}
