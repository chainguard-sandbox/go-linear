// Package initiative provides initiative commands for the Linear CLI.
package initiative

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
)

func NewInitiativeCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initiative",
		Short: "Manage Linear initiatives",
		Long:  "Commands for listing and viewing Linear initiatives.",
	}

	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))
	cmd.AddCommand(NewCreateCommand(clientFactory))
	cmd.AddCommand(NewUpdateCommand(clientFactory))

	return cmd
}
