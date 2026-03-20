// Package main demonstrates how to delete an issue in Linear.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one team
//
// This example shows:
// 1. Creating a temporary test issue
// 2. Deleting the issue
// 3. Verifying the deletion
//
// Warning: This permanently deletes issues. Use with caution!
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/delete_issue.go
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

	// Step 1: Get a team for creating a test issue
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		log.Fatalf("Failed to get teams: %v", err)
	}
	if len(teams.Nodes) == 0 {
		log.Fatal("No teams found")
	}

	teamID := teams.Nodes[0].ID

	// Step 2: Create a temporary test issue
	title := "Test issue for deletion (safe to delete)"
	description := "This issue was created for demonstration purposes and will be deleted immediately."

	issue, err := client.IssueCreate(ctx, linear.IssueCreateInput{
		TeamID:      teamID,
		Title:       &title,
		Description: &description,
	})
	if err != nil {
		log.Fatalf("Failed to create test issue: %v", err)
	}

	fmt.Printf("Created test issue:\n")
	fmt.Printf("  ID: %s\n", issue.ID)
	fmt.Printf("  Number: %.0f\n", issue.Number)
	fmt.Printf("  Title: %s\n", issue.Title)
	fmt.Println()

	// Step 3: Delete the issue
	fmt.Printf("Deleting issue %.0f...\n", issue.Number)

	err = client.IssueDelete(ctx, issue.ID, nil)
	if err != nil {
		log.Fatalf("Failed to delete issue: %v", err)
	}

	// Step 4: Verify deletion
	fmt.Printf("✓ Issue %.0f deleted successfully!\n", issue.Number)
	fmt.Printf("  Issue ID was: %s\n", issue.ID)

	// Step 5: Try to fetch the deleted issue (should fail)
	fmt.Println("\nVerifying deletion...")
	_, err = client.Issue(ctx, issue.ID)
	if err != nil {
		fmt.Printf("✓ Confirmed: Issue no longer exists (error fetching: %v)\n", err)
	} else {
		fmt.Printf("⚠ Warning: Issue still exists after deletion\n")
	}
}
