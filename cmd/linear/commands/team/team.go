// Package team provides team-related commands for the Linear CLI.
package team

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
)

// cli.ClientFactory is a function that creates a Linear client.

// NewTeamCommand creates the team command group.
func NewTeamCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage Linear teams",
		Long:  "Commands for listing and viewing Linear teams.",
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))
	cmd.AddCommand(NewMembersCommand(clientFactory))
	cmd.AddCommand(NewCreateCommand(clientFactory))
	cmd.AddCommand(NewUpdateCommand(clientFactory))
	cmd.AddCommand(NewDeleteCommand(clientFactory))
	cmd.AddCommand(NewVelocityCommand(clientFactory))

	return cmd
}
