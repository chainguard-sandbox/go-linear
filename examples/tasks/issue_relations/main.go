// Package main demonstrates how to create relationships between issues.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one team
//
// This example shows:
// 1. Creating two test issues
// 2. Creating a "blocks" relationship
// 3. Updating the relationship type
// 4. Deleting the relationship
// 5. Cleaning up test issues
//
// Relationship types:
// - "blocks": This issue blocks another (prevents work)
// - "blocked": This issue is blocked by another
// - "duplicate": This issue is a duplicate of another
// - "related": This issue is related to another
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/issue_relations
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

	// Step 1: Get team for creating issues
	teams, err := client.Teams(ctx, nil, nil)
	if err != nil {
		log.Fatalf("Failed to get teams: %v", err)
	}
	if len(teams.Nodes) == 0 {
		log.Fatal("No teams found")
	}
	teamID := teams.Nodes[0].ID

	// Step 2: Create two test issues
	fmt.Println("Creating test issues...")

	title1 := "Test Issue 1 (blocked)"
	desc1 := "This issue is blocked by Issue 2"
	issue1, err := client.IssueCreate(ctx, intgraphql.IssueCreateInput{
		TeamID:      teamID,
		Title:       &title1,
		Description: &desc1,
	})
	if err != nil {
		log.Fatalf("Failed to create issue 1: %v", err)
	}

	title2 := "Test Issue 2 (blocker)"
	desc2 := "This issue blocks Issue 1"
	issue2, err := client.IssueCreate(ctx, intgraphql.IssueCreateInput{
		TeamID:      teamID,
		Title:       &title2,
		Description: &desc2,
	})
	if err != nil {
		log.Fatalf("Failed to create issue 2: %v", err)
	}

	fmt.Printf("Created Issue 1: [%.0f] %s\n", issue1.Number, issue1.Title)
	fmt.Printf("Created Issue 2: [%.0f] %s\n", issue2.Number, issue2.Title)
	fmt.Println()

	// Step 3: Create a "blocks" relationship
	// Issue 2 blocks Issue 1
	fmt.Printf("Creating relationship: Issue 2 blocks Issue 1\n")

	relation, err := client.IssueRelationCreate(ctx, intgraphql.IssueRelationCreateInput{
		IssueID:        issue2.ID,
		RelatedIssueID: issue1.ID,
		Type:           intgraphql.IssueRelationTypeBlocks,
	})
	if err != nil {
		log.Fatalf("Failed to create relation: %v", err)
	}

	fmt.Printf("✓ Relation created!\n")
	fmt.Printf("  ID: %s\n", relation.ID)
	fmt.Printf("  Type: %s\n", relation.Type)
	fmt.Printf("  %s blocks %s\n", relation.Issue.Title, relation.RelatedIssue.Title)
	fmt.Println()

	// Step 4: Update the relationship type
	newTypeEnum := intgraphql.IssueRelationTypeRelated
	newTypeString := string(newTypeEnum)
	fmt.Printf("Updating relationship type to: related\n")

	updated, err := client.IssueRelationUpdate(ctx, relation.ID, intgraphql.IssueRelationUpdateInput{
		Type: &newTypeString,
	})
	if err != nil {
		log.Fatalf("Failed to update relation: %v", err)
	}

	fmt.Printf("✓ Relation updated!\n")
	fmt.Printf("  New type: %s\n", updated.Type)
	fmt.Println()

	// Step 5: Delete the relationship (cleanup)
	fmt.Printf("Deleting relationship (cleanup)...\n")

	err = client.IssueRelationDelete(ctx, relation.ID)
	if err != nil {
		log.Fatalf("Failed to delete relation: %v", err)
	}

	fmt.Printf("✓ Relation deleted!\n")
	fmt.Println()

	// Step 6: Clean up test issues
	fmt.Println("Cleaning up test issues...")
	if err := client.IssueDelete(ctx, issue1.ID); err != nil {
		log.Printf("Warning: Failed to delete issue 1: %v", err)
	}
	if err := client.IssueDelete(ctx, issue2.ID); err != nil {
		log.Printf("Warning: Failed to delete issue 2: %v", err)
	}
	fmt.Println("✓ Test issues deleted")
	fmt.Println()

	fmt.Println("Issue Relation Types:")
	fmt.Println("  - blocks: This issue blocks another (prevents progress)")
	fmt.Println("  - blocked: This issue is blocked by another")
	fmt.Println("  - duplicate: This issue is a duplicate")
	fmt.Println("  - related: General relationship")
	fmt.Println()
	fmt.Println("Use cases:")
	fmt.Println("  - Track dependencies (Issue A blocks Issue B)")
	fmt.Println("  - Mark duplicates (avoid duplicate work)")
	fmt.Println("  - Link related work (provide context)")
}
