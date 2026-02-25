// Package main demonstrates notification subscription management.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one team
//
// This example shows:
// 1. Creating a test issue
// 2. Subscribing to notifications for the issue
// 3. Unsubscribing from notifications
// 4. Cleaning up test data
//
// Notification subscriptions allow programmatic notification management.
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/notifications
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
	title := "Important Issue - Subscribe for updates"
	issue, err := client.IssueCreate(ctx, intgraphql.IssueCreateInput{
		TeamID: teamID,
		Title:  &title,
	})
	if err != nil {
		log.Fatalf("Failed to create issue: %v", err)
	}
	fmt.Printf("Created issue: [%.0f] %s\n", issue.Number, issue.Title)
	fmt.Println()

	// Step 2: Create a project to subscribe to (issues don't support subscriptions directly)
	fmt.Println("Creating test project for subscription demo...")
	projectName := "Test Project - Notification Demo"
	project, err := client.ProjectCreate(ctx, intgraphql.ProjectCreateInput{
		TeamIds: []string{teamID},
		Name:    projectName,
	})
	if err != nil {
		log.Fatalf("Failed to create project: %v", err)
	}
	fmt.Printf("Created project: %s\n", project.Name)
	fmt.Println()

	// Subscribe to project notifications
	fmt.Println("Subscribing to project notifications...")
	subscription, err := client.NotificationSubscriptionCreate(ctx, intgraphql.NotificationSubscriptionCreateInput{
		ProjectID: &project.ID,
	})
	if err != nil {
		fmt.Printf("Note: Notification subscriptions may require specific workspace configuration\n")
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Skipping subscription demo...")
	} else {
		fmt.Printf("✓ Subscribed!\n")
		fmt.Printf("  Subscription ID: %s\n", subscription.ID)
		fmt.Printf("  Active: %v\n", subscription.Active)
		fmt.Println()

		// Step 3: Unsubscribe
		fmt.Println("Unsubscribing from notifications...")
		err = client.NotificationSubscriptionDelete(ctx, subscription.ID)
		if err != nil {
			log.Printf("Warning: Failed to delete subscription: %v", err)
		} else {
			fmt.Println("✓ Unsubscribed!")
		}
	}
	fmt.Println()

	// Step 4: Clean up test data
	fmt.Println("Cleaning up test data...")
	if err := client.IssueDelete(ctx, issue.ID, nil); err != nil {
		log.Printf("Warning: Failed to delete issue: %v", err)
	}
	if err := client.ProjectDelete(ctx, project.ID); err != nil {
		log.Printf("Warning: Failed to delete project: %v", err)
	}
	fmt.Println("✓ Test data deleted")
	fmt.Println()

	fmt.Println("Notification Features:")
	fmt.Println("  - Subscribe to issue/project updates")
	fmt.Println("  - Mark notifications as read programmatically")
	fmt.Println("  - Archive notifications for cleanup")
	fmt.Println()
	fmt.Println("Use cases:")
	fmt.Println("  - Auto-subscribe team to critical issues")
	fmt.Println("  - Bulk mark-as-read for notification cleanup")
	fmt.Println("  - Programmatic notification inbox management")
	fmt.Println("  - Integration with external notification systems")
}
