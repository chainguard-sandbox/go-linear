// Package main demonstrates how to update an existing comment.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with at least one issue with comments
//
// This example shows:
// 1. Finding an existing issue
// 2. Listing comments on that issue
// 3. Updating the first comment's body
// 4. Verifying the update
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/update_comment.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

	// Step 1: Get an issue with comments
	first := int64(10)
	comments, err := client.Comments(ctx, &first, nil)
	if err != nil {
		log.Fatalf("Failed to get comments: %v", err)
	}
	if len(comments.Nodes) == 0 {
		log.Fatal("No comments found. Create a comment first (see add_comment.go)")
	}

	comment := comments.Nodes[0]
	fmt.Printf("Found comment to update:\n")
	fmt.Printf("  ID: %s\n", comment.ID)
	fmt.Printf("  Current body: %s\n", comment.Body)
	fmt.Printf("  Created: %s\n", comment.CreatedAt)
	fmt.Println()

	// Step 2: Update the comment
	updatedBody := fmt.Sprintf("%s\n\n---\n*Updated at %s by go-linear SDK*",
		comment.Body,
		time.Now().Format(time.RFC3339))

	updated, err := client.CommentUpdate(ctx, comment.ID, linear.CommentUpdateInput{
		Body: &updatedBody,
	})
	if err != nil {
		log.Fatalf("Failed to update comment: %v", err)
	}

	// Step 3: Display the updated comment
	fmt.Printf("✓ Comment updated successfully!\n")
	fmt.Printf("  ID: %s\n", updated.ID)
	fmt.Printf("  New body:\n%s\n", updated.Body)
	fmt.Printf("  Updated: %s\n", updated.UpdatedAt)
}
