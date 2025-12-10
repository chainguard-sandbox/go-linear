package notification

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUnsubscribeCommand creates the notification unsubscribe command.
func NewUnsubscribeCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsubscribe <subscription-id>",
		Short: "Remove a notification subscription",
		Long: `Remove a notification subscription to stop receiving updates.

This operation is safe and can be reversed by subscribing again.
Stop notifications for issues, projects, or cycles you no longer need to track.

Parameters:
  <subscription-id>: NotificationSubscription UUID to delete (required)

Examples:
  # Unsubscribe from updates
  linear notification unsubscribe <subscription-uuid>

TIP: Find subscription IDs in Linear notification settings or via API

Related Commands:
  - linear notification subscribe - Subscribe to updates
  - linear notification update - Manage notification status
  - linear notification archive - Archive notifications`,
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
	ctx := context.Background()

	err := client.NotificationSubscriptionDelete(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Unsubscribed\n")
	return nil
}
