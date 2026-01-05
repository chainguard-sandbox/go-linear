// Package cli provides shared types for CLI commands.
package cli

import "github.com/chainguard-sandbox/go-linear/pkg/linear"

// ClientFactory creates a Linear API client.
// Used by commands to defer client creation until execution time,
// enabling lazy initialization and testability.
type ClientFactory func() (*linear.Client, error)
