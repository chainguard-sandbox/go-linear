// Package main demonstrates how to list all labels in a workspace.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace (labels are optional)
//
// This example shows:
// 1. Listing all labels in the workspace
// 2. Displaying label details (name, color, team)
// 3. Filtering labels by team
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/list_labels.go
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

	// Fetch all labels
	first := int64(100)
	labels, err := client.IssueLabels(ctx, &first, nil)
	if err != nil {
		log.Fatalf("Failed to get labels: %v", err)
	}

	if len(labels.Nodes) == 0 {
		fmt.Println("No labels found in workspace.")
		fmt.Println("Create labels in Linear or use create_label.go")
		return
	}

	fmt.Printf("Found %d labels:\n\n", len(labels.Nodes))

	// Group labels by team
	type LabelInfo struct {
		ID          string
		Name        string
		Color       string
		Description *string
	}
	teamLabels := make(map[string][]LabelInfo)

	for _, label := range labels.Nodes {
		teamName := "Workspace Labels"
		teamLabels[teamName] = append(teamLabels[teamName], LabelInfo{
			ID:          label.ID,
			Name:        label.Name,
			Color:       label.Color,
			Description: label.Description,
		})
	}

	// Display labels grouped by team
	for teamName, lbls := range teamLabels {
		fmt.Printf("=== %s ===\n", teamName)
		for i, label := range lbls {
			fmt.Printf("%d. %s\n", i+1, label.Name)
			fmt.Printf("   Color: %s\n", label.Color)
			fmt.Printf("   ID: %s\n", label.ID)
			if label.Description != nil && *label.Description != "" {
				fmt.Printf("   Description: %s\n", *label.Description)
			}
			fmt.Println()
		}
	}

	// Show pagination info
	if labels.PageInfo.HasNextPage {
		fmt.Printf("More labels available (showing first %d)\n", first)
		fmt.Printf("Use cursor for next page: %s\n", *labels.PageInfo.EndCursor)
	} else {
		fmt.Println("All labels retrieved.")
	}
}
