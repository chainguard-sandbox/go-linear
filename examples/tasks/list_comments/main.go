// Package main demonstrates how to list comments on an issue.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable set
// - A Linear workspace with existing issues and comments
//
// This example shows:
// 1. Finding an existing issue
// 2. Listing all comments on that issue
// 3. Displaying comment details (author, timestamp, body)
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/list_comments.go
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

	// Step 1: Get recent comments (all comments across workspace)
	first := int64(20)
	comments, err := client.Comments(ctx, &first, nil)
	if err != nil {
		log.Fatalf("Failed to get comments: %v", err)
	}

	if len(comments.Nodes) == 0 {
		fmt.Println("No comments found in workspace.")
		fmt.Println("Add a comment using add_comment.go first.")
		return
	}

	// Step 2: Display all comments
	fmt.Printf("Found %d comments:\n\n", len(comments.Nodes))

	for i, comment := range comments.Nodes {
		fmt.Printf("%d. Comment ID: %s\n", i+1, comment.ID)

		// Show timestamps
		fmt.Printf("   Created: %s\n", comment.CreatedAt)
		if comment.UpdatedAt != comment.CreatedAt {
			fmt.Printf("   Updated: %s\n", comment.UpdatedAt)
		}

		// Show body (truncate if long)
		body := comment.Body
		if len(body) > 100 {
			body = body[:100] + "..."
		}
		fmt.Printf("   Body: %s\n", body)
		fmt.Printf("   URL: %s\n", comment.URL)
		fmt.Println()
	}

	// Show pagination status
	if comments.PageInfo.HasNextPage {
		fmt.Printf("More comments available. Use cursor: %s\n", *comments.PageInfo.EndCursor)
	} else {
		fmt.Println("All comments retrieved.")
	}
}
