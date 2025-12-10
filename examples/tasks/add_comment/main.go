// Package main demonstrates how to add a comment to an issue.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one existing issue
//
// This example shows:
// 1. Finding an existing issue
// 2. Adding a comment to the issue
// 3. Verifying the comment was created
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/add_comment.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	// Step 1: Get an existing issue
	first := int64(1)
	issues, err := client.Issues(ctx, &first, nil)
	if err != nil {
		log.Fatalf("Failed to get issues: %v", err)
	}
	if len(issues.Nodes) == 0 {
		log.Fatal("No issues found. Create an issue first.")
	}

	issue := issues.Nodes[0]
	fmt.Printf("Adding comment to issue:\n")
	fmt.Printf("  Number: %.0f\n", issue.Number)
	fmt.Printf("  Title: %s\n", issue.Title)
	fmt.Println()

	// Step 2: Create the comment
	body := "This comment was added programmatically using go-linear SDK.\n\n" +
		"Comments support **markdown** formatting:\n" +
		"- Bold, italic, code blocks\n" +
		"- Links: [Linear](https://linear.app)\n" +
		"- Mentions: @username\n" +
		"- Code: `fmt.Println(\"Hello\")`"

	issueID := issue.ID
	comment, err := client.CommentCreate(ctx, linear.CommentCreateInput{
		IssueID: &issueID,
		Body:    &body,
	})
	if err != nil {
		log.Fatalf("Failed to create comment: %v", err)
	}

	// Step 3: Display the created comment
	fmt.Printf("✓ Comment created successfully!\n")
	fmt.Printf("  ID: %s\n", comment.ID)
	fmt.Printf("  Body: %s\n", comment.Body)
	fmt.Printf("  Created: %s\n", comment.CreatedAt)
	fmt.Printf("  URL: %s\n", comment.URL)
}
