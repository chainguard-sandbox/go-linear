// Package main demonstrates how to create a label in Linear.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one team
//
// This example shows:
// 1. Getting a team for the label
// 2. Creating a label with name and color
// 3. Verifying the created label
//
// Labels are used to categorize and filter issues.
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/create_label.go
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

	// Step 1: Get a team (labels are team-scoped)
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		log.Fatalf("Failed to get teams: %v", err)
	}
	if len(teams.Nodes) == 0 {
		log.Fatal("No teams found")
	}

	teamID := teams.Nodes[0].ID
	teamName := teams.Nodes[0].Name
	fmt.Printf("Creating label in team: %s\n", teamName)
	fmt.Println()

	// Step 2: Create the label
	name := "go-linear-example"
	color := "#FF6B6B" // Red color

	label, err := client.IssueLabelCreate(ctx, linear.IssueLabelCreateInput{
		TeamID: &teamID,
		Name:   name,
		Color:  &color,
	})
	if err != nil {
		log.Fatalf("Failed to create label: %v", err)
	}

	// Step 3: Display the created label
	fmt.Printf("✓ Label created successfully!\n")
	fmt.Printf("  ID: %s\n", label.ID)
	fmt.Printf("  Name: %s\n", label.Name)
	fmt.Printf("  Color: %s\n", label.Color)

	fmt.Println("\nYou can now use this label when creating or updating issues:")
	fmt.Printf("  LabelIDs: []string{\"%s\"}\n", label.ID)

	// Optional: Clean up (delete the label)
	// Uncomment to auto-delete after creation
	// if err := client.IssueLabelDelete(ctx, label.ID); err != nil {
	// 	log.Printf("Warning: Failed to clean up label: %v", err)
	// }
}
