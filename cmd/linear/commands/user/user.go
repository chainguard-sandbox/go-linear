// Package user provides user-related commands for the Linear CLI.
package user

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// ClientFactory is a function that creates a Linear client.
type ClientFactory func() (*linear.Client, error)

// NewUserCommand creates the user command group.
func NewUserCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage Linear users",
		Long:  "Commands for listing users and viewing their completed work.",
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))
	cmd.AddCommand(NewCompletedCommand(clientFactory))

	return cmd
}
