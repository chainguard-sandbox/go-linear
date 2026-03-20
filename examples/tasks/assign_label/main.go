// Package main demonstrates how to assign a label to an issue.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one issue and one label
//
// This example shows:
// 1. Finding an existing issue
// 2. Finding available labels
// 3. Adding a label to the issue using IssueUpdate
// 4. Verifying the label was applied
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/assign_label.go
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

	// Step 1: Get an existing issue
	firstIssue := int64(1)
	issues, err := client.Issues(ctx, &firstIssue, nil)
	if err != nil {
		log.Fatalf("Failed to get issues: %v", err)
	}
	if len(issues.Nodes) == 0 {
		log.Fatal("No issues found. Create an issue first.")
	}

	issue := issues.Nodes[0]
	fmt.Printf("Found issue:\n")
	fmt.Printf("  Number: %.0f\n", issue.Number)
	fmt.Printf("  Title: %s\n", issue.Title)
	fmt.Println()

	// Step 2: Get available labels
	firstLabel := int64(5)
	labels, err := client.IssueLabels(ctx, &firstLabel, nil)
	if err != nil {
		log.Fatalf("Failed to get labels: %v", err)
	}
	if len(labels.Nodes) == 0 {
		log.Fatal("No labels found. Create a label first (see create_label.go)")
	}

	// Pick the first label
	label := labels.Nodes[0]
	fmt.Printf("Using label: %s (color: %s)\n", label.Name, label.Color)
	fmt.Println()

	// Step 3: Assign the label to the issue
	// Note: This replaces existing labels. To add to existing labels,
	// first get current labelIDs, then append the new one.
	labelIDs := []string{label.ID}

	updated, err := client.IssueUpdate(ctx, issue.ID, linear.IssueUpdateInput{
		AddedLabelIds: labelIDs,
	})
	if err != nil {
		log.Fatalf("Failed to assign label: %v", err)
	}

	// Step 4: Verify the label was applied
	fmt.Printf("✓ Label assigned successfully!\n")
	fmt.Printf("  Issue ID: %s\n", updated.ID)
	fmt.Printf("  Issue Title: %s\n", updated.Title)
	fmt.Printf("  Label added: %s\n", label.Name)

	fmt.Println("\nNote: To add to existing labels (not replace), fetch current labels first:")
	fmt.Println("  existingIDs := make([]string, len(issue.Labels.Nodes))")
	fmt.Println("  for i, lbl := range issue.Labels.Nodes {")
	fmt.Println("      existingIDs[i] = lbl.ID")
	fmt.Println("  }")
	fmt.Println("  LabelIDs: append(existingIDs, newLabelID)")
}
