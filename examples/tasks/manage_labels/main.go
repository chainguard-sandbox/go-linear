// Package main demonstrates simplified label management with IssueAddLabel and IssueRemoveLabel.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one team
//
// This example shows:
// 1. Creating a test issue
// 2. Creating a test label
// 3. Adding the label to the issue (simple method, no array manipulation)
// 4. Removing the label from the issue (simple method)
// 5. Cleaning up test data
//
// IssueAddLabel vs IssueUpdate:
// - Old way: Fetch existing labels → build array → call IssueUpdate
// - New way: Just call IssueAddLabel(issueID, labelID)
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/manage_labels
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
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

	// Step 1: Get team
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		log.Fatalf("Failed to get teams: %v", err)
	}
	if len(teams.Nodes) == 0 {
		log.Fatal("No teams found")
	}
	teamID := teams.Nodes[0].ID

	// Step 2: Create test issue
	fmt.Println("Creating test issue...")
	title := "Test Issue - Label Management Demo"
	issue, err := client.IssueCreate(ctx, intgraphql.IssueCreateInput{
		TeamID: teamID,
		Title:  &title,
	})
	if err != nil {
		log.Fatalf("Failed to create issue: %v", err)
	}
	fmt.Printf("Created issue: [%.0f] %s\n", issue.Number, issue.Title)
	fmt.Println()

	// Step 3: Create test label
	fmt.Println("Creating test label...")
	labelName := "test-label-demo"
	labelColor := "#FF6B6B"
	label, err := client.IssueLabelCreate(ctx, intgraphql.IssueLabelCreateInput{
		TeamID: &teamID,
		Name:   labelName,
		Color:  &labelColor,
	})
	if err != nil {
		log.Fatalf("Failed to create label: %v", err)
	}
	fmt.Printf("Created label: %s (ID: %s)\n", label.Name, label.ID)
	fmt.Println()

	// Step 4: Add label to issue (SIMPLE METHOD - no array manipulation!)
	fmt.Printf("Adding label '%s' to issue...\n", label.Name)
	updatedIssue, err := client.IssueAddLabel(ctx, issue.ID, label.ID)
	if err != nil {
		log.Fatalf("Failed to add label: %v", err)
	}

	fmt.Printf("✓ Label added!\n")
	fmt.Printf("  Issue now has %d labels:\n", len(updatedIssue.Labels.Nodes))
	for _, lbl := range updatedIssue.Labels.Nodes {
		fmt.Printf("    - %s\n", lbl.Name)
	}
	fmt.Println()

	// Step 5: Remove label from issue (SIMPLE METHOD - no array filtering!)
	fmt.Printf("Removing label '%s' from issue...\n", label.Name)
	removedIssue, err := client.IssueRemoveLabel(ctx, issue.ID, label.ID)
	if err != nil {
		log.Fatalf("Failed to remove label: %v", err)
	}

	fmt.Printf("✓ Label removed!\n")
	fmt.Printf("  Issue now has %d labels\n", len(removedIssue.Labels.Nodes))
	fmt.Println()

	// Step 6: Clean up test data
	fmt.Println("Cleaning up test data...")
	if err := client.IssueDelete(ctx, issue.ID, nil); err != nil {
		log.Printf("Warning: Failed to delete issue: %v", err)
	}
	if err := client.IssueLabelDelete(ctx, label.ID); err != nil {
		log.Printf("Warning: Failed to delete label: %v", err)
	}
	fmt.Println("✓ Test data deleted")
	fmt.Println()

	fmt.Println("Label Management Comparison:")
	fmt.Println()
	fmt.Println("Old way (IssueUpdate with arrays):")
	fmt.Println("  1. Fetch issue to get existing label IDs")
	fmt.Println("  2. Build new array with added/removed labels")
	fmt.Println("  3. Call IssueUpdate with entire array")
	fmt.Println("  4. Risk: May lose labels if concurrent updates")
	fmt.Println()
	fmt.Println("New way (IssueAddLabel/IssueRemoveLabel):")
	fmt.Println("  1. Call IssueAddLabel(issueID, labelID)")
	fmt.Println("  2. That's it!")
	fmt.Println("  3. Atomic operation, no race conditions")
}
