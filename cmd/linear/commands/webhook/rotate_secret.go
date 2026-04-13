package webhook

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
)

// NewRotateSecretCommand creates the webhook rotate-secret command.
func NewRotateSecretCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate-secret <webhook-id>",
		Short: "Rotate the signing secret for a webhook",
		Long: `Rotate the signing secret for a webhook. Returns the new secret.

The old secret will no longer be valid for verifying payloads.

Example: go-linear webhook rotate-secret <uuid>

Related: webhook_get, webhook_update`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			ctx := cmd.Context()

			secret, err := client.WebhookRotateSecret(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to rotate webhook secret: %w", err)
			}

			// Print secret to stderr to avoid leaking it in captured stdout/logs
			fmt.Fprintf(cmd.ErrOrStderr(), "New signing secret: %s\n", secret)

			result := map[string]string{ // #nosec G101 -- not a credential, value is redacted
				"webhookId": args[0],
				"secret":    "[redacted - see stderr]",
			}
			return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
		},
	}

	return cmd
}
