// Package cli provides shared types for CLI commands.
package cli

import (
	"io"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// ClientFactory creates a Linear API client.
// Used by commands to defer client creation until execution time,
// enabling lazy initialization and testability.
type ClientFactory func() (*linear.Client, error)

// Closer represents a client that can be closed.
// Used for resource cleanup in defer statements.
type Closer interface {
	Close() error
}

// Testing Patterns
//
// Commands can define small interfaces for testability. Example:
//
//	// In your command package:
//	type issueGetter interface {
//	    Issue(ctx context.Context, id string) (*intgraphql.GetIssue_Issue, error)
//	}
//
//	func runGet(client issueGetter, id string) error { ... }
//
// The [*linear.Client] type satisfies any such interface automatically.
// For unit tests, create a mock that implements only the needed methods:
//
//	type mockClient struct{}
//	func (m *mockClient) Issue(ctx context.Context, id string) (*intgraphql.GetIssue_Issue, error) {
//	    return &intgraphql.GetIssue_Issue{ID: id, Title: "Test"}, nil
//	}
//
// For integration tests, use the mock HTTP server pattern in helpers_test.go.

// Common interfaces for frequently-used operations.
// Commands needing these operations can use these interfaces for easier mocking.

// EntityLister provides paginated list operations.
type EntityLister[T any] interface {
	List(first *int64, after *string) (*T, error)
}

// OutputWriter abstracts output writing for testability.
type OutputWriter interface {
	io.Writer
}
