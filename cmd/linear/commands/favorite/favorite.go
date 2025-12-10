// Package favorite provides favorite (star) commands for the Linear CLI.
package favorite

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// ClientFactory is a function that creates a Linear client.
type ClientFactory func() (*linear.Client, error)

// NewFavoriteCommand creates the favorite command group.
func NewFavoriteCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "favorite",
		Short: "Manage Linear favorites (starred items)",
		Long:  "Commands for starring and unstarring issues, projects, cycles, and documents for quick access.",
	}

	// Add subcommands
	cmd.AddCommand(NewCreateCommand(clientFactory))
	cmd.AddCommand(NewDeleteCommand(clientFactory))

	return cmd
}
