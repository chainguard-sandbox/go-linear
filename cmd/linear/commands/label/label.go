// Package label provides label-related commands for the Linear CLI.
package label

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
)

// cli.ClientFactory is a function that creates a Linear client.

// NewLabelCommand creates the label command group.
func NewLabelCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "label",
		Short: "Manage Linear labels",
		Long:  "Commands for listing Linear issue labels.",
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))
	cmd.AddCommand(NewCreateCommand(clientFactory))
	cmd.AddCommand(NewUpdateCommand(clientFactory))
	cmd.AddCommand(NewDeleteCommand(clientFactory))

	return cmd
}
