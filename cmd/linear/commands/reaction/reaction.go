// Package reaction provides emoji reaction commands for the Linear CLI.
package reaction

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// ClientFactory is a function that creates a Linear client.
type ClientFactory func() (*linear.Client, error)

// NewReactionCommand creates the reaction command group.
func NewReactionCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reaction",
		Short: "Manage emoji reactions",
		Long:  "Commands for adding and removing emoji reactions on issues and comments.",
	}

	// Add subcommands
	cmd.AddCommand(NewCreateCommand(clientFactory))
	cmd.AddCommand(NewDeleteCommand(clientFactory))

	return cmd
}
