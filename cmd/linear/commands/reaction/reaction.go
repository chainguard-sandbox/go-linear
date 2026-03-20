// Package reaction provides emoji reaction commands for the Linear CLI.
package reaction

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
)

// cli.ClientFactory is a function that creates a Linear client.

// NewReactionCommand creates the reaction command group.
func NewReactionCommand(clientFactory cli.ClientFactory) *cobra.Command {
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
