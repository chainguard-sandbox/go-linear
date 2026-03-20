// Package cycle provides cycle-related commands for the Linear CLI.
package cycle

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
)

// cli.ClientFactory is a function that creates a Linear client.

// NewCycleCommand creates the cycle command group.
func NewCycleCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cycle",
		Short: "Manage Linear cycles (sprints)",
		Long:  "Commands for listing and viewing Linear cycles and sprints.",
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))
	cmd.AddCommand(NewCreateCommand(clientFactory))
	cmd.AddCommand(NewUpdateCommand(clientFactory))
	cmd.AddCommand(NewArchiveCommand(clientFactory))

	return cmd
}
