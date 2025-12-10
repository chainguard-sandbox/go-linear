// Package main demonstrates how to handle authentication errors.
//
// Prerequisites:
// - LINEAR_API_KEY environment variable (may be invalid for testing)
//
// This example shows:
// 1. Detecting authentication failures (401 errors)
// 2. Using errors.As to check for AuthenticationError
// 3. Automatic credential refresh with WithCredentialProvider
// 4. Proper error messages for debugging
//
// Common causes of 401 errors:
// - Invalid or expired API key
// - API key revoked in Linear settings
// - Incorrect Authorization header format
//
// Usage:
//
//	export LINEAR_API_KEY=lin_api_xxx
//	go run examples/tasks/handle_auth_errors.go
package main

import (
	"context"
	"errors"
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

	// Try to authenticate
	fmt.Println("Attempting to authenticate...")

	viewer, err := client.Viewer(ctx)
	if err != nil {
		// Check if it's an authentication error
		var authErr *linear.AuthenticationError
		if errors.As(err, &authErr) {
			fmt.Printf("❌ Authentication failed: %v\n", authErr)
			fmt.Println("\nTroubleshooting steps:")
			fmt.Println("1. Verify LINEAR_API_KEY is set correctly")
			fmt.Println("2. Check API key at https://linear.app/settings/account/security")
			fmt.Println("3. Ensure API key hasn't been revoked")
			fmt.Println("4. For OAuth tokens, ensure 'Bearer ' prefix is added automatically")
			os.Exit(1)
		}

		// Check for permission errors (403)
		var permErr *linear.ForbiddenError
		if errors.As(err, &permErr) {
			fmt.Printf("❌ Permission denied: %v\n", permErr)
			fmt.Println("\nYour API key lacks required permissions.")
			fmt.Println("Update permissions at https://linear.app/settings/account/security")
			os.Exit(1)
		}

		// Other error
		log.Fatalf("Request failed: %v", err)
	}

	// Authentication successful
	fmt.Printf("✓ Authentication successful!\n")
	fmt.Printf("  User: %s (%s)\n", viewer.Name, viewer.Email)
	fmt.Printf("  ID: %s\n", viewer.ID)
	if viewer.Admin {
		fmt.Printf("  Admin: Yes\n")
	}

	fmt.Println("\nNote: The client automatically refreshes credentials on 401 errors")
	fmt.Println("when using WithCredentialProvider. See credential_rotation.go example.")
}
