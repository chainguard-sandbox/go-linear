package notification

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUnarchiveCommand creates the notification unarchive command.
func NewUnarchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unarchive <notification-id>",
		Short: "Unarchive a notification",
		Long: `Restore archived notification to inbox. Safe operation.

Example: go-linear notification unarchive <uuid> --output=json

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

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runUnarchive(cmd *cobra.Command, client *linear.Client, notificationID string) error {
	ctx := cmd.Context()

	err := client.NotificationUnarchive(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to unarchive notification: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), map[string]bool{"success": true}, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Notification unarchived\n")
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
