// Package linear provides a Go client library for the Linear API.
//
// The Linear API is a GraphQL API for managing issues, projects, teams,
// and other resources in Linear (https://linear.app).
//
// # Authentication
//
// Create a client with your API key from https://linear.app/settings/account/security:
//
//	client, err := linear.NewClient("lin_api_xxx")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Usage
//
// Get the authenticated user:
//
//	viewer, err := client.Viewer(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Hello, %s!\n", viewer.Name)
//
// This package is automatically synchronized with the official Linear TypeScript SDK
// at https://github.com/linear/linear to ensure API compatibility and comprehensive coverage.
package linear
