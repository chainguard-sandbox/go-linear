// Package state provides workflow state commands for the Linear CLI.
package state

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// ClientFactory is a function that creates a Linear client.
type ClientFactory func() (*linear.Client, error)

// NewStateCommand creates the state command group.
func NewStateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Manage Linear workflow states",
		Long:  "Commands for listing Linear workflow states (Todo, In Progress, Done, etc.).",
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))

	return cmd
}
