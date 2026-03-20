package notification

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewUnarchiveCommand creates the notification unarchive command.
func NewUnarchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unarchive <notification-id>",
		Short: "Unarchive a notification",
		Long: `Restore archived notification to inbox. Safe operation.

Example: go-linear notification unarchive <uuid>

Related: notification_archive, notification_list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUnarchive(cmd, client, args[0])
		},
	}

	return cmd
}

func runUnarchive(cmd *cobra.Command, client *linear.Client, notificationID string) error {
	ctx := cmd.Context()

	err := client.NotificationUnarchive(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to unarchive notification: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
}
