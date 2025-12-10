// Package main demonstrates how to create an issue in Linear.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one team
//
// This example shows:
// 1. Authenticating with the Linear API
// 2. Getting the first available team
// 3. Creating an issue with title and description
// 4. Verifying the created issue
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/create_issue.go
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

	// Step 1: Get the first team (required for creating issues)
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		log.Fatalf("Failed to get teams: %v", err)
	}
	if len(teams.Nodes) == 0 {
		log.Fatal("No teams found. Create a team in Linear first.")
	}

	teamID := teams.Nodes[0].ID
	teamName := teams.Nodes[0].Name
	fmt.Printf("Using team: %s (ID: %s)\n", teamName, teamID)

	// Step 2: Create the issue
	title := "Example issue from go-linear"
	description := "This issue was created programmatically using the go-linear SDK.\n\n" +
		"It demonstrates:\n" +
		"- Issue creation with title and description\n" +
		"- Team assignment\n" +
		"- Basic error handling"

	issue, err := client.IssueCreate(ctx, linear.IssueCreateInput{
		TeamID:      teamID,
		Title:       &title,
		Description: &description,
	})
	if err != nil {
		log.Fatalf("Failed to create issue: %v", err)
	}

	// Step 3: Display the created issue
	fmt.Printf("\n✓ Issue created successfully!\n")
	fmt.Printf("  ID: %s\n", issue.ID)
	fmt.Printf("  Number: %.0f\n", issue.Number)
	fmt.Printf("  Title: %s\n", issue.Title)
	fmt.Printf("  URL: %s\n", issue.URL)
	fmt.Printf("  State: %s\n", issue.State.Name)

	// Optional: Clean up (delete the issue)
	// Uncomment to auto-delete after creation
	// if err := client.IssueDelete(ctx, issue.ID); err != nil {
	// 	log.Printf("Warning: Failed to clean up issue: %v", err)
	// }
}
