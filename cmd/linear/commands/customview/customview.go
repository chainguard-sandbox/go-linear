// Package customview provides custom view commands for the Linear CLI.
package customview

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
)

// NewCustomViewCommand creates the custom-view command group.
func NewCustomViewCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "custom-view",
		Aliases: []string{"cv"},
		Short:   "Manage Linear custom views",
		Long:    "Commands for listing, viewing, creating, updating, and deleting Linear custom views (saved filtered views).",
	}

	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))
	cmd.AddCommand(NewCreateCommand(clientFactory))
	cmd.AddCommand(NewUpdateCommand(clientFactory))
	cmd.AddCommand(NewDeleteCommand(clientFactory))

	return cmd
}
