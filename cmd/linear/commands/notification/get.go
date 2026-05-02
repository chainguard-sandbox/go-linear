package notification

import (
	"fmt"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the notification get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <notification-id>",
		Short: "Get a single notification by ID",
		Long: `Get notification by UUID. Returns type, dates, read/snoozed status.

Example: go-linear notification get <uuid>

Related: notification_list, notification_archive`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0])
		},
	}

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, notificationID string) error {
	ctx := cmd.Context()

	notification, err := client.Notification(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), notification, true)
}
