// Package main demonstrates how to add emoji reactions to issues.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one team
//
// This example shows:
// 1. Creating a test issue
// 2. Adding emoji reactions (👍, ❤️, 🎉)
// 3. Removing reactions
// 4. Cleaning up test data
//
// Reactions provide quick feedback without creating comments.
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/reactions
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
	title := "Great work on this feature!"
	issue, err := client.IssueCreate(ctx, intgraphql.IssueCreateInput{
		TeamID: teamID,
		Title:  &title,
	})
	if err != nil {
		log.Fatalf("Failed to create issue: %v", err)
	}
	fmt.Printf("Created issue: [%.0f] %s\n", issue.Number, issue.Title)
	fmt.Println()

	// Step 2: Add reactions
	emojis := []string{"👍", "❤️", "🎉"}
	reactions := make([]string, 0, len(emojis))

	for _, emoji := range emojis {
		fmt.Printf("Adding reaction: %s\n", emoji)
		reaction, err := client.ReactionCreate(ctx, intgraphql.ReactionCreateInput{
			Emoji:   emoji,
			IssueID: &issue.ID,
		})
		if err != nil {
			log.Printf("Warning: Failed to add reaction %s: %v\n", emoji, err)
			continue
		}
		fmt.Printf("  ✓ Reaction added (ID: %s)\n", reaction.ID)
		reactions = append(reactions, reaction.ID)
	}
	fmt.Println()

	// Step 3: Remove reactions
	fmt.Println("Removing reactions...")
	for _, reactionID := range reactions {
		err := client.ReactionDelete(ctx, reactionID)
		if err != nil {
			log.Printf("Warning: Failed to delete reaction: %v\n", err)
		}
	}
	fmt.Println("✓ All reactions removed")
	fmt.Println()

	// Step 4: Clean up test issue
	fmt.Println("Cleaning up test issue...")
	if err := client.IssueDelete(ctx, issue.ID, nil); err != nil {
		log.Printf("Warning: Failed to delete issue: %v", err)
	}
	fmt.Println("✓ Test issue deleted")
	fmt.Println()

	fmt.Println("Reaction Features:")
	fmt.Println("  - Quick feedback with emoji (no comment needed)")
	fmt.Println("  - React to issues or comments")
	fmt.Println("  - Common reactions: 👍 ❤️ 🎉 👀 🚀")
	fmt.Println()
	fmt.Println("Use cases:")
	fmt.Println("  - Acknowledge issues without commenting")
	fmt.Println("  - Show support for feature requests")
	fmt.Println("  - Quick team feedback on proposals")
}
