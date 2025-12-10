// Package notification provides notification management commands for the Linear CLI.
package notification

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// ClientFactory is a function that creates a Linear client.
type ClientFactory func() (*linear.Client, error)

// NewNotificationCommand creates the notification command group.
func NewNotificationCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notification",
		Short: "Manage Linear notifications",
		Long:  "Commands for managing notifications and notification subscriptions.",
	}

	// Add subcommands
	cmd.AddCommand(NewUpdateCommand(clientFactory))
	cmd.AddCommand(NewArchiveCommand(clientFactory))
	cmd.AddCommand(NewSubscribeCommand(clientFactory))
	cmd.AddCommand(NewUnsubscribeCommand(clientFactory))

	return cmd
}
