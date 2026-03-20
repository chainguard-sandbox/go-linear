// Package main demonstrates how to star (favorite) issues and projects.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one team
//
// This example shows:
// 1. Creating a test issue
// 2. Starring the issue (add to favorites)
// 3. Unstarring the issue
// 4. Cleaning up test data
//
// Favorites appear in Linear's "Favorites" section for quick access.
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/favorites
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
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

	// Step 1: Get team and create test issue
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		log.Fatalf("Failed to get teams: %v", err)
	}
	if len(teams.Nodes) == 0 {
		log.Fatal("No teams found")
	}
	teamID := teams.Nodes[0].ID

	fmt.Println("Creating test issue...")
	title := "Important Issue - Favorites Demo"
	desc := "This issue will be starred for quick access"
	issue, err := client.IssueCreate(ctx, intgraphql.IssueCreateInput{
		TeamID:      teamID,
		Title:       &title,
		Description: &desc,
	})
	if err != nil {
		log.Fatalf("Failed to create issue: %v", err)
	}
	fmt.Printf("Created issue: [%.0f] %s\n", issue.Number, issue.Title)
	fmt.Println()

	// Step 2: Star the issue (add to favorites)
	fmt.Println("Starring issue (add to favorites)...")
	favorite, err := client.FavoriteCreate(ctx, intgraphql.FavoriteCreateInput{
		IssueID: &issue.ID,
	})
	if err != nil {
		log.Fatalf("Failed to create favorite: %v", err)
	}

	fmt.Printf("✓ Issue starred!\n")
	fmt.Printf("  Favorite ID: %s\n", favorite.ID)
	fmt.Printf("  Type: %s\n", favorite.Type)
	if favorite.Issue != nil {
		fmt.Printf("  Issue: %s\n", favorite.Issue.Title)
	}
	fmt.Println()

	// Step 3: Unstar the issue (remove from favorites)
	fmt.Println("Unstarring issue...")
	err = client.FavoriteDelete(ctx, favorite.ID)
	if err != nil {
		log.Fatalf("Failed to delete favorite: %v", err)
	}
	fmt.Println("✓ Issue unstarred!")
	fmt.Println()

	// Step 4: Clean up test issue
	fmt.Println("Cleaning up test issue...")
	if err := client.IssueDelete(ctx, issue.ID, nil); err != nil {
		log.Printf("Warning: Failed to delete issue: %v", err)
	}
	fmt.Println("✓ Test issue deleted")
	fmt.Println()

	fmt.Println("Favorites Features:")
	fmt.Println("  - Star issues for quick access")
	fmt.Println("  - Star projects for tracking")
	fmt.Println("  - Appears in Linear's Favorites sidebar")
	fmt.Println()
	fmt.Println("Use cases:")
	fmt.Println("  - Pin important bugs for daily review")
	fmt.Println("  - Bookmark issues you're actively working on")
	fmt.Println("  - Track key projects across teams")
	fmt.Println("  - Quick access to frequently referenced items")
}
