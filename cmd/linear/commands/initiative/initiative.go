// Package initiative provides initiative commands for the Linear CLI.
package initiative

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
)

func NewInitiativeCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initiative",
		Short: "Manage Linear initiatives",
		Long:  "Commands for listing and viewing Linear initiatives.",
	}

	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewListSubCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))
	cmd.AddCommand(NewCreateCommand(clientFactory))
	cmd.AddCommand(NewUpdateCommand(clientFactory))
	cmd.AddCommand(NewDeleteCommand(clientFactory))
	cmd.AddCommand(NewArchiveCommand(clientFactory))
	cmd.AddCommand(NewUnarchiveCommand(clientFactory))
	cmd.AddCommand(NewStatusUpdateCreateCommand(clientFactory))
	cmd.AddCommand(NewStatusUpdateListCommand(clientFactory))
	cmd.AddCommand(NewStatusUpdateGetCommand(clientFactory))
	cmd.AddCommand(NewStatusUpdateArchiveCommand(clientFactory))
	cmd.AddCommand(NewAddProjectCommand(clientFactory))
	cmd.AddCommand(NewRemoveProjectCommand(clientFactory))

	return cmd
}
