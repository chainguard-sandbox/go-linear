// Package notification provides notification management commands for the Linear CLI.
package notification

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
)

// cli.ClientFactory is a function that creates a Linear client.

// NewNotificationCommand creates the notification command group.
func NewNotificationCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notification",
		Short: "Manage Linear notifications",
		Long:  "Commands for managing notifications and notification subscriptions.",
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))
	cmd.AddCommand(NewUpdateCommand(clientFactory))
	cmd.AddCommand(NewArchiveCommand(clientFactory))
	cmd.AddCommand(NewUnarchiveCommand(clientFactory))
	cmd.AddCommand(NewSubscribeCommand(clientFactory))
	cmd.AddCommand(NewUnsubscribeCommand(clientFactory))
	cmd.AddCommand(NewArchiveAllCommand(clientFactory))
	cmd.AddCommand(NewMarkReadAllCommand(clientFactory))
	cmd.AddCommand(NewMarkUnreadAllCommand(clientFactory))
	cmd.AddCommand(NewSnoozeAllCommand(clientFactory))
	cmd.AddCommand(NewUnsnoozeAllCommand(clientFactory))

	return cmd
}
