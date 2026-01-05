package notification

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUnsubscribeCommand creates the notification unsubscribe command.
func NewUnsubscribeCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsubscribe <subscription-id>",
		Short: "Remove a notification subscription",
		Long: `Unsubscribe from updates. Safe operation.

Example: go-linear notification unsubscribe <subscription-uuid>

Related: notification_subscribe, notification_update`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUnsubscribe(cmd, client, args[0])
		},
	}

	return cmd
}

func runUnsubscribe(cmd *cobra.Command, client *linear.Client, subscriptionID string) error {
	ctx := cmd.Context()

	err := client.NotificationSubscriptionDelete(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Unsubscribed\n")
	return nil
}
