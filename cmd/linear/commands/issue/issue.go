// Package issue provides issue-related commands for the Linear CLI.
package issue

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// ClientFactory is a function that creates a Linear client.
type ClientFactory func() (*linear.Client, error)

// NewIssueCommand creates the issue command group.
func NewIssueCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Manage Linear issues",
		Long:  "Commands for listing, creating, updating, and deleting Linear issues.",
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))
	cmd.AddCommand(NewSearchCommand(clientFactory))
	cmd.AddCommand(NewCreateCommand(clientFactory))
	cmd.AddCommand(NewUpdateCommand(clientFactory))
	cmd.AddCommand(NewDeleteCommand(clientFactory))
	cmd.AddCommand(NewRelateCommand(clientFactory))
	cmd.AddCommand(NewUpdateRelationCommand(clientFactory))
	cmd.AddCommand(NewUnrelateCommand(clientFactory))
	cmd.AddCommand(NewAddLabelCommand(clientFactory))
	cmd.AddCommand(NewRemoveLabelCommand(clientFactory))

	return cmd
}
