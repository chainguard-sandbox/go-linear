// Package project provides project-related commands for the Linear CLI.
package project

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
)

// cli.ClientFactory is a function that creates a Linear client.

// NewProjectCommand creates the project command group.
func NewProjectCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage Linear projects",
		Long:  "Commands for listing, creating, and managing Linear projects.",
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))
	cmd.AddCommand(NewCreateCommand(clientFactory))
	cmd.AddCommand(NewUpdateCommand(clientFactory))
	cmd.AddCommand(NewDeleteCommand(clientFactory))
	cmd.AddCommand(NewMilestoneCreateCommand(clientFactory))
	cmd.AddCommand(NewMilestoneUpdateCommand(clientFactory))
	cmd.AddCommand(NewMilestoneDeleteCommand(clientFactory))
	cmd.AddCommand(NewStatusUpdateCreateCommand(clientFactory))
	cmd.AddCommand(NewStatusUpdateListCommand(clientFactory))
	cmd.AddCommand(NewStatusUpdateGetCommand(clientFactory))
	cmd.AddCommand(NewStatusUpdateDeleteCommand(clientFactory))

	return cmd
}
