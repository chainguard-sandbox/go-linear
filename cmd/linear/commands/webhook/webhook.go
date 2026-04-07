// Package webhook provides webhook management commands for the Linear CLI.
package webhook

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
)

// NewWebhookCommand creates the webhook command group.
func NewWebhookCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Manage Linear webhooks",
		Long:  "Commands for creating, listing, updating, and deleting webhooks.",
	}

	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewGetCommand(clientFactory))
	cmd.AddCommand(NewCreateCommand(clientFactory))
	cmd.AddCommand(NewUpdateCommand(clientFactory))
	cmd.AddCommand(NewDeleteCommand(clientFactory))
	cmd.AddCommand(NewRotateSecretCommand(clientFactory))

	return cmd
}
